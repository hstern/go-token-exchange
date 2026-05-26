// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

func TestErrorCodeConstants(t *testing.T) {
	t.Parallel()

	cases := []struct {
		got, want string
	}{
		{ErrCodeInvalidRequest, "invalid_request"},
		{ErrCodeInvalidClient, "invalid_client"},
		{ErrCodeInvalidGrant, "invalid_grant"},
		{ErrCodeUnauthorizedClient, "unauthorized_client"},
		{ErrCodeUnsupportedGrantType, "unsupported_grant_type"},
		{ErrCodeInvalidScope, "invalid_scope"},
		{ErrCodeInvalidTarget, "invalid_target"},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("got %q, want %q", tc.got, tc.want)
		}
	}
}

func TestTokenExchangeErrorErrorString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  *TokenExchangeError
		want string
	}{
		{
			name: "nil receiver",
			err:  nil,
			want: "tokenexchange: <nil error>",
		},
		{
			name: "code only",
			err:  &TokenExchangeError{Code: ErrCodeInvalidTarget},
			want: "tokenexchange: invalid_target",
		},
		{
			name: "code + description",
			err: &TokenExchangeError{
				Code:        ErrCodeInvalidGrant,
				Description: "subject_token expired",
			},
			want: "tokenexchange: invalid_grant: subject_token expired",
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

func TestTokenExchangeErrorIs(t *testing.T) {
	t.Parallel()

	got := &TokenExchangeError{
		Code:        ErrCodeInvalidTarget,
		Description: "audience https://other.example.com is not authorized",
	}
	sentinel := &TokenExchangeError{Code: ErrCodeInvalidTarget}

	if !errors.Is(got, sentinel) {
		t.Errorf("errors.Is(got, sentinel) = false, want true")
	}

	wrapped := fmt.Errorf("token exchange failed: %w", got)
	if !errors.Is(wrapped, sentinel) {
		t.Errorf("errors.Is(wrapped, sentinel) = false, want true")
	}

	other := &TokenExchangeError{Code: ErrCodeInvalidGrant}
	if errors.Is(got, other) {
		t.Errorf("errors.Is(invalid_target, invalid_grant) = true, want false")
	}

	if errors.Is(got, errors.New("invalid_target")) {
		t.Errorf("errors.Is matched a plain errors.New value; should only match *TokenExchangeError")
	}
}

func TestTokenExchangeErrorJSON(t *testing.T) {
	t.Parallel()

	t.Run("code only round-trips minimal", func(t *testing.T) {
		t.Parallel()
		in := &TokenExchangeError{Code: ErrCodeInvalidRequest}
		out, err := json.Marshal(in)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		want := `{"error":"invalid_request"}`
		if string(out) != want {
			t.Errorf("Marshal = %s, want %s", out, want)
		}

		var back TokenExchangeError
		if err := json.Unmarshal(out, &back); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if back.Code != in.Code || back.Description != "" || back.URI != "" {
			t.Errorf("round-trip mismatch: got %+v, want %+v", back, *in)
		}
	})

	t.Run("all fields populated", func(t *testing.T) {
		t.Parallel()
		const in = `{"error":"invalid_target","error_description":"bad audience","error_uri":"https://example.com/errors/invalid_target"}`
		var got TokenExchangeError
		if err := json.Unmarshal([]byte(in), &got); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		want := TokenExchangeError{
			Code:        ErrCodeInvalidTarget,
			Description: "bad audience",
			URI:         "https://example.com/errors/invalid_target",
		}
		if got != want {
			t.Errorf("Unmarshal = %+v, want %+v", got, want)
		}
	})
}
