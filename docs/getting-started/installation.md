# Installation

VIIPER _currently_ comes in a single flavor:

- a standalone executable that exposes an API over TCP.
- There will eventually be a shared-library version (libVIIPER) that you can link against directly from your application.  
For more information, see [FAQ](https://github.com/Alia5/VIIPER#why-is-this-a-standalone-executable-that-i-have-to-interface-via-tcp-and-not-a-shared-object-library-in-itself)

## Requirements

VIIPER relies on USBIP.  
You must have a USBIP-Client implementation available on your system to use VIIPER's virtual devices.

### Linux

#### Ubuntu/Debian

```bash
sudo apt install linux-tools-generic
```

[Ubuntu USBIP Manual](https://manpages.ubuntu.com/manpages/noble/man8/usbip.8.html)

#### Arch Linux

```bash
sudo pacman -S usbip
```

[Arch Wiki: USBIP](https://wiki.archlinux.org/title/USB/IP)

### Windows

[usbip-win2](https://github.com/vadimgrn/usbip-win2) is by far the most complete implementation of USBIP for Windows (comes with a **SIGNED** kernel mode driver).

### Linux Kernel Module Setup

!!! info "USBIP Client Requirement"
    USBIP requires the `vhci-hcd` (Virtual Host Controller Interface) kernel module on Linux for client operations. This includes VIIPER's auto-attach feature and manual device attachment.

Most Linux distributions include this module but don't load it automatically.

#### One-Time Setup

To load the module automatically on boot:

```bash
echo "vhci-hcd" | sudo tee /etc/modules-load.d/vhci-hcd.conf
sudo modprobe vhci-hcd
```

#### Manual Loading

To load the module for the current session only:

```bash
sudo modprobe vhci-hcd
```

#### Verification

Check if the module is loaded:

```bash
lsmod | grep vhci_hcd
```

## Installing VIIPER

VIIPER does not require system-wide installation.  
The `viiper` executable is completely self-contained (and fully portable without any dependencies, except USBIP) and can be:

- Placed in any directory
- Shipped alongside your application
- Run directly without installation
- Bundled with your application's distribution

This makes VIIPER ideal for embedding in applications or distributing as part of a software package.

!!! warning "Daemon/Service Conflicts"
    If VIIPER is already running as a system service or daemon on the target machine, be aware of potential port conflicts. Applications should:
    - Check if VIIPER is already running before starting their own instance  
      - use the `ping` API endpoint to check for VIIPER presence and version  
    - Connect to the existing VIIPER instance (if accessible)
    - Use a custom port via `--api.addr` flag to run a separate instance

### Pre-built Binaries

Download the latest release from the [GitHub Releases](https://github.com/Alia5/VIIPER/releases) page. Pre-built binaries are available for:

- Windows (x64, ARM64)
- Linux (x64, ARM64)

### Automated Install Script

Regardless of portability, it can be convenient to have VIIPER start automatically on system boot, especially if end users want to use your application through a network or you want to enable that possibility.  

The following scripts will download a VIIPER release, install it to a system location, and configure it to start automatically on boot.  

!!! info "For Application Developers"
    The installation scripts are intended for **end-users** setting up a permanent VIIPER service on their system.  
  
    If you're developing an application that uses VIIPER, I **strongly** encourage you to **not** install a permanent VIIPER service on your users machines. 
    
    Instead, bundle the (no dependencies, portable) VIIPER binary with your application and start/stop the server directly from your application as needed.  
    You may need to check for existing VIIPER instances or use a custom port via `--api.addr` to avoid conflicts.   

!!! warning "USBIP not included"
    The install scripts do **not** install/setup USBIP.  
    Make sure a USBIP-client is installed and configured before installing VIIPER.

**Linux:**

```bash
curl -fsSL https://alia5.github.io/VIIPER/install.sh | sh
```

Installs to: `/usr/local/bin/viiper`

**Windows (PowerShell):**

```powershell
irm https://alia5.github.io/VIIPER/install.ps1 | iex
```

Installs to: `%LOCALAPPDATA%\VIIPER\viiper.exe`

The scripts will:

1. Download the specified VIIPER binary version
2. Install it to the system location
3. Configure automatic startup (Registry RunKey on Windows, systemd service on Linux)
4. Start the VIIPER server

**Version-Specific Installation:**

The install scripts are version-aware based on where you download them from:

- **Latest stable release:**  
  `curl -fsSL https://alia5.github.io/VIIPER/install.sh | sh`

- **Specific version (e.g., v0.2.2):**  
  `curl -fsSL https://alia5.github.io/VIIPER/0.2.2/install.sh | sh`

- **Latest _pre_-release (development snapshot):**  
  `curl -fsSL https://alia5.github.io/VIIPER/main/install.sh | sh`

## System Startup Configuration

The `install` and `uninstall` commands configure automatic startup for the VIIPER binary.

!!! info "What These Commands Do"
    These commands **do not copy or move** the VIIPER binary. They configure your system to automatically run the binary from its **current location** when the system boots.

    Make sure the binary is in a permanent location before running `viiper install`!

### `viiper install`

Configures VIIPER to start automatically on system boot:

```bash
viiper install
```

- **Windows**:  
  - Adds entry to Registry RunKey: `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run\VIIPER`
  - Value: `"<current-exe-path>" server`
  - Kills any previous autorun instances
  - Starts the server

- **Linux**:  
  - Creates systemd service: `/etc/systemd/system/viiper.service`
  - Service ExecStart points to current binary path
  - Enables and starts the service

### `viiper uninstall`

Removes VIIPER from system startup and stops any running instance:

```bash
viiper uninstall
```

- **Windows**:  
  - Removes Registry RunKey entry
  - Kills any running autorun instances

- **Linux**:  
  - Stops and disables the systemd service
  - Removes `/etc/systemd/system/viiper.service`

## Building from Source

Building from source is only necessary if you need to modify VIIPER or target an unsupported platform.

### Prerequisites

- [Go](https://go.dev/) 1.25 or newer
- USBIP installed
- (Optional) [Make](https://www.gnu.org/software/make/)
    - Linux/macOS: Usually pre-installed
    - Windows: `winget install ezwinports.make`

### Build Steps

```bash
git clone https://github.com/Alia5/VIIPER.git
cd VIIPER
make build
```

The compiled binary will be in `dist/viiper` (or `dist/viiper.exe` on Windows).

**Additional build targets:**

```bash
make help          # Show all available make targets
make test          # Run tests
```
