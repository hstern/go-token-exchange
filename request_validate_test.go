// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"errors"
	"testing"
)

// minimalValidRequest returns a TokenExchangeRequest that passes
// Validate; tests mutate one field at a time to exercise each rule
// in isolation.
func minimalValidRequest() *TokenExchangeRequest {
	return &TokenExchangeRequest{
		GrantType:        GrantTypeTokenExchange,
		SubjectToken:     "subject-token-value",
		SubjectTokenType: TokenTypeAccessToken,
	}
}

func TestValidateAcceptsMinimal(t *testing.T) {
	t.Parallel()

	if err := minimalValidRequest().Validate(); err != nil {
		t.Errorf("Validate on minimal valid request: %v", err)
	}
}

func TestValidateAcceptsFullyPopulated(t *testing.T) {
	t.Parallel()

	r := &TokenExchangeRequest{
		GrantType:          GrantTypeTokenExchange,
		Resource:           []string{"https://backend.example.com/api"},
		Audience:           []string{"urn:example:cooperation-context"},
		Scope:              []string{"read", "write"},
		RequestedTokenType: TokenTypeJWT,
		SubjectToken:       "subject-jwt",
		SubjectTokenType:   TokenTypeIDToken,
		ActorToken:         "actor-jwt",
		ActorTokenType:     TokenTypeAccessToken,
	}
	if err := r.Validate(); err != nil {
		t.Errorf("Validate on fully populated request: %v", err)
	}
}

func TestValidateNilReceiver(t *testing.T) {
	t.Parallel()

	var r *TokenExchangeRequest
	err := r.Validate()
	if err == nil {
		t.Fatalf("Validate(nil) returned nil error")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("Validate(nil) = %v, want *ValidationError", err)
	}
	if ve.Rule != ruleRequest {
		t.Errorf("Rule = %q, want %q", ve.Rule, ruleRequest)
	}
}

func TestValidateGrantTypeWrong(t *testing.T) {
	t.Parallel()

	cases := []string{
		"",
		"authorization_code",
		"client_credentials",
		"urn:ietf:params:oauth:grant-type:device_code",
	}
	for _, gt := range cases {
		t.Run(gt, func(t *testing.T) {
			t.Parallel()
			r := minimalValidRequest()
			r.GrantType = gt
			err := r.Validate()
			if err == nil {
				t.Fatalf("Validate(grant_type=%q) returned nil", gt)
			}
			if !errors.Is(err, ErrInvalidGrantType) {
				t.Errorf("errors.Is(err, ErrInvalidGrantType) = false")
			}
			if !errors.Is(err, &ValidationError{Parameter: ParamGrantType}) {
				t.Errorf("errors.Is(err, ValidationError{Parameter: grant_type}) = false")
			}
		})
	}
}

func TestValidateSubjectTokenMissing(t *testing.T) {
	t.Parallel()

	r := minimalValidRequest()
	r.SubjectToken = ""
	err := r.Validate()
	if !errors.Is(err, &ValidationError{Parameter: ParamSubjectToken, Reason: "missing"}) {
		t.Errorf("Validate(empty subject_token) = %v", err)
	}
}

func TestValidateSubjectTokenTypeMissing(t *testing.T) {
	t.Parallel()

	r := minimalValidRequest()
	r.SubjectTokenType = ""
	err := r.Validate()
	if !errors.Is(err, &ValidationError{Parameter: ParamSubjectTokenType, Reason: "missing"}) {
		t.Errorf("Validate(empty subject_token_type) = %v", err)
	}
}

func TestValidateActorPairing(t *testing.T) {
	t.Parallel()

	t.Run("token without type", func(t *testing.T) {
		t.Parallel()
		r := minimalValidRequest()
		r.ActorToken = "actor"
		err := r.Validate()
		if !errors.Is(err, &ValidationError{Parameter: ParamActorTokenType}) {
			t.Errorf("Validate(actor_token without type) = %v", err)
		}
	})

	t.Run("type without token", func(t *testing.T) {
		t.Parallel()
		r := minimalValidRequest()
		r.ActorTokenType = TokenTypeAccessToken
		err := r.Validate()
		if !errors.Is(err, &ValidationError{Parameter: ParamActorToken}) {
			t.Errorf("Validate(actor_token_type without token) = %v", err)
		}
	})

	t.Run("both present", func(t *testing.T) {
		t.Parallel()
		r := minimalValidRequest()
		r.ActorToken = "actor"
		r.ActorTokenType = TokenTypeAccessToken
		if err := r.Validate(); err != nil {
			t.Errorf("Validate(both actor fields) = %v", err)
		}
	})

	t.Run("both absent", func(t *testing.T) {
		t.Parallel()
		r := minimalValidRequest()
		// already both absent
		if err := r.Validate(); err != nil {
			t.Errorf("Validate(neither actor field) = %v", err)
		}
	})
}

func TestValidateURIShapeOnTokenTypes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		mutate      func(*TokenExchangeRequest)
		wantParam   string
		wantInvalid bool
	}{
		{
			name:        "subject_token_type not a URI",
			mutate:      func(r *TokenExchangeRequest) { r.SubjectTokenType = "not-a-uri" },
			wantParam:   ParamSubjectTokenType,
			wantInvalid: true,
		},
		{
			name:        "requested_token_type not a URI",
			mutate:      func(r *TokenExchangeRequest) { r.RequestedTokenType = "not-a-uri" },
			wantParam:   ParamRequestedTokenType,
			wantInvalid: true,
		},
		{
			name: "actor_token_type not a URI",
			mutate: func(r *TokenExchangeRequest) {
				r.ActorToken = "actor"
				r.ActorTokenType = "not-a-uri"
			},
			wantParam:   ParamActorTokenType,
			wantInvalid: true,
		},
		{
			name:        "requested_token_type empty is fine",
			mutate:      func(r *TokenExchangeRequest) { r.RequestedTokenType = "" },
			wantInvalid: false,
		},
		{
			name: "non-builtin custom URI accepted",
			mutate: func(r *TokenExchangeRequest) {
				r.RequestedTokenType = "urn:example:custom-token-type"
			},
			wantInvalid: false,
		},
		{
			name: "https URI accepted",
			mutate: func(r *TokenExchangeRequest) {
				r.RequestedTokenType = "https://example.com/token-type/custom"
			},
			wantInvalid: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := minimalValidRequest()
			tc.mutate(r)
			err := r.Validate()
			if tc.wantInvalid {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tc.wantParam != "" {
					if !errors.Is(err, &ValidationError{Parameter: tc.wantParam, Reason: "not a valid URI"}) {
						t.Errorf("err = %v, want URI shape failure on %s", err, tc.wantParam)
					}
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestIsValidTokenTypeURI(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want bool
	}{
		{"", false},
		{"not-a-uri", false},
		{"/path/only", false},
		{"#fragment", false},
		{"urn:ietf:params:oauth:token-type:access_token", true},
		{"urn:example:custom", true},
		{"https://example.com/token-type", true},
		{"http://example.com", true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := isValidTokenTypeURI(tc.in); got != tc.want {
				t.Errorf("isValidTokenTypeURI(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
