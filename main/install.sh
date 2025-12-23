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
	echo "Error: Could not fetch VIIPER release"
	exit 1
fi

echo "Version: $VERSION"

ARCH=$(uname -m)

case "$ARCH" in
	x86_64) ARCH="amd64" ;;
	aarch64|arm64) ARCH="arm64" ;;
	*)
		echo "Error: Unsupported architecture: $ARCH"
		echo "Supported: x86_64 (amd64), aarch64/arm64"
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
	echo "Error: Could not download VIIPER binary"
	exit 1
fi

chmod +x viiper

NEW_VERSION=$(./viiper --help -p | grep -Eo 'Version: [^ ]+' | head -1 | cut -d' ' -f2)
if [ -z "$NEW_VERSION" ]; then
	echo "Warning: Could not extract version from downloaded binary"
	NEW_VERSION="unknown"
fi
echo "Downloaded VIIPER version: $NEW_VERSION"

INSTALL_DIR="/usr/local/bin"
INSTALL_PATH="$INSTALL_DIR/viiper"

IS_UPDATE=0
SKIP_INSTALL=0
if [ -f "$INSTALL_PATH" ]; then
	IS_UPDATE=1

	OLD_VERSION=$("$INSTALL_PATH" --help -p | grep -Eo 'Version: [^ ]+' | head -1 | cut -d' ' -f2)
	if [ -z "$OLD_VERSION" ]; then
		echo "Warning: Could not extract version from installed binary"
		OLD_VERSION="unknown"
	fi
	echo "Installed VIIPER version: $OLD_VERSION"

	if [ "$NEW_VERSION" = "$OLD_VERSION" ] && [ "$NEW_VERSION" != "unknown" ]; then
		echo "Versions are identical. Skipping VIIPER install step."
		SKIP_INSTALL=1
	else
		if [ "$NEW_VERSION" != "unknown" ] && [ "$OLD_VERSION" != "unknown" ]; then
			LOWEST=$(printf '%s\n' "$OLD_VERSION" "$NEW_VERSION" | sort -V | head -n1)
			if [ "$LOWEST" = "$NEW_VERSION" ] && [ "$OLD_VERSION" != "$NEW_VERSION" ]; then
				echo "Detected potential downgrade (installed: $OLD_VERSION, new: $NEW_VERSION). Skipping install."
				SKIP_INSTALL=1
			fi
		fi
	fi
fi

if [ "$IS_UPDATE" -eq 1 ] && [ "$SKIP_INSTALL" -eq 0 ]; then
	echo "Stopping VIIPER service if running..."
	sudo systemctl stop viiper.service || true
fi

if [ "$SKIP_INSTALL" -eq 0 ]; then
	echo "Installing binary to $INSTALL_PATH..."
	sudo mkdir -p "$INSTALL_DIR"
	sudo cp viiper "$INSTALL_PATH"
	sudo chmod +x "$INSTALL_PATH"
else
	echo "Binary already at correct version, skipping installation."
fi


detect_package_manager() {
	if command -v pacman >/dev/null ; then
		echo "pacman"
		return 0
	fi
	if command -v apt >/dev/null ; then
		echo "apt"
		return 0
	fi
	if command -v apt-get >/dev/null ; then
		echo "apt"
		return 0
	fi
	if command -v dnf >/dev/null ; then
		echo "dnf"
		return 0
	fi
	return 1
}

is_steamos() {
	if command -v steamos-readonly >/dev/null; then
		return 0
	fi
	if [ -r /etc/os-release ] && grep -qi '^ID=steamos' /etc/os-release; then
		return 0
	fi
	return 1
}

STEAMOS_RW_TOGGLED=0

echo ""
echo "Checking USBIP installation..."

if command -v usbip >/dev/null ; then
	echo "USBIP already installed"
else
	echo "USBIP not found. Installing..."

	if is_steamos; then
		echo "SteamOS detected"
		if command -v steamos-readonly >/dev/null ; then
			if steamos-readonly status | grep -q "enabled"; then
				echo "Read-only root is enabled. Temporarily disabling..."
				if steamos-readonly disable; then
					echo "Read-only root disabled"
					STEAMOS_RW_TOGGLED=1
				else
					echo "Warning: Could not disable read-only root. USBIP installation may fail."
				fi
			else
				echo "Read-only root is already disabled"
			fi
		fi
	fi

	PM=$(detect_package_manager) || PM=""
	case "$PM" in
		pacman)
			echo "Installing USBIP via pacman..."
			sudo pacman -S --noconfirm usbip || echo "Warning: USBIP installation failed"
			;;
		apt)
			echo "Installing USBIP via apt..."
			sudo apt update
			sudo apt install -y linux-tools-generic || echo "Warning: USBIP installation failed"
			;;
		dnf)
			echo "Installing USBIP via dnf..."
			sudo dnf install -y usbip || echo "Warning: USBIP installation failed"
			;;
		*)
			echo "Warning: Could not detect package manager. Please install USBIP manually."
			echo "See: https://alia5.github.io/VIIPER/stable/getting-started/installation/"
			;;
	esac
fi

if [ "$IS_UPDATE" -eq 1 ]; then
	echo "Existing VIIPER installation detected. Preserving startup/autostart configuration..."
else
	echo "Configuring system startup..."
fi

echo "Checking vhci_hcd kernel module..."
MODULE_ALREADY_LOADED=0
if lsmod | grep -q vhci_hcd; then
	MODULE_ALREADY_LOADED=1
	echo "vhci_hcd module is already loaded"
else
	echo "Loading vhci_hcd kernel module..."
	if sudo modprobe vhci_hcd; then
		echo "vhci_hcd module loaded"
	else
		echo "Warning: Could not load vhci_hcd module"
	fi
fi

MODULES_CONF="/etc/modules-load.d/viiper.conf"
if [ $MODULE_ALREADY_LOADED -eq 0 ]; then
	echo "Configuring vhci_hcd to load at boot..."
	if echo "vhci_hcd" | sudo tee "$MODULES_CONF" >/dev/null; then
		echo "Module persistence configured: $MODULES_CONF"
	else
		echo "Warning: Could not configure module persistence"
	fi
else
	echo "vhci_hcd module is already loaded, skipping autoload configuration"
fi

if [ "$SKIP_INSTALL" -eq 0 ]; then
	echo "Creating systemd service..."
	sudo "$INSTALL_PATH" install
fi

if [ "$STEAMOS_RW_TOGGLED" -eq 1 ]; then
	echo "Re-enabling SteamOS read-only root..."
	steamos-readonly enable || echo "Warning: failed to re-enable read-only. You may re-enable it manually later."
fi

echo ""
echo "VIIPER installed successfully!"
echo "Binary installed to: $INSTALL_PATH"

if [ "$IS_UPDATE" -eq 1 ] && [ "$SKIP_INSTALL" -eq 0 ]; then
	echo "Update complete. VIIPER service has been restarted."
elif [ "$IS_UPDATE" -eq 0 ]; then
	echo "VIIPER server is now running and will start automatically on boot."
fi
