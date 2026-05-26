// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"net/url"
	"slices"
	"testing"
)

func TestParamConstants(t *testing.T) {
	t.Parallel()

	cases := []struct {
		got, want string
	}{
		{ParamGrantType, "grant_type"},
		{ParamResource, "resource"},
		{ParamAudience, "audience"},
		{ParamScope, "scope"},
		{ParamRequestedTokenType, "requested_token_type"},
		{ParamSubjectToken, "subject_token"},
		{ParamSubjectTokenType, "subject_token_type"},
		{ParamActorToken, "actor_token"},
		{ParamActorTokenType, "actor_token_type"},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("got %q, want %q", tc.got, tc.want)
		}
	}
}

// TestEncodeImpersonationSpecExample is the RFC 8693 §2.1 impersonation
// example (Figure 1 — first impersonation request). It is the canonical
// reference shape for an exchange where only subject_token and
// subject_token_type are populated alongside a single resource.
func TestEncodeImpersonationSpecExample(t *testing.T) {
	t.Parallel()

	r := &TokenExchangeRequest{
		GrantType:        GrantTypeTokenExchange,
		Resource:         []string{"https://backend.example.com/api"},
		SubjectToken:     "accVkjcJyb4BWCxGsndESCJQbdFMogUC/BLDcM3XjmKfSGGYTd1eb",
		SubjectTokenType: TokenTypeAccessToken,
	}
	got := r.Encode()
	want := url.Values{
		"grant_type":         []string{GrantTypeTokenExchange},
		"resource":           []string{"https://backend.example.com/api"},
		"subject_token":      []string{"accVkjcJyb4BWCxGsndESCJQbdFMogUC/BLDcM3XjmKfSGGYTd1eb"},
		"subject_token_type": []string{TokenTypeAccessToken},
	}
	if !valuesEqual(got, want) {
		t.Errorf("Encode() = %v, want %v", got, want)
	}
}

// TestEncodeDelegationSpecExample is the RFC 8693 §2.1 delegation
// example showing actor_token / actor_token_type pairing.
func TestEncodeDelegationSpecExample(t *testing.T) {
	t.Parallel()

	r := &TokenExchangeRequest{
		GrantType:        GrantTypeTokenExchange,
		Audience:         []string{"urn:example:cooperation-context"},
		SubjectToken:     "subject-jwt",
		SubjectTokenType: TokenTypeIDToken,
		ActorToken:       "actor-jwt",
		ActorTokenType:   TokenTypeAccessToken,
	}
	got := r.Encode()
	want := url.Values{
		"grant_type":         []string{GrantTypeTokenExchange},
		"audience":           []string{"urn:example:cooperation-context"},
		"subject_token":      []string{"subject-jwt"},
		"subject_token_type": []string{TokenTypeIDToken},
		"actor_token":        []string{"actor-jwt"},
		"actor_token_type":   []string{TokenTypeAccessToken},
	}
	if !valuesEqual(got, want) {
		t.Errorf("Encode() = %v, want %v", got, want)
	}
}

func TestEncodeMultiValuedResourceAndAudience(t *testing.T) {
	t.Parallel()

	r := &TokenExchangeRequest{
		GrantType:        GrantTypeTokenExchange,
		Resource:         []string{"https://a.example.com/", "https://b.example.com/"},
		Audience:         []string{"aud-1", "aud-2", "aud-3"},
		SubjectToken:     "tok",
		SubjectTokenType: TokenTypeAccessToken,
	}
	got := r.Encode()

	if !slices.Equal(got["resource"], []string{"https://a.example.com/", "https://b.example.com/"}) {
		t.Errorf("resource slice not preserved: %v", got["resource"])
	}
	if !slices.Equal(got["audience"], []string{"aud-1", "aud-2", "aud-3"}) {
		t.Errorf("audience slice not preserved: %v", got["audience"])
	}
}

func TestEncodeScopeJoinedWithSpace(t *testing.T) {
	t.Parallel()

	r := &TokenExchangeRequest{
		GrantType:        GrantTypeTokenExchange,
		Scope:            []string{"read", "write", "admin"},
		SubjectToken:     "tok",
		SubjectTokenType: TokenTypeAccessToken,
	}
	got := r.Encode()
	if got.Get("scope") != "read write admin" {
		t.Errorf("scope = %q, want %q", got.Get("scope"), "read write admin")
	}
}

func TestEncodeEmptyScopeSliceOmitted(t *testing.T) {
	t.Parallel()

	r := &TokenExchangeRequest{
		GrantType:        GrantTypeTokenExchange,
		Scope:            nil,
		SubjectToken:     "tok",
		SubjectTokenType: TokenTypeAccessToken,
	}
	got := r.Encode()
	if _, present := got["scope"]; present {
		t.Errorf("scope present in output despite empty slice: %v", got["scope"])
	}
}

func TestEncodeOptionalFieldsOmittedWhenEmpty(t *testing.T) {
	t.Parallel()

	r := &TokenExchangeRequest{
		GrantType:        GrantTypeTokenExchange,
		SubjectToken:     "tok",
		SubjectTokenType: TokenTypeAccessToken,
	}
	got := r.Encode()

	for _, optional := range []string{
		"resource",
		"audience",
		"scope",
		"requested_token_type",
		"actor_token",
		"actor_token_type",
	} {
		if _, present := got[optional]; present {
			t.Errorf("%s emitted despite empty Go value", optional)
		}
	}
}

func TestEncodeRequiredFieldsEmittedEvenWhenEmpty(t *testing.T) {
	t.Parallel()

	var r TokenExchangeRequest
	got := r.Encode()

	for _, required := range []string{"grant_type", "subject_token", "subject_token_type"} {
		if _, present := got[required]; !present {
			t.Errorf("%s missing from Encode output of zero value", required)
		}
		if got.Get(required) != "" {
			t.Errorf("%s = %q on zero value, want empty", required, got.Get(required))
		}
	}
}

func TestEncodeExtraRoundTripsCustomKeys(t *testing.T) {
	t.Parallel()

	r := &TokenExchangeRequest{
		GrantType:        GrantTypeTokenExchange,
		SubjectToken:     "tok",
		SubjectTokenType: TokenTypeAccessToken,
		Extra: url.Values{
			"x-policy":          []string{"pol-1"},
			"x-tenant":          []string{"acme"},
			"x-multi-value-ext": []string{"a", "b"},
		},
	}
	got := r.Encode()

	if got.Get("x-policy") != "pol-1" {
		t.Errorf("x-policy = %q", got.Get("x-policy"))
	}
	if got.Get("x-tenant") != "acme" {
		t.Errorf("x-tenant = %q", got.Get("x-tenant"))
	}
	if !slices.Equal(got["x-multi-value-ext"], []string{"a", "b"}) {
		t.Errorf("x-multi-value-ext = %v", got["x-multi-value-ext"])
	}
}

// TestEncodeExtraCannotShadowBuiltinParams verifies the "built-ins win"
// rule: an Extra entry whose key matches a reserved form parameter MUST
// be discarded, never merged.
func TestEncodeExtraCannotShadowBuiltinParams(t *testing.T) {
	t.Parallel()

	r := &TokenExchangeRequest{
		GrantType:        GrantTypeTokenExchange,
		SubjectToken:     "real-subject",
		SubjectTokenType: TokenTypeAccessToken,
		Resource:         []string{"https://real.example.com/"},
		Extra: url.Values{
			"subject_token": []string{"forged-subject"},
			"grant_type":    []string{"forged-grant-type"},
			"resource":      []string{"https://forged.example.com/"},
		},
	}
	got := r.Encode()

	if got.Get("subject_token") != "real-subject" {
		t.Errorf("subject_token = %q, want real-subject (Extra shadowed)", got.Get("subject_token"))
	}
	if got.Get("grant_type") != GrantTypeTokenExchange {
		t.Errorf("grant_type = %q, want token-exchange URI (Extra shadowed)", got.Get("grant_type"))
	}
	if !slices.Equal(got["resource"], []string{"https://real.example.com/"}) {
		t.Errorf("resource = %v, want only real entry (Extra shadowed)", got["resource"])
	}
}

// valuesEqual compares two url.Values by key sets and per-key slice
// equality, ignoring map iteration order.
func valuesEqual(a, b url.Values) bool {
	if len(a) != len(b) {
		return false
	}
	for k, av := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		if !slices.Equal(av, bv) {
			return false
		}
	}
	return true
}
