// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Extension tests exercise the registry surface — RegisterTokenType
// adds a downstream URI, IsRegisteredTokenType reports membership,
// and parse / validate / encode all accept the new URI without
// special casing. The smoke covers the happy path; the built-in
// collision test pins the reservation rule that protects spec URIs.
//
// Tests that mutate the package-global registry are NOT t.Parallel
// to avoid interleaving with each other; they DO share the
// resetRegistry helper exported by the registry test file so the
// global is restored between runs.

const (
	extensionTokenTypeURI = "urn:example:extension-token-type"
)

// TestExtensionRegisterAndRoundTrip walks the canonical extension
// flow: register a custom URI, parse a request that uses it,
// validate, and re-encode to a wire form that re-parses cleanly.
func TestExtensionRegisterAndRoundTrip(t *testing.T) {
	t.Cleanup(func() { resetRegistry(t) })

	if IsRegisteredTokenType(extensionTokenTypeURI) {
		t.Fatalf("test precondition failed: extension URI already registered")
	}
	if err := RegisterTokenType(extensionTokenTypeURI); err != nil {
		t.Fatalf("RegisterTokenType: %v", err)
	}
	if !IsRegisteredTokenType(extensionTokenTypeURI) {
		t.Fatalf("IsRegisteredTokenType reports false after registration")
	}

	body := url.Values{
		"grant_type":         []string{GrantTypeTokenExchange},
		"subject_token":      []string{"extension-subject"},
		"subject_token_type": []string{extensionTokenTypeURI},
	}
	req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	parsed, err := ParseTokenExchangeRequest(req)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if parsed.SubjectTokenType != extensionTokenTypeURI {
		t.Errorf("SubjectTokenType = %q, want %q", parsed.SubjectTokenType, extensionTokenTypeURI)
	}
	if err := parsed.Validate(); err != nil {
		t.Errorf("Validate: %v", err)
	}

	out := parsed.Encode()
	if got := out.Get("subject_token_type"); got != extensionTokenTypeURI {
		t.Errorf("re-encoded subject_token_type = %q", got)
	}

	second := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(out.Encode()))
	second.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	secondParsed, err := ParseTokenExchangeRequest(second)
	if err != nil {
		t.Fatalf("Parse second: %v", err)
	}
	if err := secondParsed.Validate(); err != nil {
		t.Errorf("Validate second: %v", err)
	}
}

// TestExtensionResponseRoundTripWithExtensionType verifies the
// response side — issuing a token whose type is a registered
// extension URI Marshal/Unmarshal cleanly and survives Validate.
func TestExtensionResponseRoundTripWithExtensionType(t *testing.T) {
	t.Cleanup(func() { resetRegistry(t) })

	if err := RegisterTokenType(extensionTokenTypeURI); err != nil {
		t.Fatalf("RegisterTokenType: %v", err)
	}

	resp := &TokenExchangeResponse{
		AccessToken:     "extension-payload",
		IssuedTokenType: extensionTokenTypeURI,
		TokenType:       "Bearer",
		ExpiresIn:       60,
	}
	if err := resp.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	rr := httptest.NewRecorder()
	if err := WriteTokenExchangeResponse(rr, resp); err != nil {
		t.Fatalf("WriteTokenExchangeResponse: %v", err)
	}
	if !strings.Contains(rr.Body.String(), extensionTokenTypeURI) {
		t.Errorf("Write output missing extension URI: %s", rr.Body.String())
	}
}

// TestExtensionRegisterBuiltinReturnsErrTokenTypeReserved pins the
// reservation rule: every RFC 8693 §3 built-in URI must reject
// re-registration. The test walks the full built-in set rather than
// spot-checking one, because the failure mode (a built-in suddenly
// shadowable) is a v0-versus-v1 compatibility risk if it ever
// regresses.
func TestExtensionRegisterBuiltinReturnsErrTokenTypeReserved(t *testing.T) {
	t.Parallel()

	for _, uri := range BuiltinTokenTypes() {
		t.Run(uri, func(t *testing.T) {
			t.Parallel()
			err := RegisterTokenType(uri)
			if err == nil {
				t.Fatalf("RegisterTokenType(%q) returned nil", uri)
			}
			if !errors.Is(err, ErrTokenTypeReserved) {
				t.Errorf("err = %v, want ErrTokenTypeReserved in chain", err)
			}
		})
	}
}

// TestExtensionIdempotentRegistration confirms a consumer can run
// registration safely on every process start without conditional
// checks — the second RegisterTokenType for a previously-registered
// extension URI is a no-op.
func TestExtensionIdempotentRegistration(t *testing.T) {
	t.Cleanup(func() { resetRegistry(t) })

	if err := RegisterTokenType(extensionTokenTypeURI); err != nil {
		t.Fatalf("first RegisterTokenType: %v", err)
	}
	if err := RegisterTokenType(extensionTokenTypeURI); err != nil {
		t.Errorf("idempotent re-registration returned %v, want nil", err)
	}
}
