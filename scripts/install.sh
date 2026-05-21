#!/usr/bin/env bash
set -euo pipefail

REPO="tidefly-oss/tidefly-tui"
BINARY="tidefly-tui"

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RESET='\033[0m'

log_info()    { echo -e "${BLUE}[tidefly]${RESET} $*"; }
log_success() { echo -e "${GREEN}[tidefly]${RESET} $*"; }
log_error()   { echo -e "${RED}[tidefly]${RESET} $*" >&2; }

# ── Install dir — root gets /usr/local/bin, user gets ~/.local/bin ────────────
if [[ $EUID -eq 0 ]]; then
    INSTALL_DIR="/usr/local/bin"
else
    INSTALL_DIR="${HOME}/.local/bin"
    mkdir -p "$INSTALL_DIR"
fi

# ── Detect arch ───────────────────────────────────────────────────────────────
detect_arch() {
    local arch
    arch=$(uname -m)
    case "$arch" in
        x86_64)  echo "amd64" ;;
        aarch64) echo "arm64" ;;
        armv7l)  echo "arm" ;;
        *)
            log_error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac
}

# ── Detect OS ─────────────────────────────────────────────────────────────────
detect_os() {
    local os
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$os" in
        linux)  echo "linux" ;;
        darwin) echo "darwin" ;;
        *)
            log_error "Unsupported OS: $os"
            exit 1
            ;;
    esac
}

# ── Check deps ────────────────────────────────────────────────────────────────
check_deps() {
    for cmd in curl grep; do
        if ! command -v "$cmd" &>/dev/null; then
            log_error "Missing dependency: $cmd"
            exit 2
        fi
    done
}

# ── Get latest version ────────────────────────────────────────────────────────
get_latest_version() {
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name"' \
        | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/'
}

# ── PATH hint for non-root ────────────────────────────────────────────────────
print_path_hint() {
    if [[ ":$PATH:" == *":${INSTALL_DIR}:"* ]]; then
        return
    fi

    echo ""
    log_info "${INSTALL_DIR} is not in your PATH yet. Add it:"
    echo ""

    # Detect shell
    local shell_name
    shell_name=$(basename "${SHELL:-bash}")

    case "$shell_name" in
        fish)
            echo -e "  ${BLUE}fish_add_path ${INSTALL_DIR}${RESET}"
            ;;
        zsh)
            echo -e "  ${BLUE}echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.zshrc && source ~/.zshrc${RESET}"
            ;;
        *)
            echo -e "  ${BLUE}echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.bashrc && source ~/.bashrc${RESET}"
            ;;
    esac

    echo ""
}

# ── Install ───────────────────────────────────────────────────────────────────
main() {
    check_deps

    local os arch version
    os=$(detect_os)
    arch=$(detect_arch)
    version=$(get_latest_version)

    if [[ -z "$version" ]]; then
        log_error "Could not determine latest version"
        exit 1
    fi

    local url="https://github.com/${REPO}/releases/download/${version}/tidefly-tui-${os}-${arch}"
    local tmp
    tmp=$(mktemp)

    log_info "Installing Tidefly TUI ${version} (${os}/${arch})..."
    log_info "Downloading from ${url}..."

    if ! curl -fsSL "$url" -o "$tmp"; then
        log_error "Download failed"
        rm -f "$tmp"
        exit 1
    fi

    chmod +x "$tmp"
    mv "$tmp" "${INSTALL_DIR}/${BINARY}"

    log_success "Installed to ${INSTALL_DIR}/${BINARY}"

    if [[ $EUID -ne 0 ]]; then
        print_path_hint
    fi

    echo ""
    echo -e "${GREEN}  → tidefly-tui${RESET}"
    echo ""
}

main "$@"
