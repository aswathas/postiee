#!/usr/bin/env sh
# Postiee installer for macOS and Linux.
#
#   curl -fsSL https://raw.githubusercontent.com/aswathas/postiee/main/install.sh | sh
#
# Installs the prebuilt binary for your platform into ~/.local/bin (or
# /usr/local/bin when you have write access). No Go toolchain required.
set -eu

REPO="aswathas/postiee"
BIN="postiee"
INSTALL_DIR="${POSTIEE_INSTALL_DIR:-}"

log()  { printf '%s\n' "$*" >&2; }
die()  { printf 'error: %s\n' "$*" >&2; exit 1; }

need() { command -v "$1" >/dev/null 2>&1 || die "$1 is required but not found"; }

need uname
need tar

# ---- detect platform -------------------------------------------------------
os="$(uname -s)"
arch="$(uname -m)"

case "$os" in
  Darwin) os_label="Darwin" ;;
  Linux)  os_label="Linux" ;;
  *)      die "unsupported OS: $os. Use install.ps1 on Windows." ;;
esac

case "$arch" in
  x86_64|amd64)  arch_label="x86_64" ;;
  arm64|aarch64) arch_label="arm64" ;;
  *)             die "unsupported architecture: $arch" ;;
esac

# ---- pick install directory ------------------------------------------------
if [ -z "$INSTALL_DIR" ]; then
  if [ -w "/usr/local/bin" ] 2>/dev/null; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="$HOME/.local/bin"
  fi
fi

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) log ""; log "note: $INSTALL_DIR is not on your PATH."; log "      add it:  export PATH=\"\$HOME/.local/bin:\$PATH\"" ;;
esac

# ---- download --------------------------------------------------------------
tag="${POSTIEE_VERSION:-latest}"
base="https://github.com/$REPO/releases"

if [ "$tag" = "latest" ]; then
  url="$base/latest/download/${BIN}_${os_label}_${arch_label}.tar.gz"
  log "resolving latest release..."
else
  url="$base/download/$tag/${BIN}_${os_label}_${arch_label}.tar.gz"
fi

log "downloading $BIN for ${os_label}/${arch_label}..."
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT INT TERM

if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$url" -o "$tmp/pkg.tar.gz" || die "download failed: $url"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "$tmp/pkg.tar.gz" "$url" || die "download failed: $url"
else
  die "neither curl nor wget found"
fi

log "extracting..."
tar -xzf "$tmp/pkg.tar.gz" -C "$tmp" || die "could not extract archive"

if [ ! -f "$tmp/$BIN" ]; then
  die "archive did not contain a '$BIN' binary"
fi

# ---- install ---------------------------------------------------------------
mkdir -p "$INSTALL_DIR"
# Install atomically so a concurrent run cannot leave a half-written binary.
mv "$tmp/$BIN" "$INSTALL_DIR/.$BIN.tmp"
chmod +x "$INSTALL_DIR/.$BIN.tmp"
mv "$INSTALL_DIR/.$BIN.tmp" "$INSTALL_DIR/$BIN"

log ""
log "installed $("$INSTALL_DIR/$BIN" --version 2>/dev/null || echo "$BIN") -> $INSTALL_DIR/$BIN"
log ""
log "next:"
log "  1. create a LinkedIn app at https://www.linkedin.com/developers/apps"
log "     Products -> add 'Share on LinkedIn' and 'Sign In with LinkedIn'"
log "     Auth -> add redirect URL: http://localhost:8085/callback"
log "  2. write your credentials to ~/.config/postiee/credentials.toml:"
log "       client_id = \"...\""
log "       client_secret = \"...\""
log "     then: chmod 600 ~/.config/postiee/credentials.toml"
log "  3. $BIN auth login --port 8085"
log ""