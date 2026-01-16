# Proxy Command

The `proxy` command starts VIIPER in proxy mode, sitting between a USBIP client and a USBIP server.  
VIIPER intercepts and logs all USB traffic passing through, without handling the devices directly.

This is useful for reverse engineering USB protocols and understanding how devices communicate.

```
USBIP Client  →  VIIPER Proxy  →  USBIP Server (real devices or VIIPER)
                      ↓
              Logs/Captures Traffic
```

## Usage

```bash
viiper proxy --upstream=<address> [OPTIONS]
```

## Options

### `--listen-addr`

Proxy listen address (where clients connect).

**Default:** `:3241`  
**Environment Variable:** `VIIPER_PROXY_ADDR`

### `--upstream`

**Required.** Upstream USBIP server address (where real devices are).

**Environment Variable:** `VIIPER_PROXY_UPSTREAM`

### `--connection-timeout`

Connection timeout for proxy operations.

**Default:** `30s`  
**Environment Variable:** `VIIPER_PROXY_TIMEOUT`

## Examples

### Basic Proxy

Start proxy between local clients and remote USBIP server:

```bash
viiper proxy --upstream=192.168.1.100:3240
```

Clients connect to `localhost:3241`, traffic is proxied to `192.168.1.100:3240`.

### Custom Listen Address

Start proxy on a different port:

```bash
viiper proxy --listen-addr=:9000 --upstream=192.168.1.100:3240
```

Clients connect to `localhost:9000`, traffic is proxied to `192.168.1.100:3240`.

### With Raw Packet Logging

Capture all USB traffic for reverse engineering:

```bash
viiper proxy --upstream=192.168.1.100:3240 --log.raw-file=usb-capture.log
```

All USB packets will be logged to `usb-capture.log`.

### With Debug Logging

Enable debug logging to see proxy operations:

```bash
viiper proxy --upstream=192.168.1.100:3240 --log.level=debug
```

## Use Cases

### Reverse Engineering

Intercept USB traffic between a client and server to understand device protocols:

```bash
viiper proxy --upstream=real-server:3240 --log.raw-file=device-capture.log
```

### Traffic Analysis

Monitor USB communication for debugging:

```bash
viiper proxy --upstream=real-server:3240 --log.level=trace
```

### Network Inspection

Route USB traffic through VIIPER to inspect and log all operations:

```bash
viiper proxy --upstream=real-server:3240 --log.level=debug --log.raw-file=traffic.log
```

## See Also

- [Server Command](server.md) - Run VIIPER as a USBIP server
- [Configuration](configuration.md) - Environment variables and configuration files
