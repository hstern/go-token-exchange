// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

// Impersonation example: a client exchanges its access token at an
// AS token endpoint to obtain a downscoped access token usable
// against a specific backend resource. The AS runs in-process via
// httptest.NewServer so the example is self-contained — run it with
// `go run ./examples/impersonation` and watch the round-trip.
//
// The exchange shape mirrors RFC 8693 §2.1 Figure 1: the request
// carries grant_type, resource, subject_token, and
// subject_token_type; the response carries the issued access_token,
// issued_token_type, token_type, and expires_in.
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
		log.Fatalf("impersonation example: %v", err)
	}
}

func run() error {
	srv := httptest.NewServer(http.HandlerFunc(authServerHandler))
	defer srv.Close()

	client := tokenexchange.NewClient(srv.URL)
	req := &tokenexchange.TokenExchangeRequest{
		GrantType:        tokenexchange.GrantTypeTokenExchange,
		Resource:         []string{"https://backend.example.com/api"},
		SubjectToken:     "callers-original-access-token",
		SubjectTokenType: tokenexchange.TokenTypeAccessToken,
	}

	resp, err := client.Exchange(context.Background(), req)
	if err != nil {
		return fmt.Errorf("exchange: %w", err)
	}

	fmt.Println("issued access_token:    ", resp.AccessToken)
	fmt.Println("issued issued_token_type:", resp.IssuedTokenType)
	fmt.Println("issued token_type:      ", resp.TokenType)
	fmt.Println("issued expires_in:      ", resp.ExpiresIn)
	return nil
}

// authServerHandler is a stub /token handler that parses the
// inbound request, validates it, issues a synthetic downscoped
// token for the requested resource, and writes the response. Real
// authorization servers would consult policy, mint a JWT, and
// involve client authentication; this handler keeps the moving
// parts to the parts the library covers.
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

	// "Issue" the downscoped token. In a real AS this is where the
	// JWT minting / opaque token allocation / policy decision lives.
	resource := ""
	if len(req.Resource) > 0 {
		resource = req.Resource[0]
	}
	resp := &tokenexchange.TokenExchangeResponse{
		AccessToken:     "downscoped-for:" + resource,
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
