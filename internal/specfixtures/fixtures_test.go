// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package specfixtures

import "testing"

// TestFixturesLoad verifies the embedded payloads are reachable and
// non-empty. It is the smoke test that fails first if a fixture file
// is renamed or removed without updating the catalog.
func TestFixturesLoad(t *testing.T) {
	t.Parallel()

	t.Run("requests", func(t *testing.T) {
		t.Parallel()
		got := Requests()
		if len(got) == 0 {
			t.Fatalf("Requests() returned empty catalog")
		}
		for _, f := range got {
			if f.Name == "" {
				t.Errorf("request fixture with empty Name")
			}
			if len(f.Wire) == 0 {
				t.Errorf("request fixture %q has empty Wire", f.Name)
			}
			if f.Want == nil {
				t.Errorf("request fixture %q has nil Want", f.Name)
			}
		}
	})

	t.Run("responses", func(t *testing.T) {
		t.Parallel()
		got := Responses()
		if len(got) == 0 {
			t.Fatalf("Responses() returned empty catalog")
		}
		for _, f := range got {
			if f.Name == "" {
				t.Errorf("response fixture with empty Name")
			}
			if len(f.Wire) == 0 {
				t.Errorf("response fixture %q has empty Wire", f.Name)
			}
			if f.Want == nil {
				t.Errorf("response fixture %q has nil Want", f.Name)
			}
		}
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()
		got := Errors()
		if len(got) == 0 {
			t.Fatalf("Errors() returned empty catalog")
		}
		for _, f := range got {
			if f.Name == "" {
				t.Errorf("error fixture with empty Name")
			}
			if len(f.Wire) == 0 {
				t.Errorf("error fixture %q has empty Wire", f.Name)
			}
			if f.Want == nil {
				t.Errorf("error fixture %q has nil Want", f.Name)
			}
		}
	})
}

// TestFixtureNamesUnique ensures no two fixtures within a catalog
// share a name; subtest discrimination depends on the invariant.
func TestFixtureNamesUnique(t *testing.T) {
	t.Parallel()

	check := func(t *testing.T, names []string) {
		t.Helper()
		seen := make(map[string]int)
		for i, n := range names {
			if prev, ok := seen[n]; ok {
				t.Errorf("duplicate name %q at indices %d and %d", n, prev, i)
			}
			seen[n] = i
		}
	}

	t.Run("requests", func(t *testing.T) {
		t.Parallel()
		fixtures := Requests()
		names := make([]string, 0, len(fixtures))
		for _, f := range fixtures {
			names = append(names, f.Name)
		}
		check(t, names)
	})

	t.Run("responses", func(t *testing.T) {
		t.Parallel()
		fixtures := Responses()
		names := make([]string, 0, len(fixtures))
		for _, f := range fixtures {
			names = append(names, f.Name)
		}
		check(t, names)
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()
		fixtures := Errors()
		names := make([]string, 0, len(fixtures))
		for _, f := range fixtures {
			names = append(names, f.Name)
		}
		check(t, names)
	})
}
