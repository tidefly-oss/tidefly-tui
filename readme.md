# Tidefly TUI

> Interactive setup wizard for [Tidefly](https://github.com/tidefly-oss/tidefly-backend) — built with Bubble Tea.

This repository contains the terminal UI installer that guides you through setting up Tidefly on your server. It configures Docker/Podman, generates secrets, sets up Traefik, SMTP, and writes the final `.env` file.

## Stack

- **Go** + [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Lip Gloss** — terminal styling

## Repositories

| Repo                                                                  | Description                         |
|-----------------------------------------------------------------------|-------------------------------------|
| [tidefly-backend](https://github.com/tidefly-oss/tidefly-backend)     | Go API + deployment engine          |
| [tidefly-ui](https://github.com/tidefly-oss/tidefly-ui)               | SvelteKit frontend                  |
| [tidefly-tui](https://github.com/tidefly-oss/tidefly-tui)             | This repo — Bubble Tea setup wizard |
| [tidefly-templates](https://github.com/tidefly-oss/tidefly-templates) | Service deploy templates            |
| [tidefly-docs](https://github.com/tidefly-oss/tidefly-docs)           | Documentation                       |

## Usage

Download the latest binary from [Releases](https://github.com/tidefly-oss/tidefly-tui/releases):

```bash
chmod +x tidefly-tui-linux-amd64
./tidefly-tui-linux-amd64
```

Or run from source:

```bash
git clone https://github.com/tidefly-oss/tidefly-tui
cd tidefly-tui
go run ./cmd/tidefly-installer
```

## Wizard Flow

```
Home → Runtime → Environment → Dashboard → Traefik → SMTP → Extras → Start → Admin → Done
```

## Contributing

See [CONTRIBUTING.md](contributing.md) for setup instructions and guidelines.

## License

[MIT](LICENSE)