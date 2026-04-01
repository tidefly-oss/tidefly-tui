# Tidefly TUI

> Interactive setup wizard for [Tidefly](https://github.com/tidefly-oss/tidefly-plane) — built with Bubble Tea.

This repository contains the terminal UI installer that guides you through setting up Tidefly on your server. It detects Docker/Podman, generates secrets, configures Caddy, SMTP, and starts all services automatically.

## Installation
```bash
curl -fsSL https://raw.githubusercontent.com/tidefly-oss/tidefly-tui/main/scripts/install.sh | bash
tidefly-tui
```

## Stack

- **Go 1.26** + [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Lip Gloss** — terminal styling

## Repositories

| Repo                                                                  | Description                                            |
|-----------------------------------------------------------------------|--------------------------------------------------------|
| [tidefly-plane](https://github.com/tidefly-oss/tidefly-plane)         | Go backend — API, deployment engine, worker management |
| [tidefly-agent](https://github.com/tidefly-oss/tidefly-agent)         | Worker agent — runs on remote nodes, connects via mTLS |
| [tidefly-ui](https://github.com/tidefly-oss/tidefly-ui)               | SvelteKit frontend dashboard                           |
| [tidefly-tui](https://github.com/tidefly-oss/tidefly-tui)             | This repo — Bubble Tea setup wizard                    |
| [tidefly-templates](https://github.com/tidefly-oss/tidefly-templates) | Service deploy templates                               |
| [tidefly-docs](https://github.com/tidefly-oss/tidefly-docs)           | Documentation (coming soon)                            |

## Wizard Flow
```
Home → Runtime → Environment → Dashboard → Caddy → SMTP → Start → Admin → Done
```

## Run from Source
```bash
git clone https://github.com/tidefly-oss/tidefly-tui
cd tidefly-tui
go run ./cmd/tidefly-installer
```

## Contributing

See [contributing.md](contributing.md) for setup instructions and guidelines.

## Security

Please do **not** open public issues for security vulnerabilities — use [GitHub Private Security Advisories](https://github.com/tidefly-oss/tidefly-tui/security/advisories/new) instead.

---

<div align="center">

Built with ❤️ ·[AGPLv3](https://github.com/tidefly-oss/tidefly-tui/blob/main/LICENSE) · [Report a vulnerability](https://github.com/tidefly-oss/tidefly-tui/security/advisories/new)

</div>