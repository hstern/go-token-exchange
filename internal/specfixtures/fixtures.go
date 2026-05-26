// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

// Package specfixtures embeds the RFC 8693 example payloads the
// library uses as its conformance corpus. Each fixture carries the
// raw wire bytes from the spec plus the expected typed value, so
// tests can assert in both directions — Parse(wire) equals Want, and
// re-encoding Want produces a wire form that re-parses to the same
// Want.
//
// The package is internal-only. It exists to give the conformance
// test suite a single source of truth for what "an RFC 8693 example
// looks like in Go"; downstream consumers should not import it.
//
// Payload filenames mirror the example shape (req_impersonation.txt,
// resp_access_token.json, err_invalid_target.json) rather than the
// spec figure number, because the spec's figure numbering is
// load-bearing for the spec reader but not for the library reader.
package specfixtures

import (
	"embed"

	tokenexchange "github.com/hstern/go-token-exchange"
)

//go:embed payloads
var payloadsFS embed.FS

// mustLoad reads a fixture file from the embedded FS and panics on
// I/O failure. The panic is appropriate because the file set is
// closed (the //go:embed directive pins it at compile time); a read
// failure means the binary is malformed.
func mustLoad(name string) []byte {
	b, err := payloadsFS.ReadFile("payloads/" + name)
	if err != nil {
		panic("specfixtures: load " + name + ": " + err.Error())
	}
	return b
}

// RequestFixture is one RFC 8693 §2.1 request example: the raw
// form-encoded body plus the expected typed value Parse returns.
type RequestFixture struct {
	// Name is the human-readable identifier for the fixture, suitable
	// for t.Run subtest names.
	Name string

	// Wire is the application/x-www-form-urlencoded body bytes as
	// they appear on the wire.
	Wire []byte

	// Want is the typed value Parse must produce on Wire and that
	// Encode must round-trip back to a wire form that re-parses to
	// the same Want.
	Want *tokenexchange.TokenExchangeRequest
}

// ResponseFixture is one RFC 8693 §2.2 success-response example.
type ResponseFixture struct {
	Name string
	Wire []byte
	Want *tokenexchange.TokenExchangeResponse
}

// ErrorFixture is one RFC 6749 §5.2 / RFC 8693 §2.4 error-response
// example.
type ErrorFixture struct {
	Name string
	Wire []byte
	Want *tokenexchange.TokenExchangeError
}

// Requests returns the catalog of request fixtures in stable order.
// The returned slice and its elements are safe to read; tests that
// mutate must defensive-copy first.
func Requests() []RequestFixture {
	return []RequestFixture{
		{
			Name: "impersonation",
			Wire: mustLoad("req_impersonation.txt"),
			Want: &tokenexchange.TokenExchangeRequest{
				GrantType:        tokenexchange.GrantTypeTokenExchange,
				Resource:         []string{"https://backend.example.com/api"},
				SubjectToken:     "accVkjcJyb4BWCxGsndESCJQbdFMogUC/BLDcM3XjmKfSGGYTd1eb",
				SubjectTokenType: tokenexchange.TokenTypeAccessToken,
			},
		},
		{
			Name: "delegation",
			Wire: mustLoad("req_delegation.txt"),
			Want: &tokenexchange.TokenExchangeRequest{
				GrantType:        tokenexchange.GrantTypeTokenExchange,
				Audience:         []string{"urn:example:cooperation-context"},
				SubjectToken:     "eyJhbGciOiJFUzI1NiIsImtpZCI6IjE2In0.subject-payload",
				SubjectTokenType: tokenexchange.TokenTypeIDToken,
				ActorToken:       "eyJhbGciOiJFUzI1NiIsImtpZCI6IjcyIn0.actor-payload",
				ActorTokenType:   tokenexchange.TokenTypeAccessToken,
			},
		},
		{
			Name: "scoped_requested_type",
			Wire: mustLoad("req_scoped_requested_type.txt"),
			Want: &tokenexchange.TokenExchangeRequest{
				GrantType:          tokenexchange.GrantTypeTokenExchange,
				Resource:           []string{"https://backend.example.com/api"},
				Audience:           []string{"urn:example:cooperation-context", "urn:example:secondary-aud"},
				Scope:              []string{"read", "write"},
				RequestedTokenType: tokenexchange.TokenTypeJWT,
				SubjectToken:       "accVkjcJyb4BWCxGsndESCJQbdFMogUC/BLDcM3XjmKfSGGYTd1eb",
				SubjectTokenType:   tokenexchange.TokenTypeAccessToken,
			},
		},
	}
}

// Responses returns the catalog of success-response fixtures in
// stable order.
func Responses() []ResponseFixture {
	return []ResponseFixture{
		{
			Name: "access_token",
			Wire: mustLoad("resp_access_token.json"),
			Want: &tokenexchange.TokenExchangeResponse{
				AccessToken:     "eyJhbGciOiJFUzI1NiIsImtpZCI6IjllciJ9.access-payload",
				IssuedTokenType: tokenexchange.TokenTypeAccessToken,
				TokenType:       "Bearer",
				ExpiresIn:       60,
			},
		},
		{
			Name: "jwt_issued_with_scope",
			Wire: mustLoad("resp_jwt.json"),
			Want: &tokenexchange.TokenExchangeResponse{
				AccessToken:     "eyJhbGciOiJFUzI1NiIsImtpZCI6IjE2In0.issued-jwt-payload",
				IssuedTokenType: tokenexchange.TokenTypeJWT,
				TokenType:       "N_A",
				ExpiresIn:       300,
				Scope:           "https://backend.example.com/api",
			},
		},
		{
			Name: "access_token_with_refresh",
			Wire: mustLoad("resp_with_refresh.json"),
			Want: &tokenexchange.TokenExchangeResponse{
				AccessToken:     "ya29.access-token-payload",
				IssuedTokenType: tokenexchange.TokenTypeAccessToken,
				TokenType:       "Bearer",
				ExpiresIn:       3600,
				RefreshToken:    "1//09refresh-token-payload",
				Scope:           "openid profile",
			},
		},
	}
}

// Errors returns the catalog of error-response fixtures in stable
// order.
func Errors() []ErrorFixture {
	return []ErrorFixture{
		{
			Name: "invalid_target_with_description",
			Wire: mustLoad("err_invalid_target.json"),
			Want: &tokenexchange.TokenExchangeError{
				Code:        tokenexchange.ErrCodeInvalidTarget,
				Description: "Resource 'https://other.example.com/' is not authorized for this client.",
			},
		},
		{
			Name: "invalid_request_minimal",
			Wire: mustLoad("err_invalid_request.json"),
			Want: &tokenexchange.TokenExchangeError{
				Code: tokenexchange.ErrCodeInvalidRequest,
			},
		},
	}
}
