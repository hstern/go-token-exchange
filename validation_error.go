// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"errors"
	"fmt"
)

// Package-level sentinel errors.
//
// These cover internal failure modes that arise at the Go boundary —
// registry collisions, explicit-rejection paths, grant-type mismatch
// — not the RFC 6749 §5.2 wire codes. The wire codes live alongside
// [TokenExchangeError] in token_error.go because they are protocol
// values, not Go errors. Callers match these with errors.Is.
var (
	// ErrTokenTypeReserved is returned by RegisterTokenType when the
	// supplied URI collides with a token-type URI defined by RFC 8693
	// §3. The library reserves the built-in URIs so a downstream
	// consumer cannot accidentally shadow a spec-defined type.
	ErrTokenTypeReserved = errors.New("tokenexchange: token-type URI is reserved by RFC 8693")

	// ErrUnknownTokenType identifies an unrecognized token-type URI.
	// The library does NOT return this by default — unknown URIs
	// parse to [UnknownTokenType] so forward-compat round-trips
	// without losing data. Callers who prefer strict rejection use
	// this sentinel from their own dispatch logic.
	ErrUnknownTokenType = errors.New("tokenexchange: token-type URI is not registered")

	// ErrInvalidGrantType is returned when a parsed request's
	// grant_type parameter is not [GrantTypeTokenExchange]. The AS-
	// side helper translates this to an RFC 6749 §5.2
	// unsupported_grant_type response.
	ErrInvalidGrantType = errors.New("tokenexchange: grant_type is not the token-exchange grant")
)

// ValidationError reports that a typed [TokenExchangeRequest] or
// [TokenExchangeResponse] failed an RFC 8693 §2.1 or §2.2 MUST
// invariant.
//
// The three fields name the failure in a way an AS-side or client
// implementer can translate to either a logged operator message or a
// user-visible message:
//
//   - Rule cites the spec clause that defines the invariant (e.g.
//     "RFC 8693 §2.1") so the reader of a log line can verify the
//     library's interpretation against the source.
//   - Parameter names the wire parameter that failed (e.g.
//     "subject_token", "actor_token_type"). Empty when the failure
//     is structural rather than scoped to one parameter.
//   - Reason is a short human-readable description (e.g. "missing"
//     or "must pair with actor_token").
//
// ValidationError is a value type — Validate methods return *
// pointers because the type carries enough context that copies are
// rare and identity matters for errors.Is on a sentinel value.
type ValidationError struct {
	// Rule is the spec citation that defines the violated invariant,
	// e.g. "RFC 8693 §2.1".
	Rule string

	// Parameter is the wire parameter name whose value caused the
	// failure, e.g. "subject_token". Empty when the failure is not
	// scoped to one parameter.
	Parameter string

	// Reason is a short human-readable explanation of why the
	// invariant failed, e.g. "missing" or "must pair with
	// actor_token".
	Reason string
}

// Error returns a brief string describing the validation failure,
// prefixed with the package name so the origin is identifiable in
// mixed-source error chains. The Parameter is included only when
// non-empty; the Rule is always included so the spec citation is
// visible at the surface of any log line.
func (e *ValidationError) Error() string {
	if e == nil {
		return "tokenexchange: <nil validation error>"
	}
	if e.Parameter != "" {
		return fmt.Sprintf("tokenexchange: %s: %s: %s", e.Rule, e.Parameter, e.Reason)
	}
	return fmt.Sprintf("tokenexchange: %s: %s", e.Rule, e.Reason)
}

// Is reports whether the error matches target. It returns true when
// target is a *ValidationError whose non-empty fields all equal the
// corresponding fields on the receiver. An empty Rule, Parameter,
// or Reason in target is treated as a wildcard, so callers can
// match on Parameter alone:
//
//	if errors.Is(err, &ValidationError{Parameter: "subject_token"}) { … }
//
// Wildcarding on all three fields would match any ValidationError
// and is intentionally allowed — that is the "is this a validation
// failure at all" question.
func (e *ValidationError) Is(target error) bool {
	t, ok := target.(*ValidationError)
	if !ok {
		return false
	}
	if e == nil || t == nil {
		return false
	}
	if t.Rule != "" && t.Rule != e.Rule {
		return false
	}
	if t.Parameter != "" && t.Parameter != e.Parameter {
		return false
	}
	if t.Reason != "" && t.Reason != e.Reason {
		return false
	}
	return true
}
