// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

// ruleResponse is the spec citation surfaced by [ValidationError] for
// every MUST defined in RFC 8693 §2.2.
const ruleResponse = "RFC 8693 §2.2"

// Validate reports whether r satisfies the RFC 8693 §2.2 MUST rules
// for a successful token-exchange response:
//
//   - access_token MUST be present and non-empty.
//   - issued_token_type MUST be present, non-empty, and a
//     syntactically valid URI per RFC 3986.
//   - token_type MUST be present and non-empty.
//
// The first failing rule returns a [*ValidationError] naming the
// rule citation, the wire field name, and a short reason. Validate
// does NOT enforce that issued_token_type be a registered URI —
// forward-compat for IANA registry growth is handled by
// [UnknownTokenType].
//
// Validate on a nil receiver returns a [*ValidationError] rather
// than panicking; the client-side Exchange uses this so a server
// that returns the literal JSON "null" surfaces as a validation
// failure rather than a panic.
func (r *TokenExchangeResponse) Validate() error {
	if r == nil {
		return &ValidationError{Rule: ruleResponse, Reason: "nil response"}
	}

	if r.AccessToken == "" {
		return &ValidationError{
			Rule:      ruleResponse,
			Parameter: FieldAccessToken,
			Reason:    "missing",
		}
	}
	if r.IssuedTokenType == "" {
		return &ValidationError{
			Rule:      ruleResponse,
			Parameter: FieldIssuedTokenType,
			Reason:    "missing",
		}
	}
	if !isValidTokenTypeURI(r.IssuedTokenType) {
		return &ValidationError{
			Rule:      ruleResponse,
			Parameter: FieldIssuedTokenType,
			Reason:    "not a valid URI",
		}
	}
	if r.TokenType == "" {
		return &ValidationError{
			Rule:      ruleResponse,
			Parameter: FieldTokenType,
			Reason:    "missing",
		}
	}

	return nil
}
