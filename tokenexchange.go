// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

// Package tokenexchange implements the OAuth 2.0 Token Exchange grant
// defined in RFC 8693.
//
// The package surface is split between protocol primitives (typed
// request and response structures, the token-type URI registry, parse
// and validate helpers) and an HTTP client that performs an exchange
// against a token endpoint. The form-encoded request body follows
// RFC 6749 §3.2; the JSON response body follows RFC 8693 §2.2.
package tokenexchange

// SpecVersion identifies the version of the OAuth 2.0 Token Exchange
// specification this library implements.
const SpecVersion = "RFC 8693"
