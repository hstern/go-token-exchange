# go-token-exchange

A Go library implementing [RFC 8693 — OAuth 2.0 Token
Exchange](https://www.rfc-editor.org/rfc/rfc8693.html).

**Status:** pre-publication. The first tagged release will be `v0.1.0`.
The API is subject to change before that tag.

```
import "github.com/hstern/go-token-exchange"
```

Targets Go 1.26 or newer. Zero non-test dependencies — standard library only.

## What this library is and is not

It is a wire-fidelity implementation of the RFC 8693 §2.1 request
and §2.2 response — typed Go structs, a form-encoded codec for the
request body (the OAuth 2.0 token endpoint convention), a JSON
codec for the response, validation against the §2.1 and §2.2 MUSTs,
a small client built on `net/http`, and à-la-carte parse / write
helpers for authorization-server handlers.

It is not a full OAuth 2.0 stack. The following are deliberately
out of scope and belong in dedicated libraries:

- **Client authentication.** RFC 8693 inherits the RFC 6749 §2.3.1
  flavor explosion (HTTP Basic, client_secret in the form body,
  private-key-JWT, mTLS, etc.). The library takes an `*http.Client`
  whose `Transport` applies whatever flavor the AS requires.
- **`act` and `may_act` JWT claims** (RFC 8693 §4.1, §4.3). These
  describe the delegation chain inside an issued JWT; constructing
  or parsing them is JWT-shaped work for a JWT library.
- **Authorization Server Metadata discovery** (RFC 8414, OIDC
  discovery). Sibling library territory.

## Client quickstart

A token-exchange client posts a `TokenExchangeRequest` at an AS
token endpoint and gets back a typed `TokenExchangeResponse` on
success or a typed `*TokenExchangeError` on a protocol failure.

```go
package main

import (
    "context"
    "log"
    "net/http"

    tokenexchange "github.com/hstern/go-token-exchange"
)

func main() {
    // The injected http.Client is where transport-layer client
    // authentication lives — HTTP Basic via a wrapping
    // RoundTripper, mTLS via http.Transport, and so on.
    c := tokenexchange.NewClient(
        "https://as.example.com/token",
        tokenexchange.WithHTTPClient(http.DefaultClient),
    )

    req := &tokenexchange.TokenExchangeRequest{
        GrantType:        tokenexchange.GrantTypeTokenExchange,
        Resource:         []string{"https://backend.example.com/api"},
        SubjectToken:     "the-callers-access-token",
        SubjectTokenType: tokenexchange.TokenTypeAccessToken,
    }
    resp, err := c.Exchange(context.Background(), req)
    if err != nil {
        log.Fatalf("token exchange failed: %v", err)
    }
    log.Printf("issued %s (%d s)", resp.AccessToken, resp.ExpiresIn)
}
```

For delegation (on-behalf-of), set the paired `ActorToken` and
`ActorTokenType` fields; the validator enforces both-or-neither.

## Authorization-server quickstart

The AS-side surface is three free functions —
`ParseTokenExchangeRequest`, `WriteTokenExchangeResponse`, and
`WriteTokenExchangeError`. They take `net/http` primitives, so they
plug into any handler stack without framework glue.

```go
package main

import (
    "context"
    "errors"
    "log"
    "net/http"

    tokenexchange "github.com/hstern/go-token-exchange"
)

func tokenHandler(w http.ResponseWriter, r *http.Request) {
    req, err := tokenexchange.ParseTokenExchangeRequest(r)
    if err != nil {
        _ = tokenexchange.WriteTokenExchangeError(w, &tokenexchange.TokenExchangeError{
            Code:        tokenexchange.ErrCodeInvalidRequest,
            Description: err.Error(),
        })
        return
    }
    if err := req.Validate(); err != nil {
        code := tokenexchange.ErrCodeInvalidRequest
        if errors.Is(err, tokenexchange.ErrInvalidGrantType) {
            code = tokenexchange.ErrCodeUnsupportedGrantType
        }
        _ = tokenexchange.WriteTokenExchangeError(w, &tokenexchange.TokenExchangeError{
            Code:        code,
            Description: err.Error(),
        })
        return
    }

    // Application policy + token issuance happens here. Whatever the
    // AS's existing pipeline does for an OAuth 2.0 token endpoint.
    issued, err := issueToken(r.Context(), req)
    if err != nil {
        _ = tokenexchange.WriteTokenExchangeError(w, &tokenexchange.TokenExchangeError{
            Code: tokenexchange.ErrCodeInvalidGrant,
        })
        return
    }
    _ = tokenexchange.WriteTokenExchangeResponse(w, issued)
}

func issueToken(ctx context.Context, req *tokenexchange.TokenExchangeRequest) (*tokenexchange.TokenExchangeResponse, error) {
    // ... application-specific issuance ...
    return &tokenexchange.TokenExchangeResponse{
        AccessToken:     "issued-token",
        IssuedTokenType: tokenexchange.TokenTypeAccessToken,
        TokenType:       "Bearer",
        ExpiresIn:       300,
    }, nil
}

func main() {
    http.HandleFunc("/token", tokenHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

`WriteTokenExchangeError` applies the RFC 6749 §5.2 status mapping
at the library boundary: `invalid_client` returns 401, every other
code returns 400.

## Forward compatibility

Two extension axes let a future RFC 8693 profile pass through this
library without code changes:

- **Token-type URI registry.** The library ships the six RFC 8693
  §3 built-ins. Downstream consumers add their own with
  `RegisterTokenType(uri)`; the registry refuses to shadow a
  built-in. Unrecognized URIs decode without error — the library
  carries them as plain strings on the wire, and `UnknownTokenType`
  wraps them at the Go-type boundary.
- **`Extra` capture.** Form parameters and JSON members the library
  does not define round-trip through `TokenExchangeRequest.Extra`
  (a `url.Values`) and `TokenExchangeResponse.Extra` (a
  `map[string]json.RawMessage`). Built-in fields take precedence on
  collision.

## Spec version

```go
const SpecVersion = "RFC 8693"
```

Library SemVer is independent of the spec version. The first
release is `v0.1.0`; a future RFC superseding 8693 would be a
library-major decision, not a minor bump.

## License

Apache-2.0. See [LICENSE](LICENSE).
