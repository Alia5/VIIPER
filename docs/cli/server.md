# Server Command

Start the VIIPER USBIP server to expose virtual devices.

## Usage

```bash
viiper server [OPTIONS]
```

## Description

The `server` command starts the VIIPER USBIP server, which allows you to create and manage virtual USB devices that appear as real hardware to USBIP clients.

The server exposes two interfaces:

1. **USBIP Server** - Standard USBIP protocol for device attachment
2. **API Server** - Management API for device/bus control

## Options

### `--usb.addr`

USBIP server listen address.

**Default:** `:3241`  
**Environment Variable:** `VIIPER_USB_ADDR`

**Example:**

```bash
viiper server --usb.addr=0.0.0.0:3241
```

### `--api.addr`

API server listen address.

**Default:** `:3242`  
**Environment Variable:** `VIIPER_API_ADDR`

**Example:**

```bash
# Enable API on custom port
viiper server --api.addr=:8080
```

### `--api.device-handler-timeout`

Time before auto-cleanup occurs when a device handler has no active connection.

**Default:** `5s`  
**Environment Variable:** `VIIPER_API_DEVICE_HANDLER_TIMEOUT`

**Example:**

```bash
viiper server --api.device-handler-timeout=10s
```

### `--connection-timeout`

Connection operation timeout for both USBIP and API servers.

**Default:** `30s`  
**Environment Variable:** `VIIPER_CONNECTION_TIMEOUT`

**Example:**

```bash
viiper server --connection-timeout=60s
```

## Examples

### Basic Server

Start server with default settings (USBIP on :3241, API on :3242):

```bash
viiper server
```

### Custom Addresses

Start server on custom ports:

```bash
viiper server --usb.addr=:9000 --api.addr=:9001
```

### With Logging

Start server with debug logging to file:

```bash
viiper server --log.level=debug --log.file=/var/log/viiper.log
```

### With Raw Packet Logging

Start server with raw USB packet logging (useful for reverse engineering):

```bash
viiper server --log.raw-file=/var/log/viiper-raw.log
```

## Connect from a client (USBIP)

After the server is running and a virtual device has been added to a bus (via the API), attach it from a client using USBIP.

Notes:

- VIIPER's USBIP server listens on `:3241` by default (configurable via `--usb.addr`).
- The BUSID-DEVICEID you need (e.g. `1-1`) is returned by the API on device add and also visible via `usbip list`.

### Linux

```bash
# Load the virtual host controller (only needed once per boot)
sudo modprobe vhci-hcd

# List exportable devices on the VIIPER host
usbip list --remote=VIIPER_HOST --tcp-port=3241

# Attach a device by busid (long flags)
sudo usbip attach --remote=VIIPER_HOST --tcp-port=3241 --busid=BUSID-DEVICEID

# Equivalent short-form flags
sudo usbip --tcp-port 3241 -r VIIPER_HOST -b BUSID-DEVICEID
```

Replace `VIIPER_HOST` with the server's hostname/IP. If you changed the USBIP port, use that port instead of `3241`.

### Windows

On Windows, use [usbip-win2](https://github.com/vadimgrn/usbip-win2):

- GUI: use the client to add a remote host and attach by busid.
- CLI (similar flags):

```powershell
usbip.exe list --remote VIIPER_HOST --tcp-port 3241
usbip.exe attach --remote VIIPER_HOST --tcp-port 3241 --busid BUSID-DEVICEID
```

Once attached, the device will appear to the OS/applications as a local USB device.

## See Also

- [Configuration](configuration.md) - Environment variables and configuration files
- [API Reference](../api/overview.md) - API server documentation
