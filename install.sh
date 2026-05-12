#!/usr/bin/env bash
#
# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

# install.sh — downloads the gardener-landscape-kit binary matching the version
# specified in a components.yaml file.
#
# Usage:
#   /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/gardener/gardener-landscape-kit/HEAD/install.sh)"
#   /bin/bash -c "$(curl -fsSL ...)" -- --components-file path/to/components.yaml --install-dir ~/.local/bin
#
# Options:
#   --components-file PATH   Path to components.yaml (default: ./components.yaml)
#   --install-dir DIR        Directory to install the binary into (default: ~/.glk/bin)
#   --no-symlink             Do not create/update a 'glk' symlink in install-dir
#   --help                   Print this help
#
# END_HELP

GITHUB_REPO="gardener/gardener-landscape-kit"
GLK_COMPONENT_NAME="github.com/gardener/gardener-landscape-kit"
BINARY_NAME="gardener-landscape-kit"

# Defaults
COMPONENTS_FILE="${COMPONENTS_FILE:-./components.yaml}"
INSTALL_DIR="${INSTALL_DIR:-${HOME}/.glk/bin}"
CREATE_SYMLINK=true

# ── helpers ──────────────────────────────────────────────────────────────────

log() { echo "[glk-install] $*"; }
die() { echo "[glk-install] ERROR: $*" >&2; exit 1; }

usage() {
  awk '/^# install\.sh/,/^# END_HELP/' "$0" | grep -v '^# END_HELP' | sed 's/^# \{0,2\}//'
  exit 0
}

# ── argument parsing ──────────────────────────────────────────────────────────

while [[ $# -gt 0 ]]; do
  case "$1" in
    --components-file) COMPONENTS_FILE="$2"; shift 2 ;;
    --install-dir)     INSTALL_DIR="$2";     shift 2 ;;
    --no-symlink)      CREATE_SYMLINK=false; shift   ;;
    --help|-h)         usage ;;
    *) die "Unknown argument: $1. Run with --help for usage." ;;
  esac
done

# ── detect version from components.yaml ──────────────────────────────────────

[[ -f "$COMPONENTS_FILE" ]] || die "components.yaml not found at '$COMPONENTS_FILE'. Pass --components-file PATH."

# Extract the version field of the GLK entry.
# The YAML structure is:
#   components:
#   - name: github.com/gardener/gardener-landscape-kit
#     version: vX.Y.Z
#
# We look for the name line, then grab the next 'version:' line that follows.
VERSION="$(awk -v component="${GLK_COMPONENT_NAME}" '
  index($0, "name:") && index($0, component) { found=1; next }
  found && index($0, "version:") { print; exit }
' "$COMPONENTS_FILE" | sed 's/.*version:[[:space:]]*["'\'']\{0,1\}\([^"'\'' ]*\)["'\'']\{0,1\}.*/\1/')"

[[ -n "$VERSION" ]] || die "Could not find version for '${GLK_COMPONENT_NAME}' in '$COMPONENTS_FILE'."

log "Found GLK version: ${VERSION}"

# ── detect OS and architecture ────────────────────────────────────────────────

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  linux)   OS="linux" ;;
  darwin)  OS="darwin" ;;
  mingw*|msys*|cygwin*|windows*) OS="windows" ;;
  *) die "Unsupported OS: $(uname -s)" ;;
esac

ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64)    ARCH="amd64" ;;
  aarch64|arm64)   ARCH="arm64" ;;
  *) die "Unsupported architecture: $(uname -m)" ;;
esac

ASSET_NAME="${BINARY_NAME}-${OS}-${ARCH}"
log "Platform: ${OS}/${ARCH} → asset '${ASSET_NAME}'"

# ── download helper ───────────────────────────────────────────────────────────

download() {
  local url="$1" dest="$2"
  if command -v curl &>/dev/null; then
    curl -fSL --progress-bar -o "$dest" "$url"
  elif command -v wget &>/dev/null; then
    wget -q --show-progress -O "$dest" "$url"
  else
    die "Neither curl nor wget found. Install one and retry."
  fi
}

# ── resolve download URL ──────────────────────────────────────────────────────

if [[ "$OS" == "windows" ]]; then
  ARCHIVE_SUFFIX=".zip"
else
  ARCHIVE_SUFFIX=".tar.gz"
fi

DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${ASSET_NAME}${ARCHIVE_SUFFIX}"

# ── prepare install directory and target path ─────────────────────────────────

mkdir -p "$INSTALL_DIR"
VERSIONED_BINARY="${INSTALL_DIR}/${ASSET_NAME}-${VERSION}"
SYMLINK_PATH="${INSTALL_DIR}/glk"

# ── extract helper ────────────────────────────────────────────────────────────

extract_asset() {
  local archive="$1" dest_dir="$2" binary_name="$3"
  local dest="${dest_dir}/${binary_name}"

  if [[ "$OS" == "windows" ]]; then
    unzip -o -d "$dest_dir" "$archive"
  else
    tar xzf "$archive" -C "$dest_dir"
  fi

  if [[ -f "$dest" ]]; then
    mv "$dest" "${VERSIONED_BINARY}"
  else
    die "Expected binary '${binary_name}' not found in archive."
  fi
}

# ── download and install ──────────────────────────────────────────────────────

# Skip download if the exact versioned binary already exists
if [[ -f "$VERSIONED_BINARY" ]]; then
  log "Binary already cached at '${VERSIONED_BINARY}', skipping download."
else
  ARCHIVE_TMP="$(mktemp)"
  trap 'rm -f "$ARCHIVE_TMP"' EXIT

  log "Downloading ${DOWNLOAD_URL} ..."
  if ! download "$DOWNLOAD_URL" "$ARCHIVE_TMP"; then
    # On darwin/arm64, older releases may only ship amd64 — fall back via Rosetta
    if [[ "$OS" == "darwin" && "$ARCH" == "arm64" ]]; then
      log "arm64 asset not found, falling back to amd64 (runs via Rosetta 2)."
      ARCH="amd64"
      ASSET_NAME="${BINARY_NAME}-${OS}-${ARCH}"
      DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${ASSET_NAME}${ARCHIVE_SUFFIX}"
      VERSIONED_BINARY="${INSTALL_DIR}/${ASSET_NAME}-${VERSION}"
      download "$DOWNLOAD_URL" "$ARCHIVE_TMP" \
        || die "Download failed. Check that version '${VERSION}' exists at https://github.com/${GITHUB_REPO}/releases"
    else
      die "Download failed. Check that version '${VERSION}' exists at https://github.com/${GITHUB_REPO}/releases"
    fi
  fi

  if [[ "$OS" == "windows" ]]; then
    INNER_BINARY="${BINARY_NAME}-${OS}-${ARCH}.exe"
  else
    INNER_BINARY="${BINARY_NAME}-${OS}-${ARCH}"
  fi

  extract_asset "$ARCHIVE_TMP" "$INSTALL_DIR" "$INNER_BINARY"
  log "Installed to '${VERSIONED_BINARY}'."
fi

chmod +x "$VERSIONED_BINARY"

# ── create / update symlink ───────────────────────────────────────────────────

if [[ "$CREATE_SYMLINK" == true ]]; then
  ln -sf "$VERSIONED_BINARY" "$SYMLINK_PATH"
  log "Symlink updated: '${SYMLINK_PATH}' → '${VERSIONED_BINARY}'"
fi

# ── print next steps ──────────────────────────────────────────────────────────

echo ""
echo "────────────────────────────────────────────────────────"
echo " gardener-landscape-kit ${VERSION} installed successfully"
echo "────────────────────────────────────────────────────────"
echo ""
echo "Binary:  ${VERSIONED_BINARY}"
if [[ "$CREATE_SYMLINK" == true ]]; then
  echo "Symlink: ${SYMLINK_PATH}"
fi
echo ""

# Check whether INSTALL_DIR is already in PATH
if [[ ":${PATH}:" != *":${INSTALL_DIR}:"* ]]; then
  echo "Add the install directory to your PATH by running:"
  echo ""
  echo "    export PATH=\"${INSTALL_DIR}:\$PATH\""
  echo ""
  echo "To make it permanent, add the line above to your ~/.bashrc or ~/.zshrc."
  echo ""
fi

echo "Next steps:"
echo ""
echo "    glk generate base <TARGET_DIR>"
echo ""
