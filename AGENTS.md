# AGENTS.md

Conventions for contributors and code-generation agents working in this
repository.

## Project shape

`go-token-exchange` is a Go library implementing RFC 8693 OAuth 2.0
Token Exchange. It is a standalone, library-vendor-neutral
implementation. It depends on the Go standard library only — no
third-party runtime dependencies in `v0.x`. Test dependencies are
similarly restricted unless a concrete reason warrants one.

Minimum Go version: 1.26.

## Code style

- `gofmt`-formatted. No exceptions.
- `go vet ./...` clean.
- `golangci-lint run ./...` clean against the repo's `.golangci.yml`.
- Public API is documented with godoc on every exported symbol. Doc
  comments cite the relevant RFC 8693 (or RFC 6749) section for
  wire-shape decisions.

## Copyright header

Every Go source file — including test files — starts with exactly
this two-line header, before the `package` declaration:

```go
// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0
```

No additional preamble; the full license text lives in `LICENSE`.

## Commit messages

Detailed by default. The minimum acceptable message is one imperative
title (≤72 chars, no trailing period) plus one body paragraph
explaining why the change exists. Title in imperative present tense
("add", not "added"). Body wrapped at ~72 columns. Cite spec section
numbers and public RFC references for wire-shape changes.

Single-line commits are reserved for trivial changes — typos,
dependency-version bumps with no API impact, scaffold commits during
repo bootstrap.

## Wire fidelity

The library is a standards implementation. Where a choice exists
between wire fidelity and ergonomic shortcuts, fidelity wins:

- Request body is `application/x-www-form-urlencoded` (RFC 6749 §3.2
  inherited convention). The codec is `net/url.Values`, not a struct
  + `form:` tags.
- Response body is JSON, decoded with `encoding/json`.
- Lenient unmarshal, strict marshal. Parsing populates the typed
  struct and stops; validation happens explicitly at `Validate()` or
  marshal time.
- Open-extension fields (`Extra`) round-trip verbatim. Unknown
  parameters and unknown token-type URIs do not cause parse errors.

## Tests

Standard library `testing` package, table-driven where natural.
Run with `go test -race -shuffle=on ./...`. Wire-shape behavior is
exercised against the embedded RFC 8693 example payloads in
`internal/specfixtures/` — every spec example round-trips.

## Branch and release flow

`main` is protected: PR-driven, required CI green, branch up to date
with `main` before merge. Direct push is disabled.

Releases are tagged `vMAJOR.MINOR.PATCH` on `main`. Major-version
bumps follow the branch-per-major convention (`v2` lives on a `v2`
branch with `module .../v2` in `go.mod`); no in-tree versioned
subdirectories.

## Reporting

Bugs and feature requests: file via the public issue tracker linked
from the repository page on the hosting platform.
