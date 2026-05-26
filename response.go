// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import "encoding/json"

// TokenExchangeResponse is the typed Go form of the RFC 8693 §2.2
// successful token-exchange response.
//
// The wire encoding is application/json (RFC 6749 §5.1 extended by
// RFC 8693 §2.2), so the struct carries encoding/json tags and is
// serialized by encoding/json. Custom Marshal and Unmarshal methods
// that round-trip the Extra map land in a later phase; until then
// Extra is annotated json:"-" so the default codec does not visit it
// (an empty Extra map would otherwise serialize as a missing field
// noisily, and a populated one would not survive a round trip).
//
// Field invariants matching RFC 8693 §2.2:
//
//   - AccessToken, IssuedTokenType, and TokenType MUST be present
//     and non-empty. Validate enforces all three.
//   - ExpiresIn, Scope, and RefreshToken are OPTIONAL. Zero values
//     are serialized as missing fields via omitempty.
//
// AccessToken note: the wire field is named access_token regardless
// of the type of the issued security token. RFC 8693 §2.2.1 inherits
// the field name from RFC 6749 §5.1, so an exchange that issues a
// JWT id_token still carries the JWT in access_token; the typed
// access lives in IssuedTokenType. Surprising once; permanent.
//
// Extra captures JSON object members this library does not define.
// It round-trips through the custom Marshal and Unmarshal that
// arrive in a later phase, so a spec extension that adds a new
// response field passes through unchanged; built-in fields take
// precedence when a key in Extra collides with a field name.
//
// Naming: the TokenExchange prefix mirrors the spec name (RFC 8693)
// and matches the request and error types in this package; see
// TokenExchangeRequest for the longer explanation.
//
//nolint:revive // spec-mandated TokenExchange* names; see TokenExchangeRequest
type TokenExchangeResponse struct {
	// AccessToken carries the issued security token, regardless of
	// type. RFC 8693 §2.2.1 inherits the field name from RFC 6749
	// §5.1, so an issued id_token, JWT, or SAML assertion also
	// arrives here. Required; Validate rejects an empty value.
	AccessToken string `json:"access_token"`

	// IssuedTokenType is the token-type URI naming the type of
	// AccessToken. Required; Validate rejects an empty value.
	IssuedTokenType string `json:"issued_token_type"`

	// TokenType is the OAuth 2.0 token_type response parameter
	// (RFC 6749 §7.1 — typically "Bearer", or "N_A" when the
	// issued token is not an access token). Required; Validate
	// rejects an empty value.
	TokenType string `json:"token_type"`

	// ExpiresIn is the recommended lifetime of AccessToken in
	// seconds. Optional; zero means the AS did not advertise a
	// lifetime.
	ExpiresIn int `json:"expires_in,omitempty"`

	// Scope is the scope granted to AccessToken, as a single
	// space-separated string per RFC 6749 §3.3. Present only when
	// the AS narrows or otherwise differs from the requested scope.
	Scope string `json:"scope,omitempty"`

	// RefreshToken is an optional refresh token the AS may issue
	// alongside the exchange result.
	RefreshToken string `json:"refresh_token,omitempty"`

	// Extra holds JSON object members this library does not define.
	// The tag json:"-" suppresses the default codec; the custom
	// Marshal and Unmarshal that arrive in a later phase round-trip
	// Extra through the JSON object.
	Extra map[string]json.RawMessage `json:"-"`
}
