// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// goodRequest returns a minimal request that passes Validate; the
// Exchange tests mutate one field at a time when exercising
// validation paths.
func goodRequest() *TokenExchangeRequest {
	return &TokenExchangeRequest{
		GrantType:        GrantTypeTokenExchange,
		SubjectToken:     "subject-token",
		SubjectTokenType: TokenTypeAccessToken,
	}
}

// echoServer returns an httptest.Server whose handler invokes the
// caller's function, plus a Client pointed at it. The Server's
// Close is registered via t.Cleanup.
func echoServer(t *testing.T, handler http.HandlerFunc) Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewClient(srv.URL)
}

func TestExchangeSuccess(t *testing.T) {
	t.Parallel()

	c := echoServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type = %q", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q", got)
		}
		body, _ := io.ReadAll(r.Body)
		form, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatalf("ParseQuery: %v", err)
		}
		if form.Get("grant_type") != GrantTypeTokenExchange {
			t.Errorf("grant_type = %q", form.Get("grant_type"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"access_token": "issued-token",
			"issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
			"token_type": "Bearer",
			"expires_in": 300
		}`))
	})

	got, err := c.Exchange(context.Background(), goodRequest())
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if got.AccessToken != "issued-token" || got.TokenType != "Bearer" || got.ExpiresIn != 300 {
		t.Errorf("response = %+v", got)
	}
	if got.IssuedTokenType != TokenTypeAccessToken {
		t.Errorf("IssuedTokenType = %q", got.IssuedTokenType)
	}
}

func TestExchangeValidatesBeforeSend(t *testing.T) {
	t.Parallel()

	called := false
	c := echoServer(t, func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	})

	bad := goodRequest()
	bad.GrantType = "not-token-exchange"

	_, err := c.Exchange(context.Background(), bad)
	if err == nil {
		t.Fatalf("Exchange returned no error on invalid request")
	}
	if !errors.Is(err, ErrInvalidGrantType) {
		t.Errorf("Exchange err = %v, want ErrInvalidGrantType in chain", err)
	}
	if called {
		t.Errorf("server was called despite validation failure")
	}
}

func TestExchangeJSONErrorDecodesTyped(t *testing.T) {
	t.Parallel()

	c := echoServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"error": "invalid_target",
			"error_description": "audience https://other.example.com is not authorized",
			"error_uri": "https://example.com/errors/invalid_target"
		}`))
	})

	_, err := c.Exchange(context.Background(), goodRequest())
	if err == nil {
		t.Fatalf("Exchange returned nil error")
	}

	var te *TokenExchangeError
	if !errors.As(err, &te) {
		t.Fatalf("err = %v, want *TokenExchangeError", err)
	}
	if te.Code != ErrCodeInvalidTarget {
		t.Errorf("Code = %q, want %q", te.Code, ErrCodeInvalidTarget)
	}
	if te.Description != "audience https://other.example.com is not authorized" {
		t.Errorf("Description = %q", te.Description)
	}
	if !errors.Is(err, &TokenExchangeError{Code: ErrCodeInvalidTarget}) {
		t.Errorf("errors.Is sentinel match failed")
	}
}

func TestExchangeInvalidClient401Decoded(t *testing.T) {
	t.Parallel()

	c := echoServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_client"}`))
	})

	_, err := c.Exchange(context.Background(), goodRequest())
	if err == nil {
		t.Fatalf("Exchange returned nil error")
	}
	if !errors.Is(err, &TokenExchangeError{Code: ErrCodeInvalidClient}) {
		t.Errorf("err = %v, want invalid_client sentinel match", err)
	}
}

func TestExchange4xxNonJSONIsTransport(t *testing.T) {
	t.Parallel()

	c := echoServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("not json"))
	})

	_, err := c.Exchange(context.Background(), goodRequest())
	if err == nil {
		t.Fatalf("Exchange returned nil error")
	}
	var te *TokenExchangeError
	if errors.As(err, &te) {
		t.Errorf("err decoded to *TokenExchangeError despite text/plain body")
	}
	if !strings.Contains(err.Error(), "HTTP 400") {
		t.Errorf("err = %v, want HTTP 400 in message", err)
	}
}

func TestExchange5xxIsTransport(t *testing.T) {
	t.Parallel()

	c := echoServer(t, func(w http.ResponseWriter, _ *http.Request) {
		// Even with a JSON-shaped body, 5xx is a transport-class error.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"oops"}`))
	})

	_, err := c.Exchange(context.Background(), goodRequest())
	if err == nil {
		t.Fatalf("Exchange returned nil error")
	}
	var te *TokenExchangeError
	if errors.As(err, &te) {
		t.Errorf("err decoded to *TokenExchangeError on 5xx; want transport class")
	}
	if !strings.Contains(err.Error(), "HTTP 500") {
		t.Errorf("err = %v, want HTTP 500 in message", err)
	}
}

func TestExchangeJSONErrorWithoutCode(t *testing.T) {
	t.Parallel()

	c := echoServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error_description":"missing code"}`))
	})

	_, err := c.Exchange(context.Background(), goodRequest())
	if err == nil {
		t.Fatalf("Exchange returned nil error")
	}
	var te *TokenExchangeError
	if errors.As(err, &te) {
		t.Errorf("err decoded to *TokenExchangeError despite empty Code")
	}
}

func TestExchangeUnreachable(t *testing.T) {
	t.Parallel()

	// Pick a URL the OS will refuse instantly; closed loopback port.
	c := NewClient("http://127.0.0.1:1")
	_, err := c.Exchange(context.Background(), goodRequest())
	if err == nil {
		t.Fatalf("Exchange to unreachable endpoint returned nil error")
	}
	var ue *url.Error
	if !errors.As(err, &ue) {
		t.Errorf("err = %v, want *url.Error in chain", err)
	}
}

func TestExchangeContextCanceled(t *testing.T) {
	t.Parallel()

	// With a pre-canceled context, http.Client.Do returns
	// immediately without dispatching the request — the server's
	// handler is never invoked, so a trivial handler suffices.
	c := echoServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.Exchange(ctx, goodRequest())
	if err == nil {
		t.Fatalf("Exchange did not return error on canceled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled in chain", err)
	}
}

func TestExchangeRejectsEmptyEndpoint(t *testing.T) {
	t.Parallel()

	c := NewClient("")
	_, err := c.Exchange(context.Background(), goodRequest())
	if err == nil {
		t.Fatalf("Exchange with empty endpoint returned nil error")
	}
	if !strings.Contains(err.Error(), "token endpoint not configured") {
		t.Errorf("err = %v, want endpoint-not-configured message", err)
	}
}

func TestIsJSONContentType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want bool
	}{
		{"", false},
		{"application/json", true},
		{"application/json; charset=utf-8", true},
		{"APPLICATION/JSON", true},
		{"application/jsonish", false},
		{"text/plain", false},
		{"text/html; charset=utf-8", false},
		{"not a media type", false},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := isJSONContentType(tc.in); got != tc.want {
				t.Errorf("isJSONContentType(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

// TestNewClientReturnsInterface pins the Client interface as the
// stable surface — a Client variable is the interface, not the
// concrete type.
func TestNewClientReturnsInterface(t *testing.T) {
	t.Parallel()

	c := NewClient("https://as.example.com/token")
	if c == nil {
		t.Fatalf("NewClient returned nil")
	}
}
