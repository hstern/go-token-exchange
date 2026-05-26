// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"
)

// Forward-compatibility tests pin the two extension axes the
// library exposes:
//
//  1. A token-type URI the library does not recognize must round-
//     trip as a raw string (the codec stores URIs as strings, and
//     UnknownTokenType wraps them at the type-system boundary
//     described by TKEX-7).
//  2. A form / JSON parameter the library does not define must
//     round-trip via the Extra capture (request: url.Values;
//     response: map[string]json.RawMessage).
//
// Neither axis exists for the spec's benefit — both exist so a
// payload generated against a future RFC 8693 profile passes
// through this library without code changes.

const (
	futureTokenTypeURI = "urn:example:future:token-type"
	futureExtraParam   = "x-future-extension"
)

// TestForwardCompatUnknownRequestTokenType verifies that a
// subject_token_type URI not in the registry parses, validates, and
// re-encodes losslessly. UnknownTokenType wraps the URI at the type
// boundary; the wire is the URI string.
func TestForwardCompatUnknownRequestTokenType(t *testing.T) {
	t.Parallel()

	if IsRegisteredTokenType(futureTokenTypeURI) {
		t.Fatalf("test precondition failed: %q is registered", futureTokenTypeURI)
	}

	body := url.Values{
		"grant_type":         []string{GrantTypeTokenExchange},
		"subject_token":      []string{"future-subject-payload"},
		"subject_token_type": []string{futureTokenTypeURI},
	}
	req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	parsed, err := ParseTokenExchangeRequest(req)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if parsed.SubjectTokenType != futureTokenTypeURI {
		t.Errorf("SubjectTokenType = %q, want %q", parsed.SubjectTokenType, futureTokenTypeURI)
	}

	if err := parsed.Validate(); err != nil {
		t.Errorf("Validate rejected an unregistered but URI-shaped token type: %v", err)
	}

	wrapper := UnknownTokenType{URI: parsed.SubjectTokenType}
	if wrapper.String() != futureTokenTypeURI {
		t.Errorf("UnknownTokenType.String() = %q, want %q", wrapper.String(), futureTokenTypeURI)
	}

	out := parsed.Encode()
	if out.Get("subject_token_type") != futureTokenTypeURI {
		t.Errorf("re-encoded subject_token_type = %q, want %q", out.Get("subject_token_type"), futureTokenTypeURI)
	}

	// Second pass must produce the same typed value.
	second := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(out.Encode()))
	second.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	secondParsed, err := ParseTokenExchangeRequest(second)
	if err != nil {
		t.Fatalf("Parse second: %v", err)
	}
	if secondParsed.SubjectTokenType != futureTokenTypeURI {
		t.Errorf("second Parse SubjectTokenType = %q", secondParsed.SubjectTokenType)
	}
}

// TestForwardCompatUnknownRequestParam verifies that an unrecognized
// form parameter lands in Extra, round-trips through Encode, and
// re-parses to the same value.
func TestForwardCompatUnknownRequestParam(t *testing.T) {
	t.Parallel()

	body := url.Values{
		"grant_type":         []string{GrantTypeTokenExchange},
		"subject_token":      []string{"subject"},
		"subject_token_type": []string{TokenTypeAccessToken},
		futureExtraParam:     []string{"first", "second"},
	}
	req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	parsed, err := ParseTokenExchangeRequest(req)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !slices.Equal(parsed.Extra[futureExtraParam], []string{"first", "second"}) {
		t.Errorf("Extra[%s] = %v, want [first second]", futureExtraParam, parsed.Extra[futureExtraParam])
	}

	out := parsed.Encode()
	if !slices.Equal(out[futureExtraParam], []string{"first", "second"}) {
		t.Errorf("re-encoded %s = %v, want [first second]", futureExtraParam, out[futureExtraParam])
	}

	second := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(out.Encode()))
	second.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	secondParsed, err := ParseTokenExchangeRequest(second)
	if err != nil {
		t.Fatalf("Parse second: %v", err)
	}
	if !slices.Equal(secondParsed.Extra[futureExtraParam], []string{"first", "second"}) {
		t.Errorf("second Parse Extra[%s] = %v", futureExtraParam, secondParsed.Extra[futureExtraParam])
	}
}

// TestForwardCompatUnknownResponseIssuedTokenType verifies the
// response side: a future token-type URI in issued_token_type
// round-trips through the JSON codec without error.
func TestForwardCompatUnknownResponseIssuedTokenType(t *testing.T) {
	t.Parallel()

	resp := TokenExchangeResponse{
		AccessToken:     "issued-payload",
		IssuedTokenType: futureTokenTypeURI,
		TokenType:       "Bearer",
	}
	if err := resp.Validate(); err != nil {
		t.Fatalf("Validate rejected unregistered issued_token_type: %v", err)
	}

	out, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(out), `"issued_token_type":"urn:example:future:token-type"`) {
		t.Errorf("Marshal missing issued_token_type: %s", out)
	}

	var back TokenExchangeResponse
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if back.IssuedTokenType != futureTokenTypeURI {
		t.Errorf("round-trip IssuedTokenType = %q", back.IssuedTokenType)
	}
}

// TestForwardCompatUnknownResponseMember verifies that an unknown
// JSON member in a response lands in Extra and re-emits on Marshal.
func TestForwardCompatUnknownResponseMember(t *testing.T) {
	t.Parallel()

	const wire = `{
		"access_token": "tok",
		"issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
		"token_type": "Bearer",
		"x-future-extension": {"nested": true, "level": 2}
	}`

	var resp TokenExchangeResponse
	if err := json.Unmarshal([]byte(wire), &resp); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	raw, ok := resp.Extra[futureExtraParam]
	if !ok {
		t.Fatalf("Extra missing %q", futureExtraParam)
	}

	// Canonicalize the captured RawMessage by re-encoding so the
	// comparison is whitespace-insensitive.
	var canon any
	if err := json.Unmarshal(raw, &canon); err != nil {
		t.Fatalf("Extra raw not valid JSON: %v", err)
	}
	canonBytes, err := json.Marshal(canon)
	if err != nil {
		t.Fatalf("canonicalize: %v", err)
	}
	if string(canonBytes) != `{"level":2,"nested":true}` {
		t.Errorf("Extra raw canonicalized = %s, want {\"level\":2,\"nested\":true}", canonBytes)
	}

	// Marshal must re-emit the unknown member.
	out, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(out), `"x-future-extension"`) {
		t.Errorf("Marshal dropped Extra member: %s", out)
	}

	// Round-trip closure.
	var second TokenExchangeResponse
	if err := json.Unmarshal(out, &second); err != nil {
		t.Fatalf("Unmarshal second: %v", err)
	}
	if _, ok := second.Extra[futureExtraParam]; !ok {
		t.Errorf("second Unmarshal Extra missing %q", futureExtraParam)
	}
}

// TestForwardCompatErrorUnknownCode verifies that an error code
// outside the seven RFC 6749 §5.2 / RFC 8693 §2.4 sentinels decodes
// without error — extensions can define new codes.
func TestForwardCompatErrorUnknownCode(t *testing.T) {
	t.Parallel()

	const wire = `{"error":"downstream:rate_limited","error_description":"slow down"}`
	var te TokenExchangeError
	if err := json.Unmarshal([]byte(wire), &te); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if te.Code != "downstream:rate_limited" {
		t.Errorf("Code = %q", te.Code)
	}

	out, err := json.Marshal(te)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(out), `"error":"downstream:rate_limited"`) {
		t.Errorf("Marshal output = %s", out)
	}
}
