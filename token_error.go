// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"errors"
	"fmt"
)

// RFC 6749 §5.2 + RFC 8693 §2.4 error codes.
//
// The OAuth 2.0 token endpoint returns errors as a JSON object whose
// "error" member is one of a fixed set of codes. RFC 6749 §5.2
// registered the first six values; RFC 8693 §2.4 added the seventh
// (ErrCodeInvalidTarget) for resource and audience validation
// failures. AS implementations MUST emit one of these values; client
// implementations MUST be prepared to receive any of them.
//
// The library uses these constants for both emission (AS-side
// helpers) and dispatch (errors.Is matches by Code).
const (
	// ErrCodeInvalidRequest indicates the request is missing a
	// required parameter, contains an unsupported parameter value
	// (other than grant type), repeats a parameter, includes
	// multiple credentials, uses more than one mechanism for
	// authenticating the client, or is otherwise malformed.
	// RFC 6749 §5.2.
	ErrCodeInvalidRequest = "invalid_request"

	// ErrCodeInvalidClient indicates client authentication failed
	// (e.g. unknown client, no client authentication included, or
	// unsupported authentication method). The AS responds with
	// HTTP 401 (instead of HTTP 400) when this code is used.
	// RFC 6749 §5.2.
	ErrCodeInvalidClient = "invalid_client"

	// ErrCodeInvalidGrant indicates the provided authorization
	// grant (in token exchange, the subject_token or actor_token)
	// is invalid, expired, revoked, does not match the redirection
	// URI used in the authorization request, or was issued to
	// another client. RFC 6749 §5.2.
	ErrCodeInvalidGrant = "invalid_grant"

	// ErrCodeUnauthorizedClient indicates the authenticated client
	// is not authorized to use this authorization grant type.
	// RFC 6749 §5.2.
	ErrCodeUnauthorizedClient = "unauthorized_client"

	// ErrCodeUnsupportedGrantType indicates the authorization
	// grant type is not supported by the AS. RFC 6749 §5.2.
	ErrCodeUnsupportedGrantType = "unsupported_grant_type"

	// ErrCodeInvalidScope indicates the requested scope is
	// invalid, unknown, malformed, or exceeds the scope granted by
	// the resource owner. RFC 6749 §5.2.
	ErrCodeInvalidScope = "invalid_scope"

	// ErrCodeInvalidTarget indicates the requested resource or
	// audience is unknown, malformed, or not authorized for the
	// client. RFC 8693 §2.4.
	ErrCodeInvalidTarget = "invalid_target"
)

// TokenExchangeError is the typed Go form of the RFC 6749 §5.2 token-
// endpoint error response (extended by RFC 8693 §2.4 with the
// invalid_target code).
//
// The wire encoding is application/json. Two of the three fields are
// optional and use omitempty so an error with only a Code round-trips
// to the minimal JSON object {"error":"<code>"}.
//
// HTTP status mapping (RFC 6749 §5.2): all codes map to HTTP 400
// except invalid_client, which maps to HTTP 401. The AS-side helper
// WriteTokenExchangeError encodes this mapping; this type carries
// only the wire shape, not the transport semantics.
//
// TokenExchangeError implements the error interface and supports
// errors.Is matching by Code via the Is method, so client code can
// write:
//
//	if errors.Is(err, &TokenExchangeError{Code: ErrCodeInvalidTarget}) { … }
//
// or, equivalently, hold the sentinel value as a package-level var.
//
// Naming: same TokenExchange prefix as TokenExchangeRequest and
// TokenExchangeResponse, for the same spec-name reason.
//
//nolint:revive // spec-mandated TokenExchange* names; see TokenExchangeRequest
type TokenExchangeError struct {
	// Code is the RFC 6749 §5.2 error code. One of the ErrCode*
	// constants in this package, or an extension code defined by
	// a downstream profile. Required; an empty Code is a wire
	// shape the AS must not emit and a client must treat as
	// invalid.
	Code string `json:"error"`

	// Description is a human-readable explanation suitable for
	// surfacing to a developer. Optional.
	Description string `json:"error_description,omitempty"`

	// URI is a URI identifying a human-readable web page with
	// information about the error. Optional.
	URI string `json:"error_uri,omitempty"`
}

// Error returns a brief string describing the error, suitable for
// logging or wrapping. It combines Code with Description when
// Description is non-empty; otherwise it returns Code alone, prefixed
// with the package name so the origin is identifiable in mixed-
// source error chains.
func (e *TokenExchangeError) Error() string {
	if e == nil {
		return "tokenexchange: <nil error>"
	}
	if e.Description != "" {
		return fmt.Sprintf("tokenexchange: %s: %s", e.Code, e.Description)
	}
	return fmt.Sprintf("tokenexchange: %s", e.Code)
}

// Is reports whether the error matches target. It returns true when
// target is a *TokenExchangeError with the same Code, ignoring
// Description and URI. This lets callers test for a specific RFC
// 6749 §5.2 condition without constructing a value that matches the
// description text:
//
//	if errors.Is(err, &TokenExchangeError{Code: ErrCodeInvalidTarget}) { … }
//
// Is does NOT match across types (a wrapped *TokenExchangeError
// still matches; an error whose string contains the code does not).
func (e *TokenExchangeError) Is(target error) bool {
	var t *TokenExchangeError
	if !errors.As(target, &t) {
		return false
	}
	return e != nil && t != nil && e.Code == t.Code
}
