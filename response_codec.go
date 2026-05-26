// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"encoding/json"
	"fmt"
)

// RFC 8693 §2.2 JSON object member names.
const (
	// FieldAccessToken is the access_token JSON member; carries the
	// issued security token regardless of its actual type (see
	// TokenExchangeResponse.AccessToken).
	FieldAccessToken = "access_token"
	// FieldIssuedTokenType is the issued_token_type JSON member; the
	// token-type URI for access_token.
	FieldIssuedTokenType = "issued_token_type"
	// FieldTokenType is the token_type JSON member (RFC 6749 §7.1).
	FieldTokenType = "token_type"
	// FieldExpiresIn is the expires_in JSON member; lifetime of
	// access_token in seconds.
	FieldExpiresIn = "expires_in"
	// FieldScope is the scope JSON member; space-separated per
	// RFC 6749 §3.3.
	FieldScope = "scope"
	// FieldRefreshToken is the refresh_token JSON member; optional
	// refresh token issued alongside the exchange.
	FieldRefreshToken = "refresh_token"
)

// builtinResponseFields is the set of JSON member names this package
// defines on TokenExchangeResponse. Used by MarshalJSON to drop
// colliding Extra entries (built-ins win) and by UnmarshalJSON to
// decide which members are captured into Extra.
var builtinResponseFields = map[string]struct{}{
	FieldAccessToken:     {},
	FieldIssuedTokenType: {},
	FieldTokenType:       {},
	FieldExpiresIn:       {},
	FieldScope:           {},
	FieldRefreshToken:    {},
}

// MarshalJSON implements [encoding/json.Marshaler] so the Extra map
// round-trips into the JSON object.
//
// Built-in fields are emitted with their json-tag semantics from
// TokenExchangeResponse (omitempty on ExpiresIn, Scope, and
// RefreshToken). Extra entries are merged into the object after the
// built-ins, but any Extra key whose name collides with a built-in
// field name is dropped — even when the built-in was omitted by
// omitempty. The drop rule keeps Extra from backfilling a built-in
// (e.g. an Extra "scope" entry cannot survive when r.Scope is
// empty); to override a built-in, set the typed field.
func (r TokenExchangeResponse) MarshalJSON() ([]byte, error) {
	type alias TokenExchangeResponse
	base, err := json.Marshal(alias(r))
	if err != nil {
		return nil, fmt.Errorf("tokenexchange: marshal response: %w", err)
	}

	if len(r.Extra) == 0 {
		return base, nil
	}

	var merged map[string]json.RawMessage
	if err := json.Unmarshal(base, &merged); err != nil {
		return nil, fmt.Errorf("tokenexchange: marshal response: re-decode failed: %w", err)
	}
	if merged == nil {
		merged = make(map[string]json.RawMessage, len(r.Extra))
	}

	for k, v := range r.Extra {
		if _, builtin := builtinResponseFields[k]; builtin {
			continue
		}
		merged[k] = v
	}

	out, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("tokenexchange: marshal response: re-encode failed: %w", err)
	}
	return out, nil
}

// UnmarshalJSON implements [encoding/json.Unmarshaler] so JSON members
// the library does not define are captured into r.Extra rather than
// silently dropped.
//
// Built-in member names decode into their typed fields with the same
// semantics as the default codec. Every other member name (including
// any future RFC 8693 extension) is placed in r.Extra verbatim, so a
// subsequent MarshalJSON re-emits it byte-stably.
//
// Decoding "null" leaves r at its zero value; decoding a non-object
// returns an error.
func (r *TokenExchangeResponse) UnmarshalJSON(data []byte) error {
	var members map[string]json.RawMessage
	if err := json.Unmarshal(data, &members); err != nil {
		return fmt.Errorf("tokenexchange: unmarshal response: %w", err)
	}

	type alias TokenExchangeResponse
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return fmt.Errorf("tokenexchange: unmarshal response: %w", err)
	}
	*r = TokenExchangeResponse(a)

	var extra map[string]json.RawMessage
	for k, v := range members {
		if _, builtin := builtinResponseFields[k]; builtin {
			continue
		}
		if extra == nil {
			extra = make(map[string]json.RawMessage, len(members))
		}
		extra[k] = v
	}
	r.Extra = extra

	return nil
}
