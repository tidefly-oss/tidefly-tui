# Changelog

All notable changes to Tidefly TUI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

#### Caddy Migration
- Traefik page replaced with Caddy page (`caddy.go`)
- `SetupConfig` fields `TraefikEnabled/Domain/Email/Staging` → `CaddyEnabled/Domain/Email/Staging`
- `PageTraefik` → `PageCaddy` in navigation constants
- `start.go` env vars updated: `TRAEFIK_*` → `CADDY_*`
- `done.go` URLs updated to use Caddy domain
- `dashboard.go` navigates to `PageCaddy` instead of `PageTraefik`
- `main.go` `mergeConfig` updated for Caddy fields

#### Auth
- `admin.go` password hashing migrated from bcrypt to Argon2id
- Argon2id implementation inlined (TUI is a separate Go module from backend)
- Hash format matches backend exactly — user created via TUI can log in immediately

#### Wizard Flow
- Full rewrite: Home → Runtime → Environment → Dashboard → Caddy → SMTP → Extras → Start → Admin → Done
- `installer` package — Docker and Podman auto-install via official scripts
- Runtime page: shows installed/not-installed status, auto-installs if missing
- Caddy page: multi-step (toggle → domain → ACME email → staging/production CA)
- All config accumulated in `SetupConfig` across pages, written to `.env` in Writing step
- `start.go` service label updated to "Caddy, Postgres, Redis, Mailpit"

---

## [0.0.1-alpha] - TBD

> First internal alpha. Interactive Bubble Tea setup wizard for Tidefly.

### Added
- Initial Bubble Tea TUI setup wizard
- Docker and Podman runtime detection and configuration
- Secret generation and `.env` writing

---

[Unreleased]: https://github.com/tidefly-oss/tidefly-tui/compare/v0.0.1-alpha...HEAD
[0.0.1-alpha]: https://github.com/tidefly-oss/tidefly-tui/releases/tag/v0.0.1-alpha