// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"fmt"
	"sync"
)

// tokenTypeRegistry is the process-wide set of token-type URIs the
// library recognizes. Two membership tiers share a single dispatch
// table: the six RFC 8693 §3 built-ins, populated at init and
// flagged so [RegisterTokenType] cannot shadow them; and consumer-
// supplied extensions added at runtime.
//
// Reads dominate the workload (every parsed token-type URI dispatch
// goes through one IsRegisteredTokenType call); writes are rare and
// happen at consumer startup. A sync.RWMutex matches that profile.
var tokenTypeRegistry = struct {
	mu sync.RWMutex
	// entries[uri] is true when uri is built-in (cannot be
	// re-registered or overridden) and false when uri is an
	// extension registered via RegisterTokenType. A missing entry
	// means the URI is not recognized.
	entries map[string]bool
}{
	entries: initialRegistry(),
}

// initialRegistry returns a fresh map seeded with the RFC 8693 §3
// built-in URIs and is invoked once at package init. The function
// is unexported but factored out so it can be re-used by test
// resets where a sibling test mutates the global; the production
// registry is never reset.
func initialRegistry() map[string]bool {
	out := make(map[string]bool, 6)
	for _, uri := range BuiltinTokenTypes() {
		out[uri] = true // true == built-in
	}
	return out
}

// RegisterTokenType adds uri to the set of recognized token-type
// URIs so [IsRegisteredTokenType] returns true for it. It returns
// [ErrTokenTypeReserved] when uri collides with one of the six
// RFC 8693 §3 built-in URIs — the library refuses to let consumers
// shadow a spec-defined type.
//
// Re-registering a previously-registered extension URI is a no-op;
// the function returns nil on the second call so consumers can run
// initialization idempotently (e.g. inside a sync.Once or at every
// process restart).
//
// Empty or non-URI-shaped uri values are rejected with a wrapped
// [*ValidationError] so callers can match the failure shape
// uniformly:
//
//	if errors.Is(err, &ValidationError{Reason: "not a valid URI"}) { … }
//
// RegisterTokenType is safe for concurrent use; the registry is
// guarded by a sync.RWMutex.
func RegisterTokenType(uri string) error {
	if !isValidTokenTypeURI(uri) {
		return &ValidationError{
			Rule:   "token-type registry",
			Reason: "not a valid URI",
		}
	}

	tokenTypeRegistry.mu.Lock()
	defer tokenTypeRegistry.mu.Unlock()

	if builtin, present := tokenTypeRegistry.entries[uri]; present && builtin {
		return fmt.Errorf("tokenexchange: register %q: %w", uri, ErrTokenTypeReserved)
	}

	tokenTypeRegistry.entries[uri] = false // false == extension
	return nil
}

// IsRegisteredTokenType reports whether uri is currently recognized
// by the registry — either as one of the six RFC 8693 §3 built-ins
// or as an extension added via [RegisterTokenType].
//
// Returning false does NOT mean uri is invalid; an unknown URI
// parses to [UnknownTokenType] so a request or response carrying
// it round-trips through the library. Strict consumers can compose
// IsRegisteredTokenType with explicit rejection ([ErrUnknownTokenType])
// at the policy layer.
//
// Safe for concurrent use.
func IsRegisteredTokenType(uri string) bool {
	if uri == "" {
		return false
	}
	tokenTypeRegistry.mu.RLock()
	defer tokenTypeRegistry.mu.RUnlock()
	_, ok := tokenTypeRegistry.entries[uri]
	return ok
}
