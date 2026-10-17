# AGENTS.md

Guidance for coding agents working in this repository.

## Overview

This is the **Sourcegraph Accounts SDK for Go** — a Go library for integrating
with the Sourcegraph Accounts Management System (SAMS). The module path is
`github.com/sourcegraph/sourcegraph-accounts-sdk-go`. See `README.md` for
user-facing API examples.

## Setup

- Go version `1.23.4` (see `.tool-versions` and `go.mod`).
- Some dependencies live in private Sourcegraph repos. Configure private module
  access before building/testing:

```zsh
go env -w GOPRIVATE=github.com/sourcegraph/*
```

## Build, Test, Lint

```zsh
# Fetch dependencies / verify module tidiness (CI fails if `go mod tidy` changes anything)
go mod tidy

# Run tests (matches CI flags)
go test -shuffle=on -race ./...

# Lint (CI runs golangci-lint; config in .golangci.yml)
golangci-lint run --timeout=30m
```

## Code Generation

[Buf](https://buf.build) and [Connect](https://connectrpc.com/) generate the
gRPC / Protocol Buffers code. After editing any `.proto` files, run `buf generate`
from the directory containing `buf.gen.yaml`. Required tools:

```zsh
go install github.com/bufbuild/buf/cmd/buf@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
```

## Conventions

- Use `github.com/sourcegraph/sourcegraph/lib/errors` instead of the standard
  `errors` package — this is enforced by the `depguard` linter in `.golangci.yml`.
- No naked returns (`nakedret` with `max-func-lines: 0`).
- Formatting is enforced via `gofmt` and `goimports`.
- Generated protobuf/Connect code (`*.pb.go`, `*connect/`) should not be edited
  by hand — regenerate with `buf generate` instead.

## Where the code lives

- Top-level `*.go` files (`sams.go`, `clientv1*.go`, `accountsv1.go`,
  `notificationsv1.go`) — the main SDK entrypoints in package `sams`.
- `auth/` — SAMS user authentication flow (OIDC login/callback handlers).
- `clients/v1/` — Clients API v1 (generated Connect/protobuf code).
- `accounts/v1/` — Accounts API v1.
- `notifications/v1/` — Notifications API v1 (Pub/Sub subscriber).
- `scopes/`, `roles/`, `services/` — supporting domain types.
