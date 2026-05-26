// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

// resetRegistry restores the registry to its post-init state. Called
// by tests that register extension URIs to prevent cross-test
// pollution. The reset acquires the same lock the production code
// uses.
func resetRegistry(t *testing.T) {
	t.Helper()
	tokenTypeRegistry.mu.Lock()
	defer tokenTypeRegistry.mu.Unlock()
	tokenTypeRegistry.entries = initialRegistry()
}

func TestRegistryRecognizesBuiltinsAtInit(t *testing.T) {
	t.Parallel()

	for _, uri := range BuiltinTokenTypes() {
		if !IsRegisteredTokenType(uri) {
			t.Errorf("IsRegisteredTokenType(%q) = false, want true (built-in)", uri)
		}
	}
}

func TestRegistryRejectsBuiltinReRegistration(t *testing.T) {
	t.Parallel()

	for _, uri := range BuiltinTokenTypes() {
		err := RegisterTokenType(uri)
		if !errors.Is(err, ErrTokenTypeReserved) {
			t.Errorf("RegisterTokenType(%q) = %v, want ErrTokenTypeReserved", uri, err)
		}
	}
}

func TestRegisterTokenTypeAddsExtension(t *testing.T) {
	// Mutates the global registry; not parallel with other registry-
	// mutating tests.
	t.Cleanup(func() { resetRegistry(t) })

	const uri = "urn:example:custom-token-type"
	if IsRegisteredTokenType(uri) {
		t.Fatalf("registry already contained test URI before RegisterTokenType")
	}
	if err := RegisterTokenType(uri); err != nil {
		t.Fatalf("RegisterTokenType(%q): %v", uri, err)
	}
	if !IsRegisteredTokenType(uri) {
		t.Errorf("IsRegisteredTokenType(%q) = false after register", uri)
	}
}

func TestRegisterTokenTypeIdempotent(t *testing.T) {
	t.Cleanup(func() { resetRegistry(t) })

	const uri = "urn:example:idempotent-token-type"
	if err := RegisterTokenType(uri); err != nil {
		t.Fatalf("first RegisterTokenType: %v", err)
	}
	if err := RegisterTokenType(uri); err != nil {
		t.Errorf("second RegisterTokenType returned %v, want nil (idempotent)", err)
	}
}

func TestRegisterTokenTypeRejectsBadURI(t *testing.T) {
	t.Parallel()

	cases := []string{"", "not-a-uri", "/path/only", "#fragment"}
	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			t.Parallel()
			err := RegisterTokenType(in)
			if err == nil {
				t.Fatalf("RegisterTokenType(%q) returned nil", in)
			}
			if !errors.Is(err, &ValidationError{Reason: "not a valid URI"}) {
				t.Errorf("err = %v, want ValidationError 'not a valid URI'", err)
			}
		})
	}
}

func TestIsRegisteredTokenTypeRejectsEmpty(t *testing.T) {
	t.Parallel()

	if IsRegisteredTokenType("") {
		t.Errorf("IsRegisteredTokenType(\"\") = true")
	}
}

func TestIsRegisteredTokenTypeUnknownReturnsFalse(t *testing.T) {
	t.Parallel()

	if IsRegisteredTokenType("urn:example:never-registered") {
		t.Errorf("unknown URI reported as registered")
	}
}

// TestRegistryConcurrentReadsAndWrites stresses the RWMutex —
// readers must not block each other and writers must serialize
// without corrupting the map. The race detector verifies safety;
// the test merely produces enough contention to exercise the lock.
func TestRegistryConcurrentReadsAndWrites(t *testing.T) {
	t.Cleanup(func() { resetRegistry(t) })

	const (
		writers      = 4
		readers      = 16
		perWriter    = 50
		readsPerLoop = 100
	)

	var wg sync.WaitGroup

	// Writers: each registers a distinct slice of URIs.
	for w := range writers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := range perWriter {
				_ = RegisterTokenType(fmt.Sprintf("urn:example:writer-%d-%d", id, i))
			}
		}(w)
	}

	// Readers: walk the built-ins and a few synthesized misses.
	builtins := BuiltinTokenTypes()
	for r := range readers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := range readsPerLoop {
				_ = IsRegisteredTokenType(builtins[i%len(builtins)])
				_ = IsRegisteredTokenType(fmt.Sprintf("urn:example:reader-miss-%d-%d", id, i))
			}
		}(r)
	}

	wg.Wait()

	// After the storm, every built-in is still recognized.
	for _, uri := range builtins {
		if !IsRegisteredTokenType(uri) {
			t.Errorf("built-in %q lost from registry after concurrent storm", uri)
		}
	}

	// At least the last writer-N-(perWriter-1) extension should be
	// registered for each writer.
	for w := range writers {
		uri := fmt.Sprintf("urn:example:writer-%d-%d", w, perWriter-1)
		if !IsRegisteredTokenType(uri) {
			t.Errorf("extension %q missing after concurrent storm", uri)
		}
	}
}
