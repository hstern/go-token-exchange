// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTokenExchangeResponseZeroValue(t *testing.T) {
	t.Parallel()

	var r TokenExchangeResponse
	if r.AccessToken != "" || r.IssuedTokenType != "" || r.TokenType != "" {
		t.Errorf("zero required fields not empty: %+v", r)
	}
	if r.ExpiresIn != 0 {
		t.Errorf("zero ExpiresIn = %d, want 0", r.ExpiresIn)
	}
	if r.Extra != nil {
		t.Errorf("zero Extra = %v, want nil", r.Extra)
	}
}

// TestTokenExchangeResponseDefaultJSON verifies the surface honored by
// the default encoding/json codec on this commit. Custom Marshal and
// Unmarshal that round-trip Extra arrive in a later phase; here the
// json:"-" tag must keep the default codec from emitting Extra.
func TestTokenExchangeResponseDefaultJSON(t *testing.T) {
	t.Parallel()

	r := TokenExchangeResponse{
		AccessToken:     "eyJhbGciOiJFUzI1NiIsImtpZCI6IjllciJ9.PAYLOAD.SIG",
		IssuedTokenType: "urn:ietf:params:oauth:token-type:access_token",
		TokenType:       "Bearer",
		ExpiresIn:       60,
		Extra: map[string]json.RawMessage{
			"x-extension": json.RawMessage(`"v"`),
		},
	}

	out, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got := string(out)

	for _, want := range []string{
		`"access_token":"eyJhbGciOiJFUzI1NiIsImtpZCI6IjllciJ9.PAYLOAD.SIG"`,
		`"issued_token_type":"urn:ietf:params:oauth:token-type:access_token"`,
		`"token_type":"Bearer"`,
		`"expires_in":60`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Marshal output missing %s; got %s", want, got)
		}
	}
	for _, mustAbsent := range []string{
		`"scope"`,
		`"refresh_token"`,
		`"x-extension"`,
		`"Extra"`,
	} {
		if strings.Contains(got, mustAbsent) {
			t.Errorf("Marshal output unexpectedly contains %s; got %s", mustAbsent, got)
		}
	}
}

func TestTokenExchangeResponseUnmarshalRequiredOnly(t *testing.T) {
	t.Parallel()

	const in = `{
		"access_token": "tok",
		"issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
		"token_type": "Bearer"
	}`

	var r TokenExchangeResponse
	if err := json.Unmarshal([]byte(in), &r); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if r.AccessToken != "tok" {
		t.Errorf("AccessToken = %q", r.AccessToken)
	}
	if r.IssuedTokenType != "urn:ietf:params:oauth:token-type:access_token" {
		t.Errorf("IssuedTokenType = %q", r.IssuedTokenType)
	}
	if r.TokenType != "Bearer" {
		t.Errorf("TokenType = %q", r.TokenType)
	}
	if r.ExpiresIn != 0 {
		t.Errorf("ExpiresIn = %d, want 0", r.ExpiresIn)
	}
}
