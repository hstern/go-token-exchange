// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteTokenExchangeResponseSuccess(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	resp := &TokenExchangeResponse{
		AccessToken:     "issued-token",
		IssuedTokenType: TokenTypeAccessToken,
		TokenType:       "Bearer",
		ExpiresIn:       300,
	}
	if err := WriteTokenExchangeResponse(rr, resp); err != nil {
		t.Fatalf("WriteTokenExchangeResponse: %v", err)
	}

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q", got)
	}
	if got := rr.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q", got)
	}
	if got := rr.Header().Get("Pragma"); got != "no-cache" {
		t.Errorf("Pragma = %q", got)
	}

	var back TokenExchangeResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &back); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if back.AccessToken != "issued-token" || back.ExpiresIn != 300 {
		t.Errorf("body did not round-trip: %+v", back)
	}
}

func TestWriteTokenExchangeResponseValidates(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	resp := &TokenExchangeResponse{
		// AccessToken missing
		IssuedTokenType: TokenTypeAccessToken,
		TokenType:       "Bearer",
	}
	err := WriteTokenExchangeResponse(rr, resp)
	if err == nil {
		t.Fatalf("WriteTokenExchangeResponse returned nil error on invalid resp")
	}
	if !errors.Is(err, &ValidationError{Parameter: FieldAccessToken, Reason: "missing"}) {
		t.Errorf("err = %v, want validation failure on access_token", err)
	}
	// httptest.ResponseRecorder reports 200 as the default until WriteHeader
	// is called; we don't assert on status because nothing was written.
	if rr.Body.Len() != 0 {
		t.Errorf("body written despite validation failure: %q", rr.Body.String())
	}
}

func TestWriteTokenExchangeResponseRejectsNilWriter(t *testing.T) {
	t.Parallel()

	resp := &TokenExchangeResponse{
		AccessToken:     "tok",
		IssuedTokenType: TokenTypeAccessToken,
		TokenType:       "Bearer",
	}
	if err := WriteTokenExchangeResponse(nil, resp); err == nil {
		t.Errorf("WriteTokenExchangeResponse(nil, ...) returned no error")
	}
}

func TestWriteTokenExchangeErrorDefault400(t *testing.T) {
	t.Parallel()

	codes := []string{
		ErrCodeInvalidRequest,
		ErrCodeInvalidGrant,
		ErrCodeUnauthorizedClient,
		ErrCodeUnsupportedGrantType,
		ErrCodeInvalidScope,
		ErrCodeInvalidTarget,
	}
	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			t.Parallel()
			rr := httptest.NewRecorder()
			e := &TokenExchangeError{
				Code:        code,
				Description: "test",
			}
			if err := WriteTokenExchangeError(rr, e); err != nil {
				t.Fatalf("WriteTokenExchangeError: %v", err)
			}
			if rr.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rr.Code)
			}
			if got := rr.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q", got)
			}
			if got := rr.Header().Get("Cache-Control"); got != "no-store" {
				t.Errorf("Cache-Control = %q", got)
			}
		})
	}
}

func TestWriteTokenExchangeErrorInvalidClient401(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	e := &TokenExchangeError{Code: ErrCodeInvalidClient}
	if err := WriteTokenExchangeError(rr, e); err != nil {
		t.Fatalf("WriteTokenExchangeError: %v", err)
	}
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
	got := strings.TrimSpace(rr.Body.String())
	if got != `{"error":"invalid_client"}` {
		t.Errorf("body = %q, want minimal invalid_client JSON", got)
	}
}

func TestWriteTokenExchangeErrorUnknownCodeDefaults400(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	e := &TokenExchangeError{Code: "downstream:custom_error"}
	if err := WriteTokenExchangeError(rr, e); err != nil {
		t.Fatalf("WriteTokenExchangeError: %v", err)
	}
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for unknown code", rr.Code)
	}
}

func TestWriteTokenExchangeErrorRejectsNilOrEmpty(t *testing.T) {
	t.Parallel()

	t.Run("nil writer", func(t *testing.T) {
		t.Parallel()
		err := WriteTokenExchangeError(nil, &TokenExchangeError{Code: ErrCodeInvalidRequest})
		if err == nil {
			t.Errorf("expected error on nil writer")
		}
	})

	t.Run("nil error", func(t *testing.T) {
		t.Parallel()
		err := WriteTokenExchangeError(httptest.NewRecorder(), nil)
		if err == nil {
			t.Errorf("expected error on nil *TokenExchangeError")
		}
	})

	t.Run("empty code", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		err := WriteTokenExchangeError(rr, &TokenExchangeError{Description: "x"})
		if err == nil {
			t.Fatalf("expected error on empty Code")
		}
		if rr.Body.Len() != 0 {
			t.Errorf("body written despite empty code: %q", rr.Body.String())
		}
	})
}

func TestWriteTokenExchangeErrorDoesNotLeakCause(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	e := (&TokenExchangeError{Code: ErrCodeInvalidTarget}).
		WithCause(errors.New("internal: kube secret 42 unreachable"))
	if err := WriteTokenExchangeError(rr, e); err != nil {
		t.Fatalf("WriteTokenExchangeError: %v", err)
	}
	if strings.Contains(rr.Body.String(), "kube secret 42") {
		t.Errorf("cause text leaked into wire body: %s", rr.Body.String())
	}
}

func TestHTTPStatusForErrorCode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		code string
		want int
	}{
		{ErrCodeInvalidRequest, http.StatusBadRequest},
		{ErrCodeInvalidClient, http.StatusUnauthorized},
		{ErrCodeInvalidGrant, http.StatusBadRequest},
		{ErrCodeUnauthorizedClient, http.StatusBadRequest},
		{ErrCodeUnsupportedGrantType, http.StatusBadRequest},
		{ErrCodeInvalidScope, http.StatusBadRequest},
		{ErrCodeInvalidTarget, http.StatusBadRequest},
		{"unknown", http.StatusBadRequest},
		{"", http.StatusBadRequest},
	}
	for _, tc := range cases {
		if got := httpStatusForErrorCode(tc.code); got != tc.want {
			t.Errorf("httpStatusForErrorCode(%q) = %d, want %d", tc.code, got, tc.want)
		}
	}
}
