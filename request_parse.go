// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ParseTokenExchangeRequest reads the form-encoded body of r and
// returns a typed [TokenExchangeRequest]. It is the inverse of
// [TokenExchangeRequest.Encode]: form parameters whose names are
// defined by RFC 8693 §2.1 populate the corresponding typed fields,
// and any other form parameters are captured into Extra so they
// round-trip on a subsequent Encode.
//
// The parser does NOT validate. It is lenient on receive per Postel's
// law: missing required fields decode to zero values, duplicate
// single-valued parameters take the first value (matching
// [url.Values.Get] semantics), and unknown values for the
// grant_type or token-type URIs decode without error. Apply
// (*TokenExchangeRequest).Validate() afterwards to enforce the §2.1
// MUSTs.
//
// ParseTokenExchangeRequest calls [http.Request.ParseForm], which
// requires the request's Content-Type to be
// application/x-www-form-urlencoded for POST/PUT/PATCH bodies. A
// parse error on the body is wrapped and returned. An empty body
// returns a non-nil request whose fields are all zero.
//
// The function does not consume r.Body any differently than
// ParseForm itself; if the caller has already invoked ParseForm
// (e.g. in a middleware), the call here is idempotent.
func ParseTokenExchangeRequest(r *http.Request) (*TokenExchangeRequest, error) {
	if r == nil {
		return nil, fmt.Errorf("tokenexchange: parse request: nil *http.Request")
	}
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("tokenexchange: parse request: %w", err)
	}

	out := &TokenExchangeRequest{
		GrantType:          r.Form.Get(ParamGrantType),
		SubjectToken:       r.Form.Get(ParamSubjectToken),
		SubjectTokenType:   r.Form.Get(ParamSubjectTokenType),
		RequestedTokenType: r.Form.Get(ParamRequestedTokenType),
		ActorToken:         r.Form.Get(ParamActorToken),
		ActorTokenType:     r.Form.Get(ParamActorTokenType),
	}

	if v := r.Form[ParamResource]; len(v) > 0 {
		out.Resource = append([]string(nil), v...)
	}
	if v := r.Form[ParamAudience]; len(v) > 0 {
		out.Audience = append([]string(nil), v...)
	}
	if raw := r.Form.Get(ParamScope); raw != "" {
		out.Scope = splitScope(raw)
	}

	var extra url.Values
	for k, vals := range r.Form {
		if _, builtin := builtinRequestParams[k]; builtin {
			continue
		}
		if extra == nil {
			extra = url.Values{}
		}
		// Defensive copy: r.Form may be mutated by the handler after
		// ParseTokenExchangeRequest returns, and the parsed request
		// should not alias that storage.
		extra[k] = append([]string(nil), vals...)
	}
	out.Extra = extra

	return out, nil
}

// splitScope splits a space-separated scope value per RFC 6749 §3.3
// and drops empty entries that result from consecutive separators or
// leading / trailing whitespace.
func splitScope(raw string) []string {
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return nil
	}
	return fields
}
