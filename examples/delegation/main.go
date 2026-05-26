// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

// Delegation example: a service exchanges its own access token
// alongside the original subject's token (the "actor_token" pair) to
// obtain a token usable on the subject's behalf for a specific
// audience. The AS sees both the subject the action is for AND the
// actor doing the acting — the foundation for on-behalf-of flows.
//
// The exchange shape mirrors RFC 8693 §2.1 Figure 3: the request
// carries grant_type, audience, subject_token, subject_token_type,
// actor_token, and actor_token_type. Both actor fields must be
// present together — the validator enforces the pairing.
//
// Run with `go run ./examples/delegation`.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	tokenexchange "github.com/hstern/go-token-exchange"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("delegation example: %v", err)
	}
}

func run() error {
	srv := httptest.NewServer(http.HandlerFunc(authServerHandler))
	defer srv.Close()

	client := tokenexchange.NewClient(srv.URL)
	req := &tokenexchange.TokenExchangeRequest{
		GrantType:        tokenexchange.GrantTypeTokenExchange,
		Audience:         []string{"urn:example:cooperation-context"},
		SubjectToken:     "subject-jwt-asserting-the-user",
		SubjectTokenType: tokenexchange.TokenTypeIDToken,
		ActorToken:       "actor-access-token-of-acting-service",
		ActorTokenType:   tokenexchange.TokenTypeAccessToken,
	}

	resp, err := client.Exchange(context.Background(), req)
	if err != nil {
		return fmt.Errorf("exchange: %w", err)
	}

	fmt.Println("issued access_token:     ", resp.AccessToken)
	fmt.Println("issued issued_token_type:", resp.IssuedTokenType)
	fmt.Println("issued token_type:       ", resp.TokenType)
	fmt.Println("issued expires_in:       ", resp.ExpiresIn)
	return nil
}

// authServerHandler is a stub /token handler that parses the
// inbound delegation request, validates it (the validator enforces
// the actor_token / actor_token_type pairing), and issues a
// synthetic token whose body names both the subject and the
// audience so the demo output shows the delegation took effect.
func authServerHandler(w http.ResponseWriter, r *http.Request) {
	req, err := tokenexchange.ParseTokenExchangeRequest(r)
	if err != nil {
		writeError(w, tokenexchange.ErrCodeInvalidRequest, err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		code := tokenexchange.ErrCodeInvalidRequest
		if errors.Is(err, tokenexchange.ErrInvalidGrantType) {
			code = tokenexchange.ErrCodeUnsupportedGrantType
		}
		writeError(w, code, err.Error())
		return
	}

	audience := ""
	if len(req.Audience) > 0 {
		audience = req.Audience[0]
	}
	resp := &tokenexchange.TokenExchangeResponse{
		AccessToken:     "delegation-to:" + audience + ";acting-for:subject",
		IssuedTokenType: tokenexchange.TokenTypeAccessToken,
		TokenType:       "Bearer",
		ExpiresIn:       60,
	}
	if err := tokenexchange.WriteTokenExchangeResponse(w, resp); err != nil {
		log.Printf("write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, code, description string) {
	err := tokenexchange.WriteTokenExchangeError(w, &tokenexchange.TokenExchangeError{
		Code:        code,
		Description: description,
	})
	if err != nil {
		log.Printf("write error response: %v", err)
	}
}
