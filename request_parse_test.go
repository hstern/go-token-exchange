// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"
)

func newFormRequest(t *testing.T, body url.Values) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func TestParseImpersonationSpecExample(t *testing.T) {
	t.Parallel()

	body := url.Values{
		"grant_type":         []string{GrantTypeTokenExchange},
		"resource":           []string{"https://backend.example.com/api"},
		"subject_token":      []string{"accVkjcJyb4BWCxGsndESCJQbdFMogUC/BLDcM3XjmKfSGGYTd1eb"},
		"subject_token_type": []string{TokenTypeAccessToken},
	}
	got, err := ParseTokenExchangeRequest(newFormRequest(t, body))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.GrantType != GrantTypeTokenExchange ||
		got.SubjectToken != "accVkjcJyb4BWCxGsndESCJQbdFMogUC/BLDcM3XjmKfSGGYTd1eb" ||
		got.SubjectTokenType != TokenTypeAccessToken {
		t.Errorf("required fields: %+v", got)
	}
	if !slices.Equal(got.Resource, []string{"https://backend.example.com/api"}) {
		t.Errorf("Resource = %v", got.Resource)
	}
	if got.Extra != nil {
		t.Errorf("Extra = %v, want nil", got.Extra)
	}
}

func TestParseDelegationSpecExample(t *testing.T) {
	t.Parallel()

	body := url.Values{
		"grant_type":         []string{GrantTypeTokenExchange},
		"audience":           []string{"urn:example:cooperation-context"},
		"subject_token":      []string{"subject-jwt"},
		"subject_token_type": []string{TokenTypeIDToken},
		"actor_token":        []string{"actor-jwt"},
		"actor_token_type":   []string{TokenTypeAccessToken},
	}
	got, err := ParseTokenExchangeRequest(newFormRequest(t, body))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.ActorToken != "actor-jwt" || got.ActorTokenType != TokenTypeAccessToken {
		t.Errorf("actor pair: ActorToken=%q ActorTokenType=%q", got.ActorToken, got.ActorTokenType)
	}
	if !slices.Equal(got.Audience, []string{"urn:example:cooperation-context"}) {
		t.Errorf("Audience = %v", got.Audience)
	}
}

func TestParseMultiValuedResourceAudience(t *testing.T) {
	t.Parallel()

	body := url.Values{
		"grant_type":         []string{GrantTypeTokenExchange},
		"resource":           []string{"https://a.example.com/", "https://b.example.com/"},
		"audience":           []string{"aud-1", "aud-2", "aud-3"},
		"subject_token":      []string{"tok"},
		"subject_token_type": []string{TokenTypeAccessToken},
	}
	got, err := ParseTokenExchangeRequest(newFormRequest(t, body))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !slices.Equal(got.Resource, []string{"https://a.example.com/", "https://b.example.com/"}) {
		t.Errorf("Resource = %v", got.Resource)
	}
	if !slices.Equal(got.Audience, []string{"aud-1", "aud-2", "aud-3"}) {
		t.Errorf("Audience = %v", got.Audience)
	}
}

func TestParseScopeSplitsOnWhitespace(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"single", "read", []string{"read"}},
		{"multiple", "read write admin", []string{"read", "write", "admin"}},
		{"extra spaces", "  read   write  ", []string{"read", "write"}},
		{"empty omitted", "", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			body := url.Values{
				"grant_type":         []string{GrantTypeTokenExchange},
				"scope":              []string{tc.in},
				"subject_token":      []string{"tok"},
				"subject_token_type": []string{TokenTypeAccessToken},
			}
			got, err := ParseTokenExchangeRequest(newFormRequest(t, body))
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if !slices.Equal(got.Scope, tc.want) {
				t.Errorf("Scope = %v, want %v", got.Scope, tc.want)
			}
		})
	}
}

func TestParseCapturesUnknownIntoExtra(t *testing.T) {
	t.Parallel()

	body := url.Values{
		"grant_type":         []string{GrantTypeTokenExchange},
		"subject_token":      []string{"tok"},
		"subject_token_type": []string{TokenTypeAccessToken},
		"x-tenant":           []string{"acme"},
		"x-policy":           []string{"strict"},
		"x-multi":            []string{"a", "b"},
	}
	got, err := ParseTokenExchangeRequest(newFormRequest(t, body))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.Extra.Get("x-tenant") != "acme" {
		t.Errorf("Extra[x-tenant] = %q", got.Extra.Get("x-tenant"))
	}
	if got.Extra.Get("x-policy") != "strict" {
		t.Errorf("Extra[x-policy] = %q", got.Extra.Get("x-policy"))
	}
	if !slices.Equal(got.Extra["x-multi"], []string{"a", "b"}) {
		t.Errorf("Extra[x-multi] = %v", got.Extra["x-multi"])
	}
}

func TestParseLeavesExtraNilWhenNoUnknownParams(t *testing.T) {
	t.Parallel()

	body := url.Values{
		"grant_type":         []string{GrantTypeTokenExchange},
		"subject_token":      []string{"tok"},
		"subject_token_type": []string{TokenTypeAccessToken},
	}
	got, err := ParseTokenExchangeRequest(newFormRequest(t, body))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.Extra != nil {
		t.Errorf("Extra = %v, want nil", got.Extra)
	}
}

// TestParseDoesNotValidate asserts the Postel's-law receive policy:
// missing required fields decode without error, leaving the zero value
// for Validate to flag.
func TestParseDoesNotValidate(t *testing.T) {
	t.Parallel()

	body := url.Values{
		// no grant_type, no subject_token, no subject_token_type
		"x-anything": []string{"sure"},
	}
	got, err := ParseTokenExchangeRequest(newFormRequest(t, body))
	if err != nil {
		t.Fatalf("Parse returned error on missing required fields: %v", err)
	}
	if got.GrantType != "" || got.SubjectToken != "" || got.SubjectTokenType != "" {
		t.Errorf("missing fields produced non-zero values: %+v", got)
	}
}

// TestParseEncodeRoundTrip verifies that Parse and Encode are inverses
// for a wire payload whose shape Encode can produce.
func TestParseEncodeRoundTrip(t *testing.T) {
	t.Parallel()

	original := &TokenExchangeRequest{
		GrantType:          GrantTypeTokenExchange,
		Resource:           []string{"https://a.example.com/", "https://b.example.com/"},
		Audience:           []string{"aud-1"},
		Scope:              []string{"read", "write"},
		RequestedTokenType: TokenTypeJWT,
		SubjectToken:       "subject-token",
		SubjectTokenType:   TokenTypeAccessToken,
		ActorToken:         "actor-token",
		ActorTokenType:     TokenTypeAccessToken,
		Extra: url.Values{
			"x-policy": []string{"strict"},
		},
	}
	body := original.Encode()
	parsed, err := ParseTokenExchangeRequest(newFormRequest(t, body))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if parsed.GrantType != original.GrantType ||
		parsed.SubjectToken != original.SubjectToken ||
		parsed.SubjectTokenType != original.SubjectTokenType ||
		parsed.RequestedTokenType != original.RequestedTokenType ||
		parsed.ActorToken != original.ActorToken ||
		parsed.ActorTokenType != original.ActorTokenType {
		t.Errorf("string fields drifted: %+v vs %+v", parsed, original)
	}
	if !slices.Equal(parsed.Resource, original.Resource) {
		t.Errorf("Resource drifted: %v vs %v", parsed.Resource, original.Resource)
	}
	if !slices.Equal(parsed.Audience, original.Audience) {
		t.Errorf("Audience drifted: %v vs %v", parsed.Audience, original.Audience)
	}
	if !slices.Equal(parsed.Scope, original.Scope) {
		t.Errorf("Scope drifted: %v vs %v", parsed.Scope, original.Scope)
	}
	if parsed.Extra.Get("x-policy") != "strict" {
		t.Errorf("Extra[x-policy] drifted: %q", parsed.Extra.Get("x-policy"))
	}
}

func TestParseRejectsNilRequest(t *testing.T) {
	t.Parallel()

	_, err := ParseTokenExchangeRequest(nil)
	if err == nil {
		t.Errorf("Parse(nil) returned no error")
	}
}

// TestParseExtraIsDefensiveCopy verifies the parsed Extra does not
// alias the request's r.Form storage; mutating r.Form after Parse
// must not change the parsed value.
func TestParseExtraIsDefensiveCopy(t *testing.T) {
	t.Parallel()

	body := url.Values{
		"grant_type":         []string{GrantTypeTokenExchange},
		"subject_token":      []string{"tok"},
		"subject_token_type": []string{TokenTypeAccessToken},
		"x-policy":           []string{"strict"},
	}
	req := newFormRequest(t, body)
	parsed, err := ParseTokenExchangeRequest(req)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Mutate the form after parse.
	req.Form["x-policy"][0] = "tampered"
	if got := parsed.Extra.Get("x-policy"); got != "strict" {
		t.Errorf("Extra aliased r.Form: got %q after tamper, want strict", got)
	}
}
