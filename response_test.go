// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"encoding/json"
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
