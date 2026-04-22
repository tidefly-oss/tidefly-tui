# Contributing to Tidefly TUI

Thanks for your interest in contributing!

## Getting Started

### Prerequisites

- Go 1.26+
- A running Docker or Podman instance for testing

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
  env/                  .env loading and path resolution
  installer/            Docker/Podman auto-install logic
  pages/                wizard pages (Home, Runtime, Environment, ...)
  styles/               Lip Gloss color and layout definitions
scripts/
  install.sh            curl-based installer for end users
```

## Pull Requests

- Branch from `develop`, not `main`
- Keep PRs focused — one feature or fix per PR
- Update `changelog.md` under `[Unreleased]`

## Reporting Security Issues

Please do **not** open a public issue for security vulnerabilities.
Use [GitHub Private Security Advisories](https://github.com/tidefly-oss/tidefly-tui/security/advisories/new) instead.

---

<div align="center">
  <sub>Built with ❤️ by <a href="https://github.com/dbuettgen">@dbuettgen</a> · Part of the <a href="https://github.com/tidefly-oss">tidefly-oss</a> project</sub>
</div>