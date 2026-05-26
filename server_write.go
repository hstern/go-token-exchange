// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// WriteTokenExchangeResponse writes resp as the success body of an
// RFC 8693 §2.2 token-exchange response.
//
// The function:
//
//  1. Calls resp.Validate. A validation failure returns the
//     [*ValidationError] without writing anything to w, so the AS
//     handler can route the failure through its usual error path
//     (typically [WriteTokenExchangeError] with
//     [ErrCodeInvalidRequest] or a server-class 5xx).
//  2. Marshals resp via the custom JSON codec (which round-trips
//     the Extra map per TKEX-14).
//  3. Sets the RFC 6749 §5.1 response headers — Content-Type:
//     application/json, Cache-Control: no-store, Pragma: no-cache.
//  4. Writes HTTP 200.
//  5. Writes the marshaled body.
//
// A nil w returns a wrapped error rather than panicking. A nil resp
// is rejected by resp.Validate before any headers are touched.
//
// The function returns a non-nil error from any of the three failure
// modes (validation, marshal, body write). It does not partially
// write a response; once the headers go out, the caller assumes a
// best-effort write of the body (the underlying http.ResponseWriter
// may have already buffered the headers).
func WriteTokenExchangeResponse(w http.ResponseWriter, resp *TokenExchangeResponse) error {
	if w == nil {
		return fmt.Errorf("tokenexchange: write response: nil http.ResponseWriter")
	}
	if err := resp.Validate(); err != nil {
		return fmt.Errorf("tokenexchange: write response: %w", err)
	}

	body, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("tokenexchange: write response: marshal: %w", err)
	}

	setTokenEndpointHeaders(w.Header())
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(body); err != nil {
		return fmt.Errorf("tokenexchange: write response: %w", err)
	}
	return nil
}

// WriteTokenExchangeError writes e as the body of an RFC 6749 §5.2
// token-endpoint error response (extended by RFC 8693 §2.4).
//
// The HTTP status follows the RFC 6749 §5.2 mapping: invalid_client
// returns 401 so the AS can challenge with WWW-Authenticate; every
// other code returns 400. The status is derived from e.Code rather
// than supplied by the caller, so the spec mapping is enforced at
// the library boundary rather than discoverable only via the spec
// text.
//
// The function:
//
//  1. Validates that e is non-nil and e.Code is non-empty. An empty
//     Code is a wire shape the AS must not emit; the function
//     returns an error rather than write malformed bytes.
//  2. Marshals e through encoding/json (the TokenExchangeError
//     custom Error/Is methods are not on the marshal path).
//  3. Sets the RFC 6749 §5.2 response headers — Content-Type:
//     application/json, Cache-Control: no-store, Pragma: no-cache.
//  4. Writes the mapped status code.
//  5. Writes the marshaled body.
//
// Note: a TokenExchangeError carrying a cause via WithCause writes
// the same wire bytes as one without — cause is a Go-side chain
// artifact and never crosses the wire.
func WriteTokenExchangeError(w http.ResponseWriter, e *TokenExchangeError) error {
	if w == nil {
		return fmt.Errorf("tokenexchange: write error: nil http.ResponseWriter")
	}
	if e == nil {
		return fmt.Errorf("tokenexchange: write error: nil *TokenExchangeError")
	}
	if e.Code == "" {
		return fmt.Errorf("tokenexchange: write error: empty Code")
	}

	body, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("tokenexchange: write error: marshal: %w", err)
	}

	setTokenEndpointHeaders(w.Header())
	w.WriteHeader(httpStatusForErrorCode(e.Code))
	if _, err := w.Write(body); err != nil {
		return fmt.Errorf("tokenexchange: write error: %w", err)
	}
	return nil
}

// setTokenEndpointHeaders sets the headers RFC 6749 §5.1 / §5.2
// require on token endpoint responses. Pragma is included for
// HTTP/1.0 caches that ignore Cache-Control.
func setTokenEndpointHeaders(h http.Header) {
	h.Set("Content-Type", "application/json")
	h.Set("Cache-Control", "no-store")
	h.Set("Pragma", "no-cache")
}

// httpStatusForErrorCode maps an RFC 6749 §5.2 error code to its
// HTTP status. The mapping is one rule: invalid_client → 401,
// everything else → 400. Unknown codes (e.g. from a downstream
// profile) default to 400, which is the safe interpretation for
// any client-class error.
func httpStatusForErrorCode(code string) int {
	if code == ErrCodeInvalidClient {
		return http.StatusUnauthorized
	}
	return http.StatusBadRequest
}
