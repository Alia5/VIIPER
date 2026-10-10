#!/usr/bin/env sh

set -e

VIIPER_VERSION="dev-snapshot"

REPO="Alia5/VIIPER"
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/tags/${VIIPER_VERSION}" \
	| grep -Eo '"tag_name"[[:space:]]*:[[:space:]]*"[^"]+"' \
	| head -n 1 \
	| cut -d'"' -f4)

case "$(uname -m)" in
	x86_64) ARCH="amd64" ;;
	aarch64|arm64) ARCH="arm64" ;;
	*)
		echo "Error: Unsupported architecture: $(uname -m)"
		exit 1
		;;
esac

INSTALLER=$(mktemp)
curl -fsSL -o "$INSTALLER" "https://github.com/${REPO}/releases/download/${VERSION}/viiper-linux-${ARCH}"
chmod +x "$INSTALLER"

SUDO=""
[ "$(id -u)" -eq 0 ] || SUDO="sudo"

if command -v systemd-run >/dev/null; then
	$SUDO systemd-run --collect "$INSTALLER" install --quiet
else
	$SUDO "$INSTALLER" install --quiet
fi
