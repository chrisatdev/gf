#!/usr/bin/env bash
# install.sh — download and install the latest gf release
set -euo pipefail

REPO="chrisatdev/gf"
INSTALL_DIR="${GF_INSTALL_DIR:-$HOME/.local/bin}"
API_URL="https://api.github.com/repos/${REPO}/releases/latest"

# ---------------------------------------------------------------------------
# Platform detection
# ---------------------------------------------------------------------------
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux)  GOOS="linux"  ;;
  Darwin) GOOS="darwin" ;;
  *)
    echo "error: unsupported OS \"${OS}\". gf supports Linux and macOS. Windows users should use WSL 2." >&2
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64)          GOARCH="amd64" ;;
  aarch64 | arm64) GOARCH="arm64" ;;
  *)
    echo "error: unsupported architecture \"${ARCH}\"." >&2
    exit 1
    ;;
esac

ASSET_NAME="gf_${GOOS}_${GOARCH}.tar.gz"

# ---------------------------------------------------------------------------
# Fetch latest release metadata
# ---------------------------------------------------------------------------
echo "Fetching latest release from ${REPO}..."
RELEASE_JSON="$(curl -fsSL "${API_URL}")"
VERSION="$(echo "${RELEASE_JSON}" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"

if [ -z "$VERSION" ]; then
  echo "error: could not determine latest release version." >&2
  exit 1
fi

echo "Latest version: ${VERSION}"

# ---------------------------------------------------------------------------
# Locate asset download URL
# ---------------------------------------------------------------------------
DOWNLOAD_URL="$(echo "${RELEASE_JSON}" | grep '"browser_download_url"' | grep "${ASSET_NAME}" | head -1 | sed 's/.*"browser_download_url": *"\([^"]*\)".*/\1/')"

if [ -z "$DOWNLOAD_URL" ]; then
  echo "error: no asset named \"${ASSET_NAME}\" found in release ${VERSION}." >&2
  exit 1
fi

CHECKSUMS_URL="$(echo "${RELEASE_JSON}" | grep '"browser_download_url"' | grep 'checksums.txt' | head -1 | sed 's/.*"browser_download_url": *"\([^"]*\)".*/\1/')"

# ---------------------------------------------------------------------------
# Download
# ---------------------------------------------------------------------------
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

echo "Downloading ${ASSET_NAME}..."
curl -fsSL -o "${TMP_DIR}/${ASSET_NAME}" "${DOWNLOAD_URL}"

# ---------------------------------------------------------------------------
# Checksum verification (best-effort — skip if checksums asset not found)
# ---------------------------------------------------------------------------
if [ -n "$CHECKSUMS_URL" ]; then
  echo "Verifying checksum..."
  curl -fsSL -o "${TMP_DIR}/checksums.txt" "${CHECKSUMS_URL}"
  (
    cd "${TMP_DIR}"
    if command -v sha256sum > /dev/null 2>&1; then
      grep "${ASSET_NAME}" checksums.txt | sha256sum -c -
    elif command -v shasum > /dev/null 2>&1; then
      grep "${ASSET_NAME}" checksums.txt | shasum -a 256 -c -
    else
      echo "warning: sha256sum / shasum not found — skipping checksum verification."
    fi
  )
fi

# ---------------------------------------------------------------------------
# Extract and install
# ---------------------------------------------------------------------------
echo "Extracting..."
tar -xzf "${TMP_DIR}/${ASSET_NAME}" -C "${TMP_DIR}"

if [ ! -f "${TMP_DIR}/gf" ]; then
  echo "error: gf binary not found in archive." >&2
  exit 1
fi

mkdir -p "${INSTALL_DIR}"
mv "${TMP_DIR}/gf" "${INSTALL_DIR}/gf"
chmod +x "${INSTALL_DIR}/gf"

# ---------------------------------------------------------------------------
# PATH hint
# ---------------------------------------------------------------------------
echo ""
echo "gf ${VERSION} installed to ${INSTALL_DIR}/gf"

case ":${PATH}:" in
  *":${INSTALL_DIR}:"*)
    ;;
  *)
    echo ""
    echo "hint: ${INSTALL_DIR} is not in your PATH."
    echo "      Add the following line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    echo ""
    echo "      export PATH=\"\$HOME/.local/bin:\$PATH\""
    ;;
esac
