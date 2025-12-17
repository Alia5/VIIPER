# CLI Overview

VIIPER provides a command-line interface for running the USBIP server and proxy.

## Commands

- [`server`](server.md) - Start the VIIPER USBIP server
- [`proxy`](proxy.md) - Start the VIIPER USBIP proxy
- `install` - Configure VIIPER to start automatically on system boot (see [Installation](../getting-started/installation.md#system-startup-configuration))
- `uninstall` - Remove VIIPER from system startup configuration
- [`codegen`](codegen.md) - Generate client libraries from source code annotations

## Global Options

### Logging

VIIPER supports flexible logging configuration via global flags or environment variables.

#### `--log.level`

Set the logging level.

**Values:** `trace`, `debug`, `info`, `warn`, `error`  
**Default:** `info`  
**Environment Variable:** `VIIPER_LOG_LEVEL`

**Example:**

```bash
viiper server --log.level=debug
```

#### `--log.file`

Log to a file in addition to console output.

**Default:** (none - logs only to console)  
**Environment Variable:** `VIIPER_LOG_FILE`

**Example:**

```bash
viiper server --log.file=/var/log/viiper.log
```

#### `--log.raw-file`

Log raw USB packet data to a file for debugging and reverse engineering.

**Default:** (none)  
**Environment Variable:** `VIIPER_LOG_RAW_FILE`

**Example:**

```bash
viiper server --log.raw-file=/var/log/viiper-raw.log
```

!!! note "Automatic Raw Logging"
    When `--log.level=trace` is set without `--log.raw-file`, raw packets are logged to stdout.

## Getting Help

Display help for any command:

```bash
viiper --help
viiper server --help
viiper proxy --help
```
