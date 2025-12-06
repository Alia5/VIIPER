# Quick Start

This guide walks you through setting up VIIPER and creating your first virtual device.

## Prerequisites

Before starting, ensure you have:

1. **USBIP installed** on your system (see [Installation](installation.md#requirements))
2. **VIIPER binary** downloaded from [GitHub Releases](https://github.com/Alia5/VIIPER/releases) or [built from source](installation.md#building-from-source)

## Starting the Server

Start VIIPER with default settings:

```bash
viiper server
```

This starts two services:

- **USBIP Server** on port `3241` (standard USBIP protocol)
- **API Server** on port `3242` (device management)

!!! tip "Auto-attach Feature"
    By default, VIIPER automatically attaches newly created devices to the local machine. You can disable this with `--api.auto-attach-local-client=false`.

### Custom Ports

To use different ports:

```bash
viiper server --usb.addr=:9000 --api.addr=:9001
```

## Creating Your First Virtual Device

VIIPER provides multiple ways to interact with the API. Choose the method that works best for you.

### Option 1: Using Client Libraries (Recommended)

Client libraries are available for C, C#, Go, and TypeScript. They handle the protocol details automatically, providing type-safe interfaces and device-specific helpers.

For complete client library documentation and code examples, see:

- [C Client Library Documentation](../clients/c.md)
- [C# Client Library Documentation](../clients/csharp.md)
- [TypeScript Client Library Documentation](../clients/typescript.md)
- [Go Client Documentation](../clients/go.md)

Full working examples for all device types are available in the `examples/` directory of the repository.

### Option 2: Using Raw TCP (netcat)

For quick testing without client libraries:

```bash
# Create a bus
printf "bus/create\0" | nc localhost 3242
# Response: {"busId":1}

# Add a keyboard device
printf 'bus/1/add {"type":"keyboard"}\0' | nc localhost 3242
# Response: {"busId":1,"devId":"1","vid":"0x2e8a","pid":"0x0010","type":"keyboard"}

# List devices on the bus
printf "bus/1/list\0" | nc localhost 3242
```

!!! note "Protocol Details"
    The API uses TCP with null-byte (`\0`) terminated requests. See [API Reference](../api/overview.md) for complete protocol documentation.

### Option 3: Using PowerShell Helper Script

VIIPER includes a PowerShell helper script for Windows users:

```powershell
# Load the helper script
. .\scripts\viiper-api.ps1

# Create a bus
Invoke-ViiperAPI "bus/create"

# Add a device
Invoke-ViiperAPI 'bus/1/add {"type":"keyboard"}'
```

## Attaching Devices (USBIP)

After creating a device via the API, attach it using your system's USBIP client.

!!! success "Automatic Attachment"
    If you're running VIIPER on the same machine where you want to use the device, it's likely already attached automatically! Check your device manager or `lsusb` to confirm.

### Manual Attachment

If auto-attach is disabled or you're connecting from a remote machine:

=== "Linux"

    ```bash
    # Load kernel module (once per boot)
    sudo modprobe vhci-hcd

    # List available devices
    usbip list --remote=localhost --tcp-port=3241

    # Attach device (use busid from API response, e.g., "1-1")
    sudo usbip attach --remote=localhost --tcp-port=3241 --busid=1-1

    # Verify attachment
    lsusb | grep "Raspberry Pi"  # For keyboard/mouse
    lsusb | grep "Microsoft"     # For Xbox 360 controller
    ```

=== "Windows"

    Using [usbip-win2](https://github.com/vadimgrn/usbip-win2):

    ```powershell
    # List available devices
    usbip.exe list --remote localhost --tcp-port 3241

    # Attach device
    usbip.exe attach --remote localhost --tcp-port 3241 --busid 1-1

    # Check Device Manager to verify attachment
    ```

## Available Device Types

VIIPER supports multiple virtual device types including keyboards, mice, and game controllers. Each device type has its own protocol and capabilities.

For a complete list of supported devices, their specifications, and wire protocols, see the [Devices](../devices/) documentation.

## Next Steps

Now that you have a working setup:

1. **Explore Examples**: Check the `examples/` directory for complete working programs in C, C#, Go, and TypeScript
2. **Read API Documentation**: Learn about all available [API commands](../api/overview.md)
3. **Choose a Client Library**: Pick a [client library](../clients/generator.md) for your preferred language
4. **Review Device Specs**: Understand device-specific protocols in [Devices](../devices/keyboard.md)

## Troubleshooting

### Server Won't Start

**Port already in use:**

```bash
# Use custom ports
viiper server --usb.addr=:9000 --api.addr=:9001
```

**Permission denied (Linux):**

```bash
# Use ports above 1024 or run with sudo
viiper server --usb.addr=:3241 --api.addr=:3242
```

### Auto-Attach Not Working

VIIPER will check prerequisites at startup when auto-attach is enabled and log warnings if requirements are missing.

**Linux - USBIP tool not found:**

```bash
# Ubuntu/Debian
sudo apt install linux-tools-generic

# Arch Linux
sudo pacman -S usbip
```

**Linux - Kernel module not loaded:**

```bash
# Load for current session
sudo modprobe vhci-hcd

# Or configure persistent loading (see Installation guide)
```

See [Linux Kernel Module Setup](installation.md#linux-kernel-module-setup-for-auto-attach) for detailed setup instructions.

**Windows - USBIP tool not found:**

Download and install [usbip-win2](https://github.com/vadimgrn/usbip-win2) and ensure `usbip.exe` is in your PATH.

### Device Not Attaching

**USBIP tool not found:**

Make sure USBIP is installed and in your PATH (see [Installation requirements](installation.md#requirements)).

**Connection refused:**

Verify the VIIPER server is running and listening on the expected ports.

### Device Not Working

**No input response:**

Ensure the device is attached via USBIP AND you've opened a device stream via the API to send input data.

**Multiple VIIPER instances:**

If you have VIIPER running as a service, your application's instance may conflict. Either connect to the existing instance or use different ports.

## See Also

- [CLI Reference](../cli/overview.md) - Complete command documentation
- [API Reference](../api/overview.md) - Management API protocol
- [Client Libraries](../clients/generator.md) - Language-specific client libraries
- [Configuration](../cli/configuration.md) - Environment variables and config files
