#!/usr/bin/env sh

set -e

VIIPER_VERSION="dev-snapshot"

REPO="Alia5/VIIPER"
API_URL="https://api.github.com/repos/${REPO}/releases/tags/${VIIPER_VERSION}"

echo "Fetching VIIPER release: $VIIPER_VERSION..."
RELEASE_DATA=$(curl -fsSL "$API_URL")
VERSION=$(printf '%s' "$RELEASE_DATA" \
	| grep -Eo '"tag_name"[[:space:]]*:[[:space:]]*"[^"]+"' \
	| head -n 1 \
	| cut -d'"' -f4)

if [ -z "$VERSION" ]; then
	echo "Error: Could not fetch VIIPER release" >&2
	exit 1
fi

echo "Version: $VERSION"

ARCH=$(uname -m)

case "$ARCH" in
	x86_64) ARCH="amd64" ;;
	aarch64|arm64) ARCH="arm64" ;;
	*)
		echo "Error: Unsupported architecture: $ARCH" >&2
		echo "Supported: x86_64 (amd64), aarch64/arm64" >&2
		exit 1
		;;
esac

BINARY_NAME="viiper-linux-${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}"

echo "Downloading from: $DOWNLOAD_URL"
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

cd "$TEMP_DIR"
if ! curl -fsSL -o viiper "$DOWNLOAD_URL"; then
	echo "Error: Could not download VIIPER binary" >&2
	exit 1
fi

chmod +x viiper

INSTALL_DIR="/usr/local/bin"
INSTALL_PATH="$INSTALL_DIR/viiper"

IS_UPDATE=0
if [ -f "$INSTALL_PATH" ]; then
	IS_UPDATE=1
fi

echo "Installing binary to $INSTALL_PATH..."
sudo mkdir -p "$INSTALL_DIR"
sudo cp viiper "$INSTALL_PATH"
sudo chmod +x "$INSTALL_PATH"

if [ "$IS_UPDATE" -eq 1 ]; then
	echo "Existing VIIPER installation detected (update). Preserving startup/autostart configuration..."
	# On update, do NOT run `viiper install` (it would enable/restart the systemd service).
	# We only replace the binary so the previous enable/disable choice remains intact.
else
	echo "Configuring system startup..."
	sudo "$INSTALL_PATH" install
fi

echo "VIIPER installed successfully!"
echo "Binary installed to: $INSTALL_PATH"
if [ "$IS_UPDATE" -eq 1 ]; then
	echo "Update complete. Startup/autostart configuration was left unchanged."
	echo "If VIIPER is running, restart it to use the updated binary."
else
	echo "VIIPER server is now running and will start automatically on boot."
fi
