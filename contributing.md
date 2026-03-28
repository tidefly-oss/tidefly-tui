# Contributing to Tidefly TUI

Thanks for your interest in contributing!

## Getting Started

### Prerequisites

- Go 1.23+

### Setup

```bash
git clone https://github.com/tidefly-oss/tidefly-tui
cd tidefly-tui
go run ./cmd/tidefly-installer
```

## Development Workflow

```bash
go run ./cmd/tidefly-installer   # run wizard
go test ./...                    # run tests
go build ./...                   # build
golangci-lint run ./...          # lint
```

## Project Structure

```
cmd/tidefly-installer/    entry point
internal/
  installer/            wizard pages and setup logic
  version/              build version info
```

## Pull Requests

- Branch from `develop`, not `main`
- Keep PRs focused — one feature or fix per PR
- Update `changelog.md` under `[Unreleased]`

## Reporting Security Issues

Please do **not** open a public issue for security vulnerabilities.
Use [GitHub Private Security Advisories](https://github.com/tidefly-oss/tidefly-tui/security/advisories/new) instead.