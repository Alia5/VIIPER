# API Reference

VIIPER ships a lightweight TCP API for managing virtual buses/devices and for device-specific streaming. It's designed to be trivial to drive from any language that can open a TCP socket and send newline-terminated commands.

!!! tip "Client SDKs Available"
    Generated client libraries are available that abstract away the protocol details described below. For most use cases, you should use one of the provided SDKs rather than implementing the raw protocol yourself:
    
    - [C SDK](../clients/c.md): Generated C library with type-safe device streams
    - [Go Client](../clients/go.md): Reference implementation included in the repository
    - [Generator Documentation](../clients/generator.md): Information about code generation for additional languages
    
    The documentation below is provided for reference and for implementing clients in languages not yet supported by the generator.

## Protocol overview

- Transport: TCP
- Default listen address: `:3242` (configurable via `--api.addr`)
- Request format: a single ASCII/UTF‑8 line terminated by `\n`
- Routing: first token is the path (e.g. `bus/list`), remaining tokens are arguments (space-separated)
- Success response: a single line containing a JSON payload (or an empty line for commands that have no payload)
- Error response: a single line JSON object `{ "error": "message" }`

Tip: You can experiment with `nc`/`ncat` or PowerShell’s `tcpclient` to send lines and read JSON back.

!!! warning "Connection timing and auto‑cleanup"
    After you add a device with `bus/{id}/add`, you must connect to its streaming endpoint within the configured `DeviceHandlerConnectTimeout` (default: 5s). If no stream connection is established in time, the device is automatically removed. Likewise, when a stream disconnects, a reconnection timer with the same timeout starts; if the client doesn’t reconnect before it expires, the device is removed.

## Commands

The server registers the following commands and streams:

- `bus/list`
    - List all virtual bus IDs.
    - Response: `{ "buses": [1, 2, ...] }`

- `bus/create [busId]`
    - Create a new bus. If `busId` is provided, VIIPER attempts to create the bus with that id; otherwise it picks the next free id.
    - Response: `{ "busId": <id> }`

- `bus/remove <busId>`
    - Remove a bus and all devices on it.
    - Response: `{ "busId": <id> }`

- `bus/{id}/list`
    - List devices on a bus.
    - Response: `{ "devices": [{ "busId": 1, "devId": "1", "vid": "0x045e", "pid": "0x028e", "type": "xbox360" }, ...] }`

- `bus/{id}/add <deviceType>`
    - Add a device to a bus. `deviceType` is a registered device name (e.g., `xbox360`).
    - Response: `{ "id": "<busId>-<devId>" }` where the id is the USBIP busid string you will attach to.
    - Important: After add, the server starts a connect timer (default `5s`). You must open a device stream (see below) before the timeout expires, otherwise the device is auto-removed.

- `bus/{id}/remove <deviceId>`
    - Remove a device by its device number on that bus (the part after the dash in the busid string).
    - Response: `{ "busId": <id>, "devId": "<dev>" }`

### Streaming endpoint

- Path: `bus/{busId}/{deviceid}`
- Type: long-lived TCP connection (no line protocol once established)
- Purpose: device-specific, bidirectional stream. The API server hands the socket to the device’s registered stream handler.
- Timeout behavior: When a stream ends, a reconnect timer is started (same `DeviceHandlerConnectTimeout`). If the client doesn’t reconnect in time, the device is removed.

#### Xbox 360 controller stream (device type: `xbox360`)

Direction: client ➜ server (input state)

- Fixed 14-byte packets, little-endian layout:
    - `Buttons` uint32 (4 bytes)
    - `LT` uint8, `RT` uint8 (2 bytes)
    - `LX, LY, RX, RY` int16 each (8 bytes)

Direction: server ➜ client (rumble)

- Fixed 2-byte packets:
    - `LeftMotor` uint8, `RightMotor` uint8

See `pkg/device/xbox360/protocol.go` for full details.

#### HID keyboard stream (device type: `keyboard`)

Direction: client ➜ server (keys pressed)

- Variable-length packets per frame:
    - Header: Modifiers uint8, KeyCount uint8
    - Body: KeyCount bytes of HID Usage IDs for currently pressed (non-modifier) keys

Direction: server ➜ client (LED state)

- 1-byte packets whenever host LED state changes:
    - Bit 0 NumLock, Bit 1 CapsLock, Bit 2 ScrollLock

Host-facing HID input report is 34 bytes: [Modifiers (1), Reserved (1), 256-bit key bitmap (32)].

See `pkg/device/keyboard/` for helpers and constants.

#### HID mouse stream (device type: `mouse`)

Direction: client ➜ server (motion/buttons)

- Fixed 5-byte packets per frame:
    - Buttons uint8 (bits 0..4)
    - dX int8, dY int8
    - Wheel int8, Pan int8

Direction: server ➜ client

- None (mouse is input-only)

Note: Motion and wheel deltas are consumed after each IN report so movement is relative.

Note on protocol compatibility:

- The wire format is modeled after the XInput gamepad state (XINPUT_GAMEPAD) but is not byte‑for‑byte identical. Key differences:
    - Buttons are encoded as a 32‑bit little‑endian field (XInput uses a 16‑bit bitmask), making the packet 14 bytes instead of 12.
    - No header or framing: packets are fixed‑length and back‑to‑back on the TCP stream.
    - Endianness is little‑endian for all multi‑byte fields.

## Example sessions

### Using netcat (Linux/macOS)

```bash
# List buses
printf "bus/list\n" | nc localhost 3242

# Create a bus
printf "bus/create\n" | nc localhost 3242
# → {"busId":1}

# Add a virtual Xbox 360 controller to bus 1
printf "bus/1/add xbox360\n" | nc localhost 3242
# → {"id":"1-1"}

# List devices on bus 1
printf "bus/1/list\n" | nc localhost 3242
```

Then, open a second TCP connection for streaming to `bus/1/1` (the API port, not the USBIP port). You’ll write 14‑byte input packets and read 2‑byte rumble packets. Any language with raw TCP support works.

### WIndows (PowerShell)

VIIPER includes convenience scripts for quick testing and automation:

```powershell
# Source the script to load helper functions
. .\scripts\viiper-api.ps1

# Use the Invoke-ViiperApi function (or 'viiper' alias)
viiper "bus/list"
viiper "bus/create"
viiper "bus/1/add xbox360" -Port 3242 -Hostname localhost
```

The script provides `Invoke-ViiperApi` (alias: `viiper`) for sending commands and `Connect-ViiperDevice` for testing persistent device connections.

### Go snippet (raw)

```go
package main

import (
    "bufio"
    "fmt"
    "net"
)

func main() {
    conn, _ := net.Dial("tcp", "localhost:3242")
    defer conn.Close()
    fmt.Fprintln(conn, "bus/create")
    r := bufio.NewReader(conn)
    line, _ := r.ReadString('\n')
    fmt.Println(line) // {"busId":1}
}
```

For a higher-level experience, see the Go client in `pkg/apiclient/`.

## How this relates to USBIP

The API controls which virtual devices exist and exposes a device stream for live input/feedback. Separately, the USBIP server (default `:3241`) makes these devices attachable from clients. Typical flow:

1) Create a bus ➜ 2) Add a device ➜ 3) Connect the device stream ➜ 4) From a client, attach using USBIP by `busid` (see the Server command page for exact `usbip` syntax).
