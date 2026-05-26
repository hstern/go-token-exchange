// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestResponseFieldConstants(t *testing.T) {
	t.Parallel()

	cases := []struct{ got, want string }{
		{FieldAccessToken, "access_token"},
		{FieldIssuedTokenType, "issued_token_type"},
		{FieldTokenType, "token_type"},
		{FieldExpiresIn, "expires_in"},
		{FieldScope, "scope"},
		{FieldRefreshToken, "refresh_token"},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("got %q, want %q", tc.got, tc.want)
		}
	}
}

// TestResponseMarshalSpecExample reproduces the RFC 8693 §2.2
// impersonation example response byte-stably (modulo map iteration
// order — assertions use strings.Contains).
func TestResponseMarshalSpecExample(t *testing.T) {
	t.Parallel()

	r := TokenExchangeResponse{
		AccessToken:     "eyJhbGciOiJFUzI1NiIsImtpZCI6IjllciJ9...",
		IssuedTokenType: TokenTypeAccessToken,
		TokenType:       "Bearer",
		ExpiresIn:       60,
	}
	out, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got := string(out)
	for _, want := range []string{
		`"access_token":"eyJhbGciOiJFUzI1NiIsImtpZCI6IjllciJ9..."`,
		`"issued_token_type":"urn:ietf:params:oauth:token-type:access_token"`,
		`"token_type":"Bearer"`,
		`"expires_in":60`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Marshal output missing %s; got %s", want, got)
		}
	}
	for _, mustAbsent := range []string{`"scope"`, `"refresh_token"`} {
		if strings.Contains(got, mustAbsent) {
			t.Errorf("Marshal unexpectedly contains %s; got %s", mustAbsent, got)
		}
	}
}

func TestResponseRoundTripExtra(t *testing.T) {
	t.Parallel()

	const in = `{
		"access_token": "tok",
		"issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
		"token_type": "Bearer",
		"expires_in": 300,
		"x-extension-bool": true,
		"x-extension-num": 42,
		"x-extension-obj": {"nested": "value"},
		"x-extension-arr": ["a", "b"]
	}`

	var r TokenExchangeResponse
	if err := json.Unmarshal([]byte(in), &r); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if r.AccessToken != "tok" || r.TokenType != "Bearer" || r.ExpiresIn != 300 {
		t.Errorf("known fields wrong: %+v", r)
	}
	if r.IssuedTokenType != TokenTypeAccessToken {
		t.Errorf("IssuedTokenType = %q", r.IssuedTokenType)
	}
	wantExtra := map[string]string{
		"x-extension-bool": "true",
		"x-extension-num":  "42",
		"x-extension-obj":  `{"nested":"value"}`,
		"x-extension-arr":  `["a","b"]`,
	}
	if len(r.Extra) != len(wantExtra) {
		t.Errorf("Extra has %d entries, want %d: %v", len(r.Extra), len(wantExtra), r.Extra)
	}
	for k, want := range wantExtra {
		got, ok := r.Extra[k]
		if !ok {
			t.Errorf("Extra missing %q", k)
			continue
		}
		// Canonicalize Extra raw bytes via re-encode so whitespace doesn't
		// foil the equality.
		var anyV any
		if err := json.Unmarshal(got, &anyV); err != nil {
			t.Errorf("Extra[%q] raw not valid JSON: %v", k, err)
			continue
		}
		reBytes, err := json.Marshal(anyV)
		if err != nil {
			t.Errorf("Extra[%q] re-marshal failed: %v", k, err)
			continue
		}
		if string(reBytes) != want {
			t.Errorf("Extra[%q] = %s, want %s", k, reBytes, want)
		}
	}

	// Re-marshal and re-unmarshal; the second round must equal the first.
	out, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal round 1: %v", err)
	}
	var second TokenExchangeResponse
	if err := json.Unmarshal(out, &second); err != nil {
		t.Fatalf("Unmarshal round 2: %v", err)
	}
	if second.AccessToken != r.AccessToken ||
		second.IssuedTokenType != r.IssuedTokenType ||
		second.TokenType != r.TokenType ||
		second.ExpiresIn != r.ExpiresIn {
		t.Errorf("round-trip known fields drifted: round1=%+v round2=%+v", r, second)
	}
	if len(second.Extra) != len(r.Extra) {
		t.Errorf("round-trip Extra count: round1=%d round2=%d", len(r.Extra), len(second.Extra))
	}
}

// TestResponseMarshalExtraCannotShadowBuiltins verifies the "built-ins
// win" rule: an Extra entry whose key matches a reserved JSON member
// MUST be dropped, even when the built-in is omitted by omitempty.
func TestResponseMarshalExtraCannotShadowBuiltins(t *testing.T) {
	t.Parallel()

	r := TokenExchangeResponse{
		AccessToken:     "real-token",
		IssuedTokenType: TokenTypeAccessToken,
		TokenType:       "Bearer",
		Extra: map[string]json.RawMessage{
			"access_token":  json.RawMessage(`"forged-token"`),
			"scope":         json.RawMessage(`"forged-scope"`),
			"refresh_token": json.RawMessage(`"forged-refresh"`),
			"x-allowed":     json.RawMessage(`"kept"`),
		},
	}
	out, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got := string(out)

	if !strings.Contains(got, `"access_token":"real-token"`) {
		t.Errorf("real access_token lost; got %s", got)
	}
	if strings.Contains(got, "forged-token") {
		t.Errorf("Extra access_token leaked into output; got %s", got)
	}
	if strings.Contains(got, "forged-scope") {
		t.Errorf("Extra scope leaked into output (omitempty backfill); got %s", got)
	}
	if strings.Contains(got, "forged-refresh") {
		t.Errorf("Extra refresh_token leaked into output (omitempty backfill); got %s", got)
	}
	if !strings.Contains(got, `"x-allowed":"kept"`) {
		t.Errorf("non-builtin Extra entry dropped; got %s", got)
	}
}

func TestResponseUnmarshalNullAndEmpty(t *testing.T) {
	t.Parallel()

	t.Run("null", func(t *testing.T) {
		t.Parallel()
		r := TokenExchangeResponse{AccessToken: "stale"}
		if err := json.Unmarshal([]byte(`null`), &r); err != nil {
			t.Fatalf("Unmarshal null: %v", err)
		}
		// The default codec leaves r unchanged for null on a non-pointer
		// alias unmarshal — but we use a pointer receiver and explicit
		// *r = ... so the second alias.Unmarshal of null is a no-op.
		// The captured-Extra pass over a nil map leaves Extra nil.
		if r.Extra != nil {
			t.Errorf("Extra after null = %v, want nil", r.Extra)
		}
	})

	t.Run("empty object", func(t *testing.T) {
		t.Parallel()
		var r TokenExchangeResponse
		if err := json.Unmarshal([]byte(`{}`), &r); err != nil {
			t.Fatalf("Unmarshal empty: %v", err)
		}
		if r.AccessToken != "" || r.Extra != nil {
			t.Errorf("empty object decoded to %+v", r)
		}
	})

	t.Run("non-object errors", func(t *testing.T) {
		t.Parallel()
		var r TokenExchangeResponse
		err := json.Unmarshal([]byte(`"oops"`), &r)
		if err == nil {
			t.Errorf("expected error decoding string into response, got nil")
		}
	})
}

func TestResponseMarshalNoExtraUsesDefaultPath(t *testing.T) {
	t.Parallel()

	r := TokenExchangeResponse{
		AccessToken:     "tok",
		IssuedTokenType: TokenTypeAccessToken,
		TokenType:       "Bearer",
	}
	out, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// No Extra → declaration-order output (no re-decode through a map).
	want := `{"access_token":"tok","issued_token_type":"urn:ietf:params:oauth:token-type:access_token","token_type":"Bearer"}`
	if string(out) != want {
		t.Errorf("Marshal = %s, want %s", out, want)
	}
}
