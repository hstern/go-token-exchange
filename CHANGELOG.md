# Changelog

All notable changes to `go-token-exchange` are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

The library's version is independent of the spec version it implements.
The implemented spec version is exposed as the `SpecVersion` constant.

## [Unreleased]

## [0.1.0] - 2026-05-26

Initial release. Implements RFC 8693 — OAuth 2.0 Token Exchange.

### Added

- Typed `TokenExchangeRequest`, `TokenExchangeResponse`, and
  `TokenExchangeError` structures covering the RFC 8693 §2.1
  request, §2.2 success response, and §2.4 / RFC 6749 §5.2 error
  response wire shapes.
- Form-encoded request codec: `(*TokenExchangeRequest).Encode()`
  and `ParseTokenExchangeRequest(*http.Request)`, with `Extra`
  capture for forward-compatibility on unknown parameters.
- JSON response codec with custom `MarshalJSON` / `UnmarshalJSON`
  that round-trip unknown JSON members through `Extra`.
- `Validate()` methods on request and response enforcing every
  RFC 8693 §2.1 and §2.2 MUST. Failures surface as typed
  `*ValidationError` with rule citation, wire parameter, and
  reason.
- Six built-in token-type URI constants from RFC 8693 §3
  (`TokenTypeAccessToken`, `TokenTypeRefreshToken`,
  `TokenTypeIDToken`, `TokenTypeSAML1`, `TokenTypeSAML2`,
  `TokenTypeJWT`), the `GrantTypeTokenExchange` grant URI, and an
  `UnknownTokenType` wrapper for unrecognized URIs.
- Token-type registry with `RegisterTokenType(uri) error` and
  `IsRegisteredTokenType(uri) bool`. Built-in URIs are reserved;
  collision returns `ErrTokenTypeReserved`.
- `Client` interface with single `Exchange(ctx, req)` method,
  unexported HTTP implementation returned by
  `NewClient(endpoint, opts...)`. Functional option
  `WithHTTPClient` is the seam for transport-layer client
  authentication (HTTP Basic, mTLS, private-key-JWT, etc.).
- AS-side helpers `ParseTokenExchangeRequest`,
  `WriteTokenExchangeResponse`, and `WriteTokenExchangeError` —
  free functions taking `net/http` primitives, no framework
  glue. `WriteTokenExchangeError` enforces the RFC 6749 §5.2
  status mapping (`invalid_client` → 401, everything else →
  400) at the library boundary.
- Seven RFC 6749 §5.2 / RFC 8693 §2.4 error code constants
  (`ErrCodeInvalidRequest`, `ErrCodeInvalidClient`,
  `ErrCodeInvalidGrant`, `ErrCodeUnauthorizedClient`,
  `ErrCodeUnsupportedGrantType`, `ErrCodeInvalidScope`,
  `ErrCodeInvalidTarget`).
- `(*TokenExchangeError).WithCause(error)` and `Unwrap()` for
  attaching transport / decode causes to a typed protocol error
  without leaking them to the wire.
- Conformance corpus in `internal/specfixtures/` embedding the
  RFC 8693 §2.1 / §2.2 / §2.4 example payloads; conformance,
  forward-compatibility, and registry-extension tests assert
  round-trip stability.

### Notes

- Module path is `github.com/hstern/go-token-exchange` (no `/v1`
  suffix per Go SemVer for the v0 / v1 series).
- Go 1.26 minimum.
- Zero non-test dependencies.
- Client authentication, `act` / `may_act` JWT claim handling,
  and RFC 8414 Authorization Server Metadata discovery are
  deliberately out of scope; see the README "What this library
  is and is not" section.

[0.1.0]: https://github.com/hstern/go-token-exchange/releases/tag/v0.1.0
