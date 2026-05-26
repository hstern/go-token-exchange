// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"errors"
	"fmt"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
	}{
		{"ErrTokenTypeReserved", ErrTokenTypeReserved},
		{"ErrUnknownTokenType", ErrUnknownTokenType},
		{"ErrInvalidGrantType", ErrInvalidGrantType},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.err == nil {
				t.Fatalf("%s is nil", tc.name)
			}
			if tc.err.Error() == "" {
				t.Errorf("%s.Error() is empty", tc.name)
			}
		})
	}
}

func TestSentinelErrorIsItself(t *testing.T) {
	t.Parallel()

	wrapped := fmt.Errorf("registry: %w", ErrTokenTypeReserved)
	if !errors.Is(wrapped, ErrTokenTypeReserved) {
		t.Errorf("errors.Is(wrapped, ErrTokenTypeReserved) = false, want true")
	}
	if errors.Is(wrapped, ErrInvalidGrantType) {
		t.Errorf("errors.Is(reserved-wrap, ErrInvalidGrantType) = true, want false")
	}
}

func TestValidationErrorString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  *ValidationError
		want string
	}{
		{
			name: "nil receiver",
			err:  nil,
			want: "tokenexchange: <nil validation error>",
		},
		{
			name: "with parameter",
			err: &ValidationError{
				Rule:      "RFC 8693 §2.1",
				Parameter: "subject_token",
				Reason:    "missing",
			},
			want: "tokenexchange: RFC 8693 §2.1: subject_token: missing",
		},
		{
			name: "structural (no parameter)",
			err: &ValidationError{
				Rule:   "RFC 8693 §2.1",
				Reason: "actor_token and actor_token_type must be paired",
			},
			want: "tokenexchange: RFC 8693 §2.1: actor_token and actor_token_type must be paired",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.err.Error(); got != tc.want {
				t.Errorf("Error() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestValidationErrorIs(t *testing.T) {
	t.Parallel()

	got := &ValidationError{
		Rule:      "RFC 8693 §2.1",
		Parameter: "subject_token",
		Reason:    "missing",
	}

	t.Run("matches all fields exactly", func(t *testing.T) {
		t.Parallel()
		exact := &ValidationError{
			Rule:      "RFC 8693 §2.1",
			Parameter: "subject_token",
			Reason:    "missing",
		}
		if !errors.Is(got, exact) {
			t.Errorf("errors.Is(exact match) = false, want true")
		}
	})

	t.Run("wildcards on Parameter only", func(t *testing.T) {
		t.Parallel()
		wild := &ValidationError{Parameter: "subject_token"}
		if !errors.Is(got, wild) {
			t.Errorf("errors.Is(Parameter wildcard) = false, want true")
		}
	})

	t.Run("wildcards on everything matches any ValidationError", func(t *testing.T) {
		t.Parallel()
		anyVE := &ValidationError{}
		if !errors.Is(got, anyVE) {
			t.Errorf("errors.Is(empty target) = false, want true")
		}
	})

	t.Run("non-match on Parameter", func(t *testing.T) {
		t.Parallel()
		other := &ValidationError{Parameter: "actor_token"}
		if errors.Is(got, other) {
			t.Errorf("errors.Is(actor_token) = true, want false")
		}
	})

	t.Run("non-match across types", func(t *testing.T) {
		t.Parallel()
		if errors.Is(got, errors.New("missing")) {
			t.Errorf("errors.Is(plain error) = true, want false")
		}
	})

	t.Run("wrapped still matches", func(t *testing.T) {
		t.Parallel()
		wrapped := fmt.Errorf("Validate: %w", got)
		if !errors.Is(wrapped, &ValidationError{Parameter: "subject_token"}) {
			t.Errorf("errors.Is(wrapped, parameter wildcard) = false, want true")
		}
	})
}
