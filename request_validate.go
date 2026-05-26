// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"fmt"
	"net/url"
)

// ruleRequest is the spec citation surfaced by [ValidationError] for
// every MUST defined in RFC 8693 §2.1.
const ruleRequest = "RFC 8693 §2.1"

// Validate reports whether r satisfies the RFC 8693 §2.1 MUST rules
// for a token-exchange request:
//
//   - grant_type MUST equal [GrantTypeTokenExchange].
//   - subject_token MUST be present and non-empty.
//   - subject_token_type MUST be present and non-empty.
//   - actor_token and actor_token_type MUST both be present or both
//     absent (paired).
//   - subject_token_type, requested_token_type (when present), and
//     actor_token_type (when present) MUST each be a syntactically
//     valid URI per RFC 3986.
//
// The first failing rule returns a [*ValidationError] naming the
// rule, the wire parameter, and a short reason. Validate does not
// inspect contents beyond the spec MUSTs — token-type URIs are not
// required to be registered, and token contents (JWT shape, opacity,
// signatures) are out of scope.
//
// A grant_type that is not the token-exchange URI returns the
// sentinel [ErrInvalidGrantType] wrapped with the validation
// citation, so callers can match both shapes:
//
//	if errors.Is(err, ErrInvalidGrantType) { … }
//	if errors.Is(err, &ValidationError{Parameter: ParamGrantType}) { … }
//
// Validate on a nil receiver returns a [*ValidationError] rather
// than panicking; the AS-side helper uses that to translate to a
// 400 invalid_request without bespoke nil handling.
func (r *TokenExchangeRequest) Validate() error {
	if r == nil {
		return &ValidationError{Rule: ruleRequest, Reason: "nil request"}
	}

	if r.GrantType != GrantTypeTokenExchange {
		return fmt.Errorf("%w: %w", ErrInvalidGrantType, &ValidationError{
			Rule:      ruleRequest,
			Parameter: ParamGrantType,
			Reason:    "must equal " + GrantTypeTokenExchange,
		})
	}

	if r.SubjectToken == "" {
		return &ValidationError{
			Rule:      ruleRequest,
			Parameter: ParamSubjectToken,
			Reason:    "missing",
		}
	}
	if r.SubjectTokenType == "" {
		return &ValidationError{
			Rule:      ruleRequest,
			Parameter: ParamSubjectTokenType,
			Reason:    "missing",
		}
	}
	if !isValidTokenTypeURI(r.SubjectTokenType) {
		return &ValidationError{
			Rule:      ruleRequest,
			Parameter: ParamSubjectTokenType,
			Reason:    "not a valid URI",
		}
	}

	// actor_token and actor_token_type are paired. Both present or
	// both absent; mixed presence is the §2.1 violation that bites
	// every implementer once.
	switch {
	case r.ActorToken == "" && r.ActorTokenType != "":
		return &ValidationError{
			Rule:      ruleRequest,
			Parameter: ParamActorToken,
			Reason:    "must be present when " + ParamActorTokenType + " is set",
		}
	case r.ActorToken != "" && r.ActorTokenType == "":
		return &ValidationError{
			Rule:      ruleRequest,
			Parameter: ParamActorTokenType,
			Reason:    "must be present when " + ParamActorToken + " is set",
		}
	case r.ActorTokenType != "" && !isValidTokenTypeURI(r.ActorTokenType):
		return &ValidationError{
			Rule:      ruleRequest,
			Parameter: ParamActorTokenType,
			Reason:    "not a valid URI",
		}
	}

	if r.RequestedTokenType != "" && !isValidTokenTypeURI(r.RequestedTokenType) {
		return &ValidationError{
			Rule:      ruleRequest,
			Parameter: ParamRequestedTokenType,
			Reason:    "not a valid URI",
		}
	}

	return nil
}

// isValidTokenTypeURI reports whether s is a syntactically valid URI
// per RFC 3986 — i.e. has a scheme component followed by ":". This
// is the shape check the spec requires for token-type URIs; the
// library does not enforce that the URI be registered (forward-
// compatibility for IANA registry growth is handled by
// [UnknownTokenType]).
func isValidTokenTypeURI(s string) bool {
	if s == "" {
		return false
	}
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	// RFC 3986 §3.1: a URI must have a scheme. url.Parse on a bare
	// path or fragment returns no error but leaves Scheme empty.
	return u.Scheme != ""
}
