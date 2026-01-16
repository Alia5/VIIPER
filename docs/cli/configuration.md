# Configuration

Aside from passing CLI flags, VIIPER can also be configured via environment variables and configuration files.

For configuration files, VIIPER supports `JSON`, `YAML`, and `TOML` formats.

## Environment Variables

All command-line flags have corresponding environment variables for easier deployment and configuration management.

### Global Configuration

| Environment Variable | CLI Flag | Default | Description |
|---------------------|----------|---------|-------------|
| `VIIPER_LOG_LEVEL` | `--log.level` | `info` | Logging level: `trace`, `debug`, `info`, `warn`, `error` |
| `VIIPER_LOG_FILE` | `--log.file` | (none) | Log file path (logs only to console if not set) |
| `VIIPER_LOG_RAW_FILE` | `--log.raw-file` | (none) | Raw packet log file path |

### Server Configuration

| Environment Variable | CLI Flag | Default | Description |
|---------------------|----------|---------|-------------|
| `VIIPER_USB_ADDR` | `--usb.addr` | `:3241` | USBIP server listen address |
| `VIIPER_API_ADDR` | `--api.addr` | `:3242` | API server listen address |
| `VIIPER_API_DEVICE_HANDLER_TIMEOUT` | `--api.device-handler-timeout` | `5s` | Device handler auto-cleanup timeout |
| `VIIPER_API_AUTO_ATTACH_LOCAL_CLIENT` | `--api.auto-attach-local-client` | `true` | Auto-attach exported devices to local usbip client |
| `VIIPER_CONNECTION_TIMEOUT` | `--connection-timeout` | `30s` | Connection operation timeout |

### Proxy Configuration

| Environment Variable | CLI Flag | Default | Description |
|---------------------|----------|---------|-------------|
| `VIIPER_PROXY_ADDR` | `--listen-addr` | `:3241` | Proxy listen address |
| `VIIPER_PROXY_UPSTREAM` | `--upstream` | (required) | Upstream USBIP server address |
| `VIIPER_PROXY_TIMEOUT` | `--connection-timeout` | `30s` | Connection timeout |

## Configuration Files

VIIPER supports JSON, YAML, and TOML configuration files.  
Generate a starter file (default configuration) with:

```bash
viiper config init server --format=json   # or yaml/toml
```

If no output path is provided, the file is written to the current working directory (e.g., `server.json`, `proxy.yaml`).

You can also specify a custom location:

```bash
viiper config init server --format=json --output ./server.json
```

To use a specific configuration file when starting VIIPER, pass the --config flag (or set the `VIIPER_CONFIG` environment variable):

```bash
viiper --config ./server.json server
```

If --config is not provided, VIIPER will search for configuration in this order and first-found is used for each format:

1. Working directory: server.(json|yaml|yml|toml), proxy.(json|yaml|yml|toml), viiper.(json|yaml|yml|toml), config.(json|yaml|yml|toml)
2. Platform config directory (see above): server.(json|yaml|yml|toml), proxy.(json|yaml|yml|toml), config.(json|yaml|yml|toml)
3. Linux system-wide: /etc/viiper/server.(json|yaml|yml|toml), /etc/viiper/proxy.(json|yaml|yml|toml), /etc/viiper/config.(json|yaml|yml|toml)

## Configuration Examples

=== "Config files"

    Server:

    ```json
    {
    "api": {
        "addr": ":3242",
        "device-handler-connect-timeout": "5s",
        "auto-attach-local-client": true
    },
    "usb": {
        "addr": ":3241"
    },
    "connection-timeout": "30s"
    }
    ```

    Proxy:

    ```json
    {
    "listen-addr": ":3241",
    "upstream-addr": "127.0.0.1:3242",
    "connection-timeout": "30s"
    }
    ```

=== "Environment Variables"

    Create a `.env` file or export variables:

    ```bash
    export VIIPER_LOG_LEVEL=debug
    export VIIPER_USB_ADDR=:3241
    export VIIPER_API_ADDR=:3242
    export VIIPER_LOG_FILE=/var/log/viiper.log
    ```

    Then run:

    ```bash
    viiper server
    ```

### Systemd Service

Example systemd service file for running VIIPER as a service:

```ini
[Unit]
Description=VIIPER USBIP Server
After=network.target

[Service]
Type=simple
User=viiper
Group=viiper
Environment="VIIPER_LOG_LEVEL=info"
Environment="VIIPER_LOG_FILE=/var/log/viiper/viiper.log"
Environment="VIIPER_USB_ADDR=:3241"
Environment="VIIPER_API_ADDR=:3242"
ExecStart=/usr/local/bin/viiper server
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

## Configuration Priority

When both CLI flags and environment variables are set, CLI flags take precedence:

1. **CLI flags** (highest priority)
2. **Environment variables**
3. **Configuration file values**
4. **Default values** (lowest priority)

## See Also

- [Server Command](server.md)
- [Proxy Command](proxy.md)
