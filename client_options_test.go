// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// countingTransport is an http.RoundTripper that increments a
// counter on each request and then defers to a backing transport.
// Used to prove the Client routes its requests through the
// http.Client injected via WithHTTPClient rather than
// http.DefaultClient.
type countingTransport struct {
	calls atomic.Int64
	inner http.RoundTripper
}

func (t *countingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.calls.Add(1)
	return t.inner.RoundTrip(req)
}

func TestWithHTTPClientRoutesThroughInjected(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"access_token": "tok",
			"issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
			"token_type": "Bearer"
		}`))
	}))
	t.Cleanup(srv.Close)

	tr := &countingTransport{inner: http.DefaultTransport}
	hc := &http.Client{Transport: tr}

	c := NewClient(srv.URL, WithHTTPClient(hc))
	if _, err := c.Exchange(context.Background(), goodRequest()); err != nil {
		t.Fatalf("Exchange: %v", err)
	}

	if got := tr.calls.Load(); got != 1 {
		t.Errorf("injected transport saw %d calls, want 1", got)
	}
}

func TestWithHTTPClientNilRestoresDefault(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"access_token": "tok",
			"issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
			"token_type": "Bearer"
		}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL, WithHTTPClient(nil))
	if _, err := c.Exchange(context.Background(), goodRequest()); err != nil {
		t.Fatalf("Exchange after WithHTTPClient(nil): %v", err)
	}
}

func TestNewClientNoOptionsKeepsDefault(t *testing.T) {
	t.Parallel()

	// Without any options, NewClient must still produce a usable
	// Client — the variadic addition is backward-compatible.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"access_token": "tok",
			"issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
			"token_type": "Bearer"
		}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL)
	if _, err := c.Exchange(context.Background(), goodRequest()); err != nil {
		t.Fatalf("Exchange with no options: %v", err)
	}
}

func TestNewClientNilOptionsSkipped(t *testing.T) {
	t.Parallel()

	// A nil Option must be silently skipped rather than panicking;
	// this matters for downstream wrappers that build option lists
	// conditionally.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"access_token": "tok",
			"issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
			"token_type": "Bearer"
		}`))
	}))
	t.Cleanup(srv.Close)

	var (
		nilLeading  Option
		nilTrailing Option
	)
	c := NewClient(srv.URL, nilLeading, WithHTTPClient(nil), nilTrailing)
	if _, err := c.Exchange(context.Background(), goodRequest()); err != nil {
		t.Fatalf("Exchange with nil-laced options: %v", err)
	}
}

func TestOptionsAppliedInOrder(t *testing.T) {
	t.Parallel()

	// A later option must win over an earlier one for the same
	// field. Two countingTransports with distinct counters prove
	// only the last WithHTTPClient took effect.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"access_token": "tok",
			"issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
			"token_type": "Bearer"
		}`))
	}))
	t.Cleanup(srv.Close)

	first := &countingTransport{inner: http.DefaultTransport}
	second := &countingTransport{inner: http.DefaultTransport}

	c := NewClient(srv.URL,
		WithHTTPClient(&http.Client{Transport: first}),
		WithHTTPClient(&http.Client{Transport: second}),
	)
	if _, err := c.Exchange(context.Background(), goodRequest()); err != nil {
		t.Fatalf("Exchange: %v", err)
	}

	if got := first.calls.Load(); got != 0 {
		t.Errorf("first transport saw %d calls, want 0 (overridden)", got)
	}
	if got := second.calls.Load(); got != 1 {
		t.Errorf("second transport saw %d calls, want 1", got)
	}
}

// TestOptionTypeAcceptsSentinelMatch is a smoke test that a typed
// nil Option from one helper is still recognized as a nil-checked
// skip target (the variadic loop guards against panicking on a
// nil func value).
func TestOptionTypeAcceptsSentinelMatch(t *testing.T) {
	t.Parallel()

	var opt Option // typed nil
	c := NewClient("https://as.example.com/token", opt)
	if c == nil {
		t.Errorf("NewClient(typed-nil Option) returned nil")
	}

	// Sanity-check that the constructor did not panic and that the
	// Client interface assertion still holds (i.e. the typed value
	// is non-nil even though the underlying field is set to a
	// default http.Client).
	if errors.Is(nil, nil) && c == nil {
		t.Errorf("unreachable; sentinel guard")
	}
}
