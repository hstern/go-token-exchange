// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
)

// Client is the client-side surface of the OAuth 2.0 Token Exchange
// grant defined in RFC 8693. A Client performs one method —
// Exchange — against a single token endpoint.
//
// The interface shape lets callers wrap or mock without depending
// on the library's concrete implementation. Tests can substitute a
// stub Client; downstream packages can decorate the default with
// retry, telemetry, or client-authentication transport. The
// concrete implementation returned by [NewClient] is intentionally
// unexported so the only stable surface is this interface plus the
// constructors that build it.
//
// Client authentication is out of scope for v0.1. RFC 8693 §2.1
// inherits the RFC 6749 §2.3.1 client-authentication rules; the
// library lets the caller bring an *http.Client whose Transport
// applies whatever flavor the AS requires (HTTP Basic, mTLS,
// private-key-JWT, etc.) rather than bundling the auth flavor
// explosion.
type Client interface {
	// Exchange performs a single RFC 8693 §2.1 token-exchange
	// request and returns the parsed §2.2 success response on
	// HTTP 2xx, a typed [*TokenExchangeError] on a JSON 4xx, or
	// a wrapped transport error otherwise.
	Exchange(ctx context.Context, req *TokenExchangeRequest) (*TokenExchangeResponse, error)
}

// NewClient returns a [Client] that talks to tokenEndpoint using
// [http.DefaultClient]. Use [WithHTTPClient] (or whichever option
// the consumer ships) to inject a configured *http.Client carrying
// the client authentication the AS requires.
//
// tokenEndpoint must be the absolute URL of the AS token endpoint
// (e.g. https://as.example.com/token). The constructor does not
// validate the URL beyond being non-empty; downstream callers that
// want strict validation should do that at configuration time.
func NewClient(tokenEndpoint string) Client {
	return &httpClient{
		tokenEndpoint: tokenEndpoint,
		httpClient:    http.DefaultClient,
	}
}

// httpClient is the default Client implementation: an HTTP POST to
// tokenEndpoint with the form-encoded request body and a JSON
// response. Unexported because the Client interface is the stable
// surface.
type httpClient struct {
	tokenEndpoint string
	httpClient    *http.Client
}

// Exchange performs the RFC 8693 §2.1 token-exchange grant against
// the token endpoint configured at construction time.
//
// The function validates req via (*TokenExchangeRequest).Validate
// before any network activity, so a malformed request never reaches
// the AS. On success it returns the typed response with body bytes
// fully consumed; on a JSON-shaped 4xx it returns a
// [*TokenExchangeError] (with the original transport response body
// optionally available via the error's Unwrap chain when a decode
// failure also occurs); on any other HTTP status it returns a
// wrapped transport error naming the status.
//
// The context's Cancel propagates to the underlying HTTP request via
// http.NewRequestWithContext.
func (c *httpClient) Exchange(ctx context.Context, req *TokenExchangeRequest) (*TokenExchangeResponse, error) {
	if c == nil || c.tokenEndpoint == "" {
		return nil, fmt.Errorf("tokenexchange: client: token endpoint not configured")
	}
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("tokenexchange: exchange: %w", err)
	}

	body := req.Encode().Encode()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenEndpoint, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("tokenexchange: exchange: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	hc := c.httpClient
	if hc == nil {
		hc = http.DefaultClient
	}
	resp, err := hc.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("tokenexchange: exchange: %w", err)
	}
	defer func() {
		//nolint:errcheck // body is fully drained below; a Close error after
		// the response is already in hand is not actionable.
		resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("tokenexchange: exchange: read response body: %w", err)
	}

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		var out TokenExchangeResponse
		if err := json.Unmarshal(respBody, &out); err != nil {
			return nil, fmt.Errorf("tokenexchange: exchange: decode response (HTTP %d): %w", resp.StatusCode, err)
		}
		return &out, nil

	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		if !isJSONContentType(resp.Header.Get("Content-Type")) {
			return nil, fmt.Errorf("tokenexchange: exchange: HTTP %d %s with non-JSON body (%d bytes)",
				resp.StatusCode, http.StatusText(resp.StatusCode), len(respBody))
		}
		var te TokenExchangeError
		if err := json.Unmarshal(respBody, &te); err != nil {
			return nil, fmt.Errorf("tokenexchange: exchange: decode error response (HTTP %d): %w", resp.StatusCode, err)
		}
		if te.Code == "" {
			return nil, fmt.Errorf("tokenexchange: exchange: HTTP %d JSON body without error code", resp.StatusCode)
		}
		return nil, &te

	default:
		return nil, fmt.Errorf("tokenexchange: exchange: HTTP %d %s",
			resp.StatusCode, http.StatusText(resp.StatusCode))
	}
}

// isJSONContentType reports whether v parses as a media type whose
// primary type/subtype is application/json (ignoring parameters
// like charset). Empty Content-Type returns false.
func isJSONContentType(v string) bool {
	if v == "" {
		return false
	}
	mt, _, err := mime.ParseMediaType(v)
	if err != nil {
		return false
	}
	return mt == "application/json"
}
