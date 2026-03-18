# Changelog

All notable changes to Tidefly TUI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Full rewrite of wizard flow: Home → Runtime → Environment → Dashboard → Traefik → SMTP → Extras → Start → Admin → Done
- `installer` package — Docker and Podman auto-install via official scripts
- Runtime page: shows installed/not-installed status, auto-installs if missing
- Traefik page: multi-step (toggle → domain → ACME email → staging/production CA)
- All config accumulated in `SetupConfig` across pages, written to `.env` in the Writing step

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