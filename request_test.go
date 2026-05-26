// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"net/url"
	"testing"
)

// Literal URI strings keep this PR independent of the token-type
// constants that land in a sibling PR.
const (
	uriGrantType   = "urn:ietf:params:oauth:grant-type:token-exchange"
	uriAccessToken = "urn:ietf:params:oauth:token-type:access_token"
	uriIDToken     = "urn:ietf:params:oauth:token-type:id_token"
	uriJWT         = "urn:ietf:params:oauth:token-type:jwt"
)

func TestTokenExchangeRequestZeroValue(t *testing.T) {
	t.Parallel()

	var r TokenExchangeRequest
	if r.GrantType != "" {
		t.Errorf("zero GrantType = %q, want empty", r.GrantType)
	}
	if r.Resource != nil {
		t.Errorf("zero Resource = %v, want nil", r.Resource)
	}
	if r.Audience != nil {
		t.Errorf("zero Audience = %v, want nil", r.Audience)
	}
	if r.Scope != nil {
		t.Errorf("zero Scope = %v, want nil", r.Scope)
	}
	if r.Extra != nil {
		t.Errorf("zero Extra = %v, want nil", r.Extra)
	}
}

func TestRequestedOrSubjectTokenType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		requested string
		subject   string
		want      string
	}{
		{
			name:      "requested set wins",
			requested: uriJWT,
			subject:   uriAccessToken,
			want:      uriJWT,
		},
		{
			name:      "requested empty falls back to subject",
			requested: "",
			subject:   uriAccessToken,
			want:      uriAccessToken,
		},
		{
			name:      "both empty returns empty",
			requested: "",
			subject:   "",
			want:      "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := &TokenExchangeRequest{
				RequestedTokenType: tc.requested,
				SubjectTokenType:   tc.subject,
			}
			if got := r.RequestedOrSubjectTokenType(); got != tc.want {
				t.Errorf("RequestedOrSubjectTokenType() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestTokenExchangeRequestPopulated(t *testing.T) {
	t.Parallel()

	r := &TokenExchangeRequest{
		GrantType:          uriGrantType,
		Resource:           []string{"https://backend.example.com/api"},
		Audience:           []string{"urn:example:cooperation-context"},
		Scope:              []string{"read", "write"},
		RequestedTokenType: uriAccessToken,
		SubjectToken:       "subject-token-value",
		SubjectTokenType:   uriIDToken,
		ActorToken:         "actor-token-value",
		ActorTokenType:     uriAccessToken,
		Extra:              url.Values{"x-extension": []string{"v"}},
	}

	if r.GrantType != uriGrantType {
		t.Errorf("GrantType = %q, want %q", r.GrantType, uriGrantType)
	}
	if len(r.Resource) != 1 || r.Resource[0] != "https://backend.example.com/api" {
		t.Errorf("Resource = %v", r.Resource)
	}
	if got := r.Extra.Get("x-extension"); got != "v" {
		t.Errorf("Extra[x-extension] = %q, want %q", got, "v")
	}
}
