# go-token-exchange

A Go library implementing [RFC 8693 — OAuth 2.0 Token
Exchange](https://www.rfc-editor.org/rfc/rfc8693.html).

**Status:** pre-publication. The first tagged release will be `v0.1.0`.
The API is subject to change before that tag.

```
import "github.com/hstern/go-token-exchange"
```

Targets Go 1.26 or newer. Zero non-test dependencies — standard library only.

## Scope

Implements the wire-level primitives for the OAuth 2.0 Token Exchange
grant defined in RFC 8693:

- Typed request and response structures.
- Form-encoded request body (per RFC 6749 §3.2 + RFC 8693 §2.1) and
  JSON response body (per RFC 8693 §2.2).
- Token-type URI constants for the six built-ins from RFC 8693 §3, plus
  a `RegisterTokenType` extension point for IANA registry growth.
- Client-side `Exchange` against a token endpoint.
- Authorization-server-side parse / validate / write helpers, usable
  with any `net/http`-compatible handler stack.

### Out of scope (for v0.1)

- Client authentication. The caller supplies an `*http.Client` carrying
  whatever transport-layer credentials the AS requires.
- `act` and `may_act` JWT claim handling. These are JWT-level concerns
  and belong in a JWT library, not here.
- OAuth 2.0 Authorization Server Metadata discovery (RFC 8414).

## License

Apache-2.0. See [LICENSE](LICENSE).
