// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

// Example functions in package tokenexchange_test render on
// pkg.go.dev as executable documentation. They live in the external
// test package so the imports they show in the rendered doc match
// what a consumer would write.
package tokenexchange_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"

	tokenexchange "github.com/hstern/go-token-exchange"
)

// ExampleNewClient constructs a default Client pointed at an AS
// token endpoint. The injected http.Client is the seam for
// transport-layer client authentication (HTTP Basic via a wrapping
// RoundTripper, mTLS via http.Transport, private-key-JWT via a
// signing wrapper).
func ExampleNewClient() {
	c := tokenexchange.NewClient(
		"https://as.example.com/token",
		tokenexchange.WithHTTPClient(http.DefaultClient),
	)
	_ = c // use c.Exchange(ctx, req) to perform a token exchange
}

// ExampleClient_Exchange performs a token-exchange round-trip
// against an in-process AS that returns a canned response.
func ExampleClient_Exchange() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"access_token": "issued-token",
			"issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
			"token_type": "Bearer",
			"expires_in": 300
		}`))
	}))
	defer srv.Close()

	c := tokenexchange.NewClient(srv.URL)
	resp, err := c.Exchange(context.Background(), &tokenexchange.TokenExchangeRequest{
		GrantType:        tokenexchange.GrantTypeTokenExchange,
		Resource:         []string{"https://backend.example.com/api"},
		SubjectToken:     "the-callers-access-token",
		SubjectTokenType: tokenexchange.TokenTypeAccessToken,
	})
	if err != nil {
		fmt.Println("exchange failed:", err)
		return
	}
	fmt.Println(resp.AccessToken, resp.TokenType, resp.ExpiresIn)
	// Output: issued-token Bearer 300
}

// ExampleParseTokenExchangeRequest parses an inbound form-encoded
// token-exchange request from an http.Request.
func ExampleParseTokenExchangeRequest() {
	body := `grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Atoken-exchange` +
		`&subject_token=incoming-token` +
		`&subject_token_type=urn%3Aietf%3Aparams%3Aoauth%3Atoken-type%3Aaccess_token` +
		`&resource=https%3A%2F%2Fbackend.example.com%2Fapi`
	r := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	req, err := tokenexchange.ParseTokenExchangeRequest(r)
	if err != nil {
		fmt.Println("parse failed:", err)
		return
	}
	fmt.Println("subject:", req.SubjectToken)
	fmt.Println("type:", req.SubjectTokenType)
	fmt.Println("resource:", req.Resource)
	// Output:
	// subject: incoming-token
	// type: urn:ietf:params:oauth:token-type:access_token
	// resource: [https://backend.example.com/api]
}

// ExampleWriteTokenExchangeResponse writes a success response with
// the RFC 6749 §5.1 headers from inside an AS handler.
func ExampleWriteTokenExchangeResponse() {
	rr := httptest.NewRecorder()
	resp := &tokenexchange.TokenExchangeResponse{
		AccessToken:     "newly-issued",
		IssuedTokenType: tokenexchange.TokenTypeAccessToken,
		TokenType:       "Bearer",
		ExpiresIn:       60,
	}
	if err := tokenexchange.WriteTokenExchangeResponse(rr, resp); err != nil {
		fmt.Println("write failed:", err)
		return
	}
	fmt.Println("status:", rr.Code)
	fmt.Println("content-type:", rr.Header().Get("Content-Type"))
	fmt.Println("cache-control:", rr.Header().Get("Cache-Control"))
	// Output:
	// status: 200
	// content-type: application/json
	// cache-control: no-store
}

// ExampleWriteTokenExchangeError shows the RFC 6749 §5.2 status
// mapping the AS-side helper enforces: invalid_client returns 401,
// every other code returns 400.
func ExampleWriteTokenExchangeError() {
	for _, code := range []string{
		tokenexchange.ErrCodeInvalidTarget,
		tokenexchange.ErrCodeInvalidClient,
	} {
		rr := httptest.NewRecorder()
		_ = tokenexchange.WriteTokenExchangeError(rr, &tokenexchange.TokenExchangeError{
			Code: code,
		})
		fmt.Printf("%s -> %d\n", code, rr.Code)
	}
	// Output:
	// invalid_target -> 400
	// invalid_client -> 401
}

// ExampleTokenExchangeRequest_Encode shows how a typed request
// turns into a form-encoded body. The form keys are alphabetized
// by net/url.Values.Encode, so output ordering is stable.
func ExampleTokenExchangeRequest_Encode() {
	req := &tokenexchange.TokenExchangeRequest{
		GrantType:        tokenexchange.GrantTypeTokenExchange,
		SubjectToken:     "tok",
		SubjectTokenType: tokenexchange.TokenTypeAccessToken,
		Scope:            []string{"read", "write"},
	}
	values := req.Encode()
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%s=%v\n", k, values[k])
	}
	// Output:
	// grant_type=[urn:ietf:params:oauth:grant-type:token-exchange]
	// scope=[read write]
	// subject_token=[tok]
	// subject_token_type=[urn:ietf:params:oauth:token-type:access_token]
}

// ExampleTokenExchangeRequest_Validate flags a missing required
// field with a typed *ValidationError naming the spec citation,
// the wire parameter, and a reason.
func ExampleTokenExchangeRequest_Validate() {
	req := &tokenexchange.TokenExchangeRequest{
		GrantType: tokenexchange.GrantTypeTokenExchange,
		// SubjectToken intentionally empty
		SubjectTokenType: tokenexchange.TokenTypeAccessToken,
	}
	err := req.Validate()
	fmt.Println(err)
	// Output: tokenexchange: RFC 8693 §2.1: subject_token: missing
}

// ExampleTokenExchangeError_Is matches a parsed error against a
// sentinel value by Code. The sentinel can be a freshly-constructed
// value or a package-level variable; both work with errors.Is.
func ExampleTokenExchangeError_Is() {
	var got error = &tokenexchange.TokenExchangeError{
		Code:        tokenexchange.ErrCodeInvalidTarget,
		Description: "audience https://other.example.com is not authorized",
	}
	if errors.Is(got, &tokenexchange.TokenExchangeError{Code: tokenexchange.ErrCodeInvalidTarget}) {
		fmt.Println("matched invalid_target")
	}
	// Output: matched invalid_target
}

// ExampleRegisterTokenType adds a downstream URI to the recognized
// set; IsRegisteredTokenType returns true for it on subsequent
// calls. Re-registration is idempotent.
func ExampleRegisterTokenType() {
	const uri = "urn:example:demo-token-type"
	if err := tokenexchange.RegisterTokenType(uri); err != nil {
		fmt.Println("register failed:", err)
		return
	}
	fmt.Println("registered:", tokenexchange.IsRegisteredTokenType(uri))
	// Output: registered: true
}

// ExampleUnknownTokenType wraps a URI the registry does not yet
// recognize. The wrapper exists at the Go-type boundary; the wire
// carries the URI string unchanged.
func ExampleUnknownTokenType() {
	u := tokenexchange.UnknownTokenType{URI: "urn:example:future:token-type"}
	fmt.Println(u)
	// Output: urn:example:future:token-type
}

// ExampleTokenExchangeResponse_MarshalJSON demonstrates that the
// custom codec round-trips the Extra map: unknown JSON members
// emit alongside the typed fields.
func ExampleTokenExchangeResponse_MarshalJSON() {
	resp := tokenexchange.TokenExchangeResponse{
		AccessToken:     "tok",
		IssuedTokenType: tokenexchange.TokenTypeAccessToken,
		TokenType:       "Bearer",
		Extra: map[string]json.RawMessage{
			"x-tenant": json.RawMessage(`"acme"`),
		},
	}
	out, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("marshal failed:", err)
		return
	}
	// json.Marshal of a map sorts keys alphabetically; the output
	// shape is deterministic.
	fmt.Println(string(out))
	// Output: {"access_token":"tok","issued_token_type":"urn:ietf:params:oauth:token-type:access_token","token_type":"Bearer","x-tenant":"acme"}
}
