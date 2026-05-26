// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import "net/http"

// Option configures a [Client] at construction time. It is the
// functional-options shape used by [NewClient]: each Option is a
// small closure that mutates the unexported implementation struct
// before NewClient hands the wrapped value back to the caller as a
// Client interface.
//
// The Option type is intentionally opaque (its parameter is an
// unexported type) so the set of valid options is closed: only the
// With* helpers in this package can produce one. New options can be
// added in later releases without breaking callers — the variadic
// signature on NewClient and the closed Option set together form a
// forward-compatible surface.
type Option func(*httpClient)

// WithHTTPClient returns an Option that overrides the http.Client
// the [Client] uses for the underlying POST. The injected client is
// where transport-layer client authentication lives: a wrapping
// http.RoundTripper that adds HTTP Basic, an mTLS-configured
// Transport, a private-key-JWT signing wrapper, and so on are all
// applied at the http.Client / http.Transport layer rather than as
// library options.
//
// Passing nil restores the default ([http.DefaultClient]). This is
// the explicit-reset shape: a caller writing
// NewClient(endpoint, WithHTTPClient(nil)) gets the documented
// default rather than a silently broken client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *httpClient) {
		if hc == nil {
			c.httpClient = http.DefaultClient
			return
		}
		c.httpClient = hc
	}
}
