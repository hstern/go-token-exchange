// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

// Conformance tests live in the external tokenexchange_test package
// so the test corpus can import both the library and the
// internal/specfixtures package — the latter pulls the library
// itself, which would be an import cycle if the tests were in
// package tokenexchange.
package tokenexchange_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"testing"

	tokenexchange "github.com/hstern/go-token-exchange"
	"github.com/hstern/go-token-exchange/internal/specfixtures"
)

// TestRequestConformance walks the request fixtures and asserts the
// three-step round-trip the design calls out:
//
//  1. Parse(spec wire) must equal Want.
//  2. Want.Validate() must accept.
//  3. Encode(Want) → re-parse → Parse(re-encoded) must equal Want.
//
// The byte form is NOT compared directly to the spec wire, because
// the spec wire uses spec-figure ordering and net/url.Values.Encode
// alphabetizes — that's the "modulo form-parameter ordering
// canonicalization" the issue note calls out.
func TestRequestConformance(t *testing.T) {
	t.Parallel()

	for _, fx := range specfixtures.Requests() {
		t.Run(fx.Name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(
				http.MethodPost,
				"/token",
				bytes.NewReader(fx.Wire),
			)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			parsed, err := tokenexchange.ParseTokenExchangeRequest(req)
			if err != nil {
				t.Fatalf("Parse(spec wire): %v", err)
			}
			if !reflect.DeepEqual(parsed, fx.Want) {
				t.Errorf("Parse(spec wire) = %+v\nwant                 %+v", parsed, fx.Want)
			}

			if err := fx.Want.Validate(); err != nil {
				t.Errorf("Want.Validate(): %v", err)
			}

			encoded := fx.Want.Encode()
			second := httptest.NewRequest(
				http.MethodPost,
				"/token",
				strings.NewReader(encoded.Encode()),
			)
			second.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			reparsed, err := tokenexchange.ParseTokenExchangeRequest(second)
			if err != nil {
				t.Fatalf("Parse(re-encoded): %v", err)
			}
			if !reflect.DeepEqual(reparsed, fx.Want) {
				t.Errorf("re-encode → re-parse = %+v\nwant               %+v", reparsed, fx.Want)
			}
		})
	}
}

// TestResponseConformance walks the response fixtures and asserts
// the JSON-side round-trip: Unmarshal(spec wire) must equal Want;
// Want.Validate must accept; Marshal(Want) → Unmarshal must equal
// Want with whitespace canonicalization via json.Compact.
func TestResponseConformance(t *testing.T) {
	t.Parallel()

	for _, fx := range specfixtures.Responses() {
		t.Run(fx.Name, func(t *testing.T) {
			t.Parallel()

			var parsed tokenexchange.TokenExchangeResponse
			if err := json.Unmarshal(fx.Wire, &parsed); err != nil {
				t.Fatalf("Unmarshal(spec wire): %v", err)
			}
			if !reflect.DeepEqual(&parsed, fx.Want) {
				t.Errorf("Unmarshal(spec wire) = %+v\nwant                 %+v", &parsed, fx.Want)
			}

			if err := fx.Want.Validate(); err != nil {
				t.Errorf("Want.Validate(): %v", err)
			}

			out, err := json.Marshal(fx.Want)
			if err != nil {
				t.Fatalf("Marshal(Want): %v", err)
			}

			// Byte-stability assertion: the canonicalized spec wire
			// (whitespace stripped) must contain every byte of the
			// marshaled output and vice versa, modulo JSON object
			// key order. Canonicalize both via Unmarshal → Marshal so
			// the comparison is order-insensitive.
			canonSpec, err := canonicalizeJSON(fx.Wire)
			if err != nil {
				t.Fatalf("canonicalize spec wire: %v", err)
			}
			canonOut, err := canonicalizeJSON(out)
			if err != nil {
				t.Fatalf("canonicalize marshal output: %v", err)
			}
			if !bytes.Equal(canonSpec, canonOut) {
				t.Errorf("byte-stability mismatch:\n  spec:    %s\n  marshal: %s", canonSpec, canonOut)
			}

			var second tokenexchange.TokenExchangeResponse
			if err := json.Unmarshal(out, &second); err != nil {
				t.Fatalf("Unmarshal(re-encoded): %v", err)
			}
			if !reflect.DeepEqual(&second, fx.Want) {
				t.Errorf("re-encode → re-parse = %+v\nwant               %+v", &second, fx.Want)
			}
		})
	}
}

// TestErrorConformance walks the error fixtures and asserts the
// JSON round-trip plus the byte-stability invariant.
func TestErrorConformance(t *testing.T) {
	t.Parallel()

	for _, fx := range specfixtures.Errors() {
		t.Run(fx.Name, func(t *testing.T) {
			t.Parallel()

			var parsed tokenexchange.TokenExchangeError
			if err := json.Unmarshal(fx.Wire, &parsed); err != nil {
				t.Fatalf("Unmarshal(spec wire): %v", err)
			}
			if !reflect.DeepEqual(&parsed, fx.Want) {
				t.Errorf("Unmarshal(spec wire) = %+v\nwant                 %+v", &parsed, fx.Want)
			}

			out, err := json.Marshal(fx.Want)
			if err != nil {
				t.Fatalf("Marshal(Want): %v", err)
			}

			canonSpec, err := canonicalizeJSON(fx.Wire)
			if err != nil {
				t.Fatalf("canonicalize spec wire: %v", err)
			}
			canonOut, err := canonicalizeJSON(out)
			if err != nil {
				t.Fatalf("canonicalize marshal output: %v", err)
			}
			if !bytes.Equal(canonSpec, canonOut) {
				t.Errorf("byte-stability mismatch:\n  spec:    %s\n  marshal: %s", canonSpec, canonOut)
			}
		})
	}
}

// TestRequestEncodeKeySetMatchesParseSet checks that Encode and Parse
// agree on which form parameters appear for a given fixture. Different
// orderings are allowed; missing or extra keys are not.
func TestRequestEncodeKeySetMatchesParseSet(t *testing.T) {
	t.Parallel()

	for _, fx := range specfixtures.Requests() {
		t.Run(fx.Name, func(t *testing.T) {
			t.Parallel()

			parsedFromSpec, err := url.ParseQuery(string(fx.Wire))
			if err != nil {
				t.Fatalf("ParseQuery(spec wire): %v", err)
			}
			encoded := fx.Want.Encode()

			specKeys := keySet(parsedFromSpec)
			encodedKeys := keySet(encoded)

			// Encode emits the required spec fields unconditionally
			// (grant_type, subject_token, subject_token_type). The spec
			// wire happens to include all three, so the sets must
			// match exactly for the bundled fixtures.
			slices.Sort(specKeys)
			slices.Sort(encodedKeys)
			if !slices.Equal(specKeys, encodedKeys) {
				t.Errorf("key set mismatch:\n  spec:    %v\n  encoded: %v", specKeys, encodedKeys)
			}
		})
	}
}

// canonicalizeJSON decodes b into a generic any and re-encodes it.
// The result is byte-stable across whitespace differences and object
// key ordering (encoding/json sorts map keys alphabetically), giving
// the conformance assertions a robust equality test.
func canonicalizeJSON(b []byte) ([]byte, error) {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

// keySet returns the keys of v as a freshly allocated slice; used by
// the EncodeKeySetMatchesParseSet check to compare without aliasing.
func keySet(v url.Values) []string {
	out := make([]string, 0, len(v))
	for k := range v {
		out = append(out, k)
	}
	return out
}
