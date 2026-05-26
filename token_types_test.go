// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"slices"
	"testing"
)

func TestTokenTypeURIs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		got  string
		want string
	}{
		{"GrantTypeTokenExchange", GrantTypeTokenExchange, "urn:ietf:params:oauth:grant-type:token-exchange"},
		{"TokenTypeAccessToken", TokenTypeAccessToken, "urn:ietf:params:oauth:token-type:access_token"},
		{"TokenTypeRefreshToken", TokenTypeRefreshToken, "urn:ietf:params:oauth:token-type:refresh_token"},
		{"TokenTypeIDToken", TokenTypeIDToken, "urn:ietf:params:oauth:token-type:id_token"},
		{"TokenTypeSAML1", TokenTypeSAML1, "urn:ietf:params:oauth:token-type:saml1"},
		{"TokenTypeSAML2", TokenTypeSAML2, "urn:ietf:params:oauth:token-type:saml2"},
		{"TokenTypeJWT", TokenTypeJWT, "urn:ietf:params:oauth:token-type:jwt"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.got != tc.want {
				t.Fatalf("%s = %q, want %q", tc.name, tc.got, tc.want)
			}
		})
	}
}

func TestBuiltinTokenTypes(t *testing.T) {
	t.Parallel()

	want := []string{
		TokenTypeAccessToken,
		TokenTypeRefreshToken,
		TokenTypeIDToken,
		TokenTypeSAML1,
		TokenTypeSAML2,
		TokenTypeJWT,
	}
	got := BuiltinTokenTypes()
	if !slices.Equal(got, want) {
		t.Fatalf("BuiltinTokenTypes() = %q, want %q", got, want)
	}

	got[0] = "mutated"
	again := BuiltinTokenTypes()
	if again[0] != TokenTypeAccessToken {
		t.Fatalf("BuiltinTokenTypes() must return a fresh slice; second call returned %q at index 0", again[0])
	}
}

func TestUnknownTokenType(t *testing.T) {
	t.Parallel()

	const uri = "urn:example:custom-token-type"
	u := UnknownTokenType{URI: uri}
	if u.String() != uri {
		t.Fatalf("UnknownTokenType.String() = %q, want %q", u.String(), uri)
	}

	var zero UnknownTokenType
	if zero.String() != "" {
		t.Fatalf("zero UnknownTokenType.String() = %q, want empty", zero.String())
	}
}
