// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import "net/url"

// TokenExchangeRequest is the typed Go form of the RFC 8693 §2.1
// token-exchange request body.
//
// The wire encoding is application/x-www-form-urlencoded (RFC 6749
// §3.2 inherited convention), not JSON; the struct therefore carries
// no encoding/json tags. Encode and parse helpers land alongside the
// codec in a later phase, and they convert to and from net/url.Values.
//
// Field invariants matching RFC 8693 §2.1:
//
//   - GrantType is always [GrantTypeTokenExchange]. Validate() rejects
//     any other value.
//   - SubjectToken and SubjectTokenType MUST be present and non-empty.
//   - ActorToken and ActorTokenType MUST both be present or both
//     absent. Validate() enforces the pairing.
//   - Resource and Audience are slices because the spec allows each
//     parameter to appear multiple times on the wire; the slice
//     preserves wire order on encode.
//   - Scope is a slice because RFC 6749 §3.3 defines the scope value
//     as a space-separated list; Encode joins the slice with a single
//     space, and the parser splits on whitespace and drops empties.
//   - RequestedTokenType is optional. RFC 8693 §2.1 SHOULD treat an
//     omitted value as a request for the same type as
//     SubjectTokenType; the library preserves the wire shape (omitted
//     = "") and leaves the default to the AS implementer. See
//     [TokenExchangeRequest.RequestedOrSubjectTokenType].
//
// Extra captures form parameters this library does not define. It
// round-trips on Encode so a spec extension that adds a new request
// parameter passes through unchanged; built-in fields take precedence
// when a key in Extra collides with a field name.
//
// Naming: the TokenExchange prefix mirrors the spec name (RFC 8693)
// and is load-bearing in the AS-side helper signatures
// (ParseTokenExchangeRequest, WriteTokenExchangeResponse,
// WriteTokenExchangeError) where the protocol identity disambiguates
// "request" / "response" / "error" at the call site.
//
//nolint:revive // spec-mandated TokenExchange* names; see paragraph above
type TokenExchangeRequest struct {
	// GrantType is the OAuth 2.0 grant-type URI. For a token-exchange
	// request it is [GrantTypeTokenExchange]; any other value is
	// rejected by Validate.
	GrantType string

	// Resource is the URI or URIs identifying the target resource
	// the issued token will be used against, per RFC 8707 and
	// RFC 8693 §2.1. May appear zero or more times on the wire.
	Resource []string

	// Audience is the logical name or names of the target service.
	// May appear zero or more times on the wire (RFC 8693 §2.1).
	Audience []string

	// Scope is the requested scope, modeled as a slice of individual
	// scope values. RFC 6749 §3.3 transports scope as a single
	// space-separated parameter; the codec joins on encode and splits
	// on parse.
	Scope []string

	// RequestedTokenType is the token-type URI naming the type the
	// caller would like back. Optional; when empty, RFC 8693 §2.1
	// SHOULD treat the request as asking for the same type as
	// SubjectTokenType.
	RequestedTokenType string

	// SubjectToken is the security token being exchanged. Required;
	// Validate rejects an empty value.
	SubjectToken string

	// SubjectTokenType is the token-type URI naming the type of
	// SubjectToken. Required; Validate rejects an empty value.
	SubjectTokenType string

	// ActorToken is the security token identifying the acting party
	// in a delegation. Optional, but if set, ActorTokenType MUST also
	// be set.
	ActorToken string

	// ActorTokenType is the token-type URI naming the type of
	// ActorToken. Optional, but if set, ActorToken MUST also be set.
	ActorTokenType string

	// Extra holds form parameters this library does not define. It
	// round-trips on Encode. Keys colliding with built-in fields are
	// shadowed by the built-in value; the codec emits built-in fields
	// last so they win on the wire as well.
	Extra url.Values
}

// RequestedOrSubjectTokenType returns RequestedTokenType when set, or
// SubjectTokenType otherwise. It implements the RFC 8693 §2.1 SHOULD
// for AS implementers that want one-call access to the effective
// requested token type without having to encode the default in their
// own policy layer.
func (r *TokenExchangeRequest) RequestedOrSubjectTokenType() string {
	if r.RequestedTokenType != "" {
		return r.RequestedTokenType
	}
	return r.SubjectTokenType
}
