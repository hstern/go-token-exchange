// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"errors"
	"testing"
)

func minimalValidResponse() *TokenExchangeResponse {
	return &TokenExchangeResponse{
		AccessToken:     "issued-token",
		IssuedTokenType: TokenTypeAccessToken,
		TokenType:       "Bearer",
	}
}

func TestResponseValidateAcceptsMinimal(t *testing.T) {
	t.Parallel()

	if err := minimalValidResponse().Validate(); err != nil {
		t.Errorf("Validate on minimal valid response: %v", err)
	}
}

func TestResponseValidateAcceptsFullyPopulated(t *testing.T) {
	t.Parallel()

	r := &TokenExchangeResponse{
		AccessToken:     "issued-jwt",
		IssuedTokenType: TokenTypeJWT,
		TokenType:       "Bearer",
		ExpiresIn:       3600,
		Scope:           "read write",
		RefreshToken:    "refresh-token",
	}
	if err := r.Validate(); err != nil {
		t.Errorf("Validate on fully populated response: %v", err)
	}
}

func TestResponseValidateNilReceiver(t *testing.T) {
	t.Parallel()

	var r *TokenExchangeResponse
	err := r.Validate()
	if err == nil {
		t.Fatalf("Validate(nil) returned nil error")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("Validate(nil) = %v, want *ValidationError", err)
	}
	if ve.Rule != ruleResponse {
		t.Errorf("Rule = %q, want %q", ve.Rule, ruleResponse)
	}
}

func TestResponseValidateMissingFields(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		mutate    func(*TokenExchangeResponse)
		wantParam string
	}{
		{
			name:      "access_token missing",
			mutate:    func(r *TokenExchangeResponse) { r.AccessToken = "" },
			wantParam: FieldAccessToken,
		},
		{
			name:      "issued_token_type missing",
			mutate:    func(r *TokenExchangeResponse) { r.IssuedTokenType = "" },
			wantParam: FieldIssuedTokenType,
		},
		{
			name:      "token_type missing",
			mutate:    func(r *TokenExchangeResponse) { r.TokenType = "" },
			wantParam: FieldTokenType,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := minimalValidResponse()
			tc.mutate(r)
			err := r.Validate()
			if err == nil {
				t.Fatalf("expected error")
			}
			if !errors.Is(err, &ValidationError{Parameter: tc.wantParam, Reason: "missing"}) {
				t.Errorf("err = %v, want missing on %s", err, tc.wantParam)
			}
		})
	}
}

func TestResponseValidateIssuedTokenTypeURIShape(t *testing.T) {
	t.Parallel()

	r := minimalValidResponse()
	r.IssuedTokenType = "not-a-uri"
	err := r.Validate()
	if !errors.Is(err, &ValidationError{Parameter: FieldIssuedTokenType, Reason: "not a valid URI"}) {
		t.Errorf("err = %v, want URI shape failure on issued_token_type", err)
	}
}

func TestResponseValidateAcceptsExtensionTokenTypeURI(t *testing.T) {
	t.Parallel()

	r := minimalValidResponse()
	r.IssuedTokenType = "urn:example:custom-token-type"
	if err := r.Validate(); err != nil {
		t.Errorf("Validate rejected unregistered but URI-shaped token type: %v", err)
	}
}
