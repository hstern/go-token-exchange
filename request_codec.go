// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"net/url"
	"strings"
)

// RFC 8693 §2.1 wire parameter names. Exported as constants so AS-
// side helpers and callers building custom Extra entries can name the
// reserved keys by symbol rather than re-typing the strings.
const (
	// ParamGrantType is the OAuth 2.0 grant_type form parameter.
	ParamGrantType = "grant_type"
	// ParamResource is the RFC 8707 resource form parameter; multi-
	// valued.
	ParamResource = "resource"
	// ParamAudience is the RFC 8693 §2.1 audience form parameter;
	// multi-valued.
	ParamAudience = "audience"
	// ParamScope is the RFC 6749 §3.3 scope form parameter; one
	// space-separated string value.
	ParamScope = "scope"
	// ParamRequestedTokenType is the RFC 8693 §2.1
	// requested_token_type form parameter.
	ParamRequestedTokenType = "requested_token_type"
	// ParamSubjectToken is the RFC 8693 §2.1 subject_token form
	// parameter.
	ParamSubjectToken = "subject_token"
	// ParamSubjectTokenType is the RFC 8693 §2.1 subject_token_type
	// form parameter.
	ParamSubjectTokenType = "subject_token_type"
	// ParamActorToken is the RFC 8693 §2.1 actor_token form
	// parameter.
	ParamActorToken = "actor_token"
	// ParamActorTokenType is the RFC 8693 §2.1 actor_token_type form
	// parameter.
	ParamActorTokenType = "actor_token_type"
)

// builtinRequestParams is the set of form parameter names this
// package defines for TokenExchangeRequest. Used to decide which
// Extra entries to drop on Encode (built-ins win) and which to
// capture on Parse (anything not in this set).
var builtinRequestParams = map[string]struct{}{
	ParamGrantType:          {},
	ParamResource:           {},
	ParamAudience:           {},
	ParamScope:              {},
	ParamRequestedTokenType: {},
	ParamSubjectToken:       {},
	ParamSubjectTokenType:   {},
	ParamActorToken:         {},
	ParamActorTokenType:     {},
}

// Encode returns the form-encoded body of r per RFC 8693 §2.1.
//
// Emission rules:
//
//   - GrantType, SubjectToken, SubjectTokenType are always emitted
//     (even when empty) so a Parse / Encode round-trip preserves
//     "present but empty" shapes. Validate, not Encode, enforces
//     the §2.1 MUSTs.
//   - Resource emits one resource form parameter per slice element,
//     in slice order. An empty or nil slice emits nothing.
//   - Audience emits one audience form parameter per slice element,
//     in slice order. An empty or nil slice emits nothing.
//   - Scope is space-joined (RFC 6749 §3.3) into a single scope
//     value. An empty or nil slice emits nothing; a slice whose only
//     element is the empty string emits an empty scope value.
//   - RequestedTokenType, ActorToken, ActorTokenType emit only when
//     non-empty.
//   - Extra contributions are added first; the built-in fields
//     overwrite or extend collisions, so a built-in key in Extra
//     cannot smuggle a competing value onto the wire.
//
// The returned url.Values is freshly allocated; the caller may
// mutate it (e.g. to add transport headers carried as form fields).
// Marshaling to bytes is the caller's responsibility — typically
// via the returned Values.Encode().
func (r *TokenExchangeRequest) Encode() url.Values {
	v := url.Values{}

	// Extra first; built-in keys are excluded so built-ins always win.
	for k, vals := range r.Extra {
		if _, builtin := builtinRequestParams[k]; builtin {
			continue
		}
		for _, val := range vals {
			v.Add(k, val)
		}
	}

	// Required-by-spec parameters: emit unconditionally so the
	// wire shape and the Go struct stay in lockstep through Parse /
	// Encode. Validate catches empty values.
	v.Set(ParamGrantType, r.GrantType)
	v.Set(ParamSubjectToken, r.SubjectToken)
	v.Set(ParamSubjectTokenType, r.SubjectTokenType)

	// Optional parameters: emit only when present.
	for _, res := range r.Resource {
		v.Add(ParamResource, res)
	}
	for _, aud := range r.Audience {
		v.Add(ParamAudience, aud)
	}
	if len(r.Scope) > 0 {
		v.Set(ParamScope, strings.Join(r.Scope, " "))
	}
	if r.RequestedTokenType != "" {
		v.Set(ParamRequestedTokenType, r.RequestedTokenType)
	}
	if r.ActorToken != "" {
		v.Set(ParamActorToken, r.ActorToken)
	}
	if r.ActorTokenType != "" {
		v.Set(ParamActorTokenType, r.ActorTokenType)
	}

	return v
}
