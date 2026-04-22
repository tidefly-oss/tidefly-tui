# Changelog

All notable changes to Tidefly TUI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

---

## [0.0.1-alpha.1] - 2026-03-31

> First public alpha. Interactive Bubble Tea setup wizard for Tidefly.

### Added
- Interactive setup wizard — Home → Runtime → Environment → Dashboard → Caddy → SMTP → Start → Admin → Done
- Docker and Podman runtime detection — auto-installs if not found
- Secret generation via `init-env.sh` — all secrets generated on first run
- Environment config writing — runtime, Caddy, SMTP vars patched into `.env`
- Docker network setup — `tidefly_proxy` and `tidefly_internal` created automatically
- Cleanup step — removes orphaned containers before each setup run
- Rollback on failure — `docker compose down` called automatically if any step fails
- Caddy configuration — enable/skip, configure domain now or later in the UI
- SMTP configuration — optional, supports None/STARTTLS/TLS
- Admin account creation — Argon2id hashed, written directly to Postgres
- curl install script — `scripts/install.sh` for one-line server installation
- Module path migrated to `github.com/tidefly-oss/tidefly-tui`

---

## Roadmap

### Next (Beta)
- [ ] Update command — pull latest images and restart services
- [ ] Uninstall command
- [ ] Agent node setup wizard

---

[Unreleased]: https://github.com/tidefly-oss/tidefly-tui/compare/v0.0.1-alpha.1...HEAD
[0.0.1-alpha.1]: https://github.com/tidefly-oss/tidefly-tui/releases/tag/v0.0.1-alpha.1


<div align="center">
  <sub>Built with ❤️ by <a href="https://github.com/dbuettgen">@dbuettgen</a> · Part of the <a href="https://github.com/tidefly-oss">tidefly-oss</a> project</sub>
</div>