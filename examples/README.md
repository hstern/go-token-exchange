# Examples

Self-contained runnable examples of `go-token-exchange`. Each spins
up a stub authorization server in-process via `httptest.NewServer`,
performs the exchange, and prints the result — no external services
needed.

```
go run ./examples/impersonation
go run ./examples/delegation
```

| Example | Wire shape | Demonstrates |
|---|---|---|
| `impersonation/` | RFC 8693 §2.1 Figure 1 | client + AS for the simplest token-exchange pattern: client downscopes its own access token for a specific `resource`. |
| `delegation/` | RFC 8693 §2.1 Figure 3 | client + AS for on-behalf-of: paired `actor_token` / `actor_token_type` identifies the acting service while `subject_token` carries the subject's identity. |

Both examples cover the full library round-trip: client construction
via `NewClient`, request build, `Exchange`, plus the AS-side handler
shape (`ParseTokenExchangeRequest` → `Validate` → application
issuance → `WriteTokenExchangeResponse`). The handler treats token
issuance as a trivial string concatenation so the spec wire shape is
the focus, not JWT minting.

A real authorization server would replace the stub `authServerHandler`
with: client authentication (RFC 6749 §2.3.1, applied at the
transport / `http.Client` layer rather than as a library option),
policy evaluation, and actual token issuance (JWT signing, opaque
token allocation, or whatever the AS deploys).
