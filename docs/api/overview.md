# API Reference

<style>
    .md-typeset details.info {
        border-color: rgba(128, 128, 128, 0.33);
        &:focus-within {
            box-shadow: 0 0 0 .2rem #448aff1a;
        }
        & summary {
            background: transparent;
            &::before {
                color: #227399a9;
                background-color: #227399a9;
                outline: transparent;
            }
            &::before:focus,
            &::before:focus-visible {
                outline: transparent;
                box-shadow: transparent;
            }
            &::after {
                color: var(--md-default-fg-color);
            }
        }
    }
    .toc-anchor {
        position: absolute;
        opacity: 0;
        overflow: hidden;
        width: 0;
        height: 0;
        padding: 0;
        margin: 0 !important;
        pointer-events: none;
    }
 </style>

 <script>
(()=>{
    const open=(hash)=>{
        if(!hash||hash==="#")return;
        const h=document.getElementById(hash.slice(1));
        let n=h?.nextElementSibling;
        while(n){
            if(/^H[1-6]$/.test(n.tagName)) break;
            if(n.tagName==="DETAILS"){n.open=true;break;}
            n=n.nextElementSibling;
        }
    };
    let last="";
    const tick=()=>{const h=location.hash;if(h!==last){last=h;open(h);}requestAnimationFrame(tick);};
    requestAnimationFrame(tick);
})();
</script>

VIIPER provides a lightweight TCP API for managing and controlling virtual buses/devices.  
It's designed to be trivial to drive from any language that can open a TCP socket and send null-byte-terminated payloads.

!!! tip "Client Libraries Available"
    Client libraries are available that abstract away any protocol details described below.  
    For most use cases, you should use one of the provided client libraries rather than implementing the raw protocol yourself:

    - [Go Client](../clients/go.md): Reference implementation included in the repository
    - [Generator Documentation](../clients/generator.md): Information about code generation
    - [C++ Client Library](../clients/cpp.md): Header-only C++20 library (requires external JSON parser)
    - [C# Client Library](../clients/csharp.md): Generated .NET library with async/await support
    - [TypeScript Client Library](../clients/typescript.md): Generated Node.js library with EventEmitter streams
    - [Rust Client Library](../clients/rust.md): Generated Rust library with sync/async support

    
    The documentation below is provided for reference and for implementing clients in languages not supported officially.

## Protocol overview

The TCP API is inspired by the ubiquitous HTTP REST style, but is more lightweight.  
If you ever worked with HTTP APIs before, you'll feel right at home.  
The exception to this are the device-control and feedback streams, which are raw binary streams specific to each device type.

- **Transport**: TCP with optional encryption (ChaCha20-Poly1305)
- **Default listen address**: `:3242` (configurable via `--api.addr`)
- **Authentication**: Required for remote connections, optional for localhost (password-based with HMAC validation)
- **Encryption**: Automatic for authenticated connections (ChaCha20-Poly1305 with unique session keys)
- **Request format**: a single ASCII/UTF‑8 line terminated by `\0`
- **Routing**: path followed by optional payload separated by whitespace  
  (e.g., `bus/list\0` or `bus/create 5\0`)
- **Payload**: optional string that can be a JSON object, numeric value, or plain string depending on the endpoint.  
  The payload may contain newlines (e.g., pretty-printed JSON) as only the null byte terminates the request.
- **Success response**: a single line containing a JSON payload (or an empty line for commands that have no payload), terminated by connection close
- **Error response**: a single line JSON object following RFC 7807 Problem Details format with a `status` field (HTTP-style status code) and other error details, terminated by connection close

!!! tip "Testing the API"
    For quick testing, you can use tools like `netcat` (Linux/macOS) or PowerShell scripts (Windows) to send requests and read responses.

!!! warning "Connection timing and auto‑cleanup"
    After you add a device with `bus/{id}/add`, you must connect to its streaming endpoint within the configured `DeviceHandlerConnectTimeout` (default: 5s). If no stream connection is established in time, the device is automatically removed. Likewise, when a stream disconnects, a reconnection timer with the same timeout starts; if the client doesn’t reconnect before it expires, the device is removed.

!!! warning "Authentication Required for Remote Connections"
    **VIIPER requires authentication for all non-localhost connections.**  

    - **Localhost clients** (`127.0.0.1`, `::1`, `localhost`): Authentication is **optional** (but supported) by default
    - **Remote clients**: Authentication is **required** and enforced
    
    On first start, VIIPER generates a random password
    and saves it to `<USER_CONFIG_DIR>/viiper.key.txt`.  
    Windows: `%APPDATA%\VIIPER\viiper.key.txt`  
    Linux (user): `~/.config/github.com/Alia5/viiper/viiper.key.txt`  
    Linux (root/systemd): `/etc/viiper/viiper.key.txt`

    Remote clients must provide this password to establish a connection.  

    See the [Configuration](../cli/configuration.md) documentation for details on password management and the `--api.require-localhost-auth` option.

## Endpoints

!!! info "null byte excluded"
    The `\0` (null byte) terminator is excluded from all examples below for readability.  
    All requests must be terminated with a null byte (unless otherwise noted).

<div class="grid cards" markdown>

- **Bus Management**
  
    ---

    Create, list, and remove virtual buses

    [Jump to section](#bus-management)

- **Device Management**
  
    ---

    Add, list, and remove devices on a bus

    [Jump to section](#device-management)

- **Device Control / Feedback**
  
    ---

    Real-time input and feedback streams for devices

    [Jump to section](#device-control--feedback)

- **Error Handling**
  
    ---

    Error response format and common error codes

    [Jump to section](#error-handling)

</div>

### Bus Management {#bus-management}

#### `ping` {.toc-anchor}

??? info "ping - Simple identity and version check"
    **Request:** `ping`

    **Response:** `{ "server": "VIIPER", "version": "1.2.3[-dev-abcd]" }`

#### `bus/list` {.toc-anchor}

??? info "bus/list - List all virtual bus IDs"
    **Request:** `bus/list`

    **Response:** `{ "buses": [1, 2, ...] }`

#### `bus/create [busId]` {.toc-anchor}

??? info "bus/create - Create a new bus"
    **Request:** `bus/create` or `bus/create 5`

    **Payload:** Optional numeric bus ID (e.g., `5`)  
    If provided, VIIPER attempts to create the bus with that id; otherwise it picks the next free id.
    
    **Response:** `{ "busId": <id> }`

#### `bus/remove <busId>` {.toc-anchor}

??? info "bus/remove - Remove a bus and all devices on it"
    **Request:** `bus/remove 1`

    **Payload:** Numeric bus ID (e.g., `1`)
    
    **Response:** `{ "busId": <id> }`

### Device Management {#device-management}

#### `bus/{id}/list` {.toc-anchor}

??? info "bus/{id}/list - List devices on a bus"
    **Request:** `bus/1/list`

    **Response:** 
    ```json
    {
      "devices": [
        {
          "busId": 1,
          "devId": "1",
          "vid": "0x045e",
          "pid": "0x028e",
          "type": "xbox360"
          "deviceSpecific": {
            "subType": 1
          }
        }
      ]
    }
    ```

#### `bus/{id}/add <json_payload>` {.toc-anchor}

??? info "bus/{id}/add - Add a device to a bus"
    **Request:** `bus/1/add {"type":"xbox360"}`

    **Payload:** JSON object with device creation parameters
    ```json
    {
      "type": "<deviceType>",
      "idVendor": <optional_vid>,
      "idProduct": <optional_pid>,
      "deviceSpecific": <optional device specific args>
    }
    ```
    
    **Examples:**
    - `{"type":"xbox360"}`
    - `{"type":"keyboard","idVendor":1234,"idProduct":5678}`
    - `{"type":"xbox360", "deviceSpecific": {"subType": 7}}`
    
    **Response:**
    ```json
    {
      "busId": 1,
      "devId": "1",
      "vid": "0x045e",
      "pid": "0x028e",
      "type": "xbox360",
      "deviceSpecific": {
        "subType":7
      }
    }
    ```
    
    !!! warning "Connection timeout"
        After add, the server starts a connect timer (default `5s`). You must open a device stream before the timeout expires, otherwise the device is auto-removed.
    
    !!! info "Auto-attach"
        If [auto-attach](../cli/server.md#api.auto-attach-local-client) is enabled (default), the server automatically attaches the new device to a local USBIP client on the same host (localhost only). Failures are logged but do not affect the API response.

#### `bus/{id}/remove <deviceId>` {.toc-anchor}

??? info "bus/{id}/remove - Remove a device from a bus"
    **Request:** `bus/1/remove 1`

    **Payload:** Numeric device ID (e.g., `1` for device 1-1 on the bus)
    
    **Response:** `{ "busId": <id>, "devId": "<dev>" }`

### Device Control / Feedback {#device-control--feedback}

Device Control and Feedback requires an initial "handshake" request, afterwards the connection is used as a long-lived (device-specific, binary) bidirectional stream.

!!! info "Establish control/feedback connection/stream"
    **Path:** `bus/{busId}/{deviceId}`

    **Handshake:** Send the path followed by `\0` (null byte)  
    Example: `bus/1/1\0`
    
    **Type:** Long-lived TCP connection
    
    **Purpose:** Device-specific, bidirectional stream.  
    
    !!! warning "Timeout behavior"
        When a stream ends, a reconnect timer is started.  
        If the client doesn't reconnect in time, the device is removed.

Device control and feedback is **device-specific**.  
Each device type defines it's own packet formats.  

**In general** the client (your code) sends sends binary input state packets to the VIIPER server and possibly receives binary feedback packets (rumble, keyboard leds, etc.) back.

Refer to the individual [device documentation](../devices/overview.md) for details on packet formats and behavior.

### Error Handling {#error-handling}

All errors are inspired by HTTP REST APIs and are returned as single-line JSON objects in the style of [RFC 7807 Problem Details](https://tools.ietf.org/html/rfc7807).  
The connection closes immediately after the error response.  

If you have ever worked with HTTP APIs, the errors and status codes will feel familiar.

#### Error Response Format

```json
{
  "status": 400,
  "title": "Bad Request",
  "detail": "missing payload"
}
```

**Fields:**

- `status` (number): HTTP-style status code indicating the error type
- `title` (string): Short, human-readable summary of the problem
- `detail` (string): Explanation specific to this occurrence

#### Common Error Codes

Error codes are basically HTTP carbon copies:

| Status | Title | Cause | Example |
|--------|-------|-------|---------|
| 400 | Bad Request | Invalid request format, missing payload, or invalid JSON | Missing device type in `bus/{id}/add`, invalid busId format |
| 404 | Not Found | Resource does not exist | Bus ID not found, device ID not found |
| 409 | Conflict | Resource already exists or cannot be modified | Bus ID already exists, auto-attach failure |
| 500 | Internal Server Error | (Unhandled) Server-side error during operation | Failed to marshal response, device add failure, unknown error |

## Example sessions

=== "PowerShell (Windows)"

    VIIPER includes convenience scripts for quick testing and automation:

    ```powershell
    # Source the script to load helper functions
    . .\scripts\viiper-api.ps1

    # Use the Invoke-ViiperApi function (or 'viiper' alias)
    viiper "bus/list"
    viiper "bus/create"
    viiper "bus/1/add {\"type\":\"xbox360\"}" -Port 3242 -Hostname localhost
    ```

    The script provides `Invoke-ViiperApi` (alias: `viiper`) for sending commands and `Connect-ViiperDevice` for testing persistent device connections.

=== "netcat (Linux/macOS)""

    ```bash
    # List buses
    printf "bus/list\0" | nc localhost 3242

    # Create a bus
    printf "bus/create\0" | nc localhost 3242
    # → {"busId":1}

    # Create a bus with specific ID
    printf "bus/create 5\0" | nc localhost 3242
    # → {"busId":5}

    # Add a virtual Xbox 360 controller to bus 1
    printf 'bus/1/add {"type":"xbox360"}\0' | nc localhost 3242
    # → {"busId":1,"devId":"1","vid":"0x045e","pid":"0x028e","type":"xbox360"}

    # List devices on bus 1
    printf "bus/1/list\0" | nc localhost 3242
    ```

Then, open a second TCP connection for device control to `bus/1/1` (the API port, not the USBIP port).  
First send the "handshake" `bus/1/1\0`,  
then you'll write device-specific input packets and read device-specific feedback packets.  
Any language with raw TCP support works.

### Go snippet (raw TCP Socket)

```go
package main

import (
    "fmt"
    "io"
    "net"
)

func main() {
    conn, _ := net.Dial("tcp", "localhost:3242")
    defer conn.Close()
    
    // Send request with null terminator
    fmt.Fprint(conn, "bus/create\x00")
    
    // Read entire response until connection closes
    resp, _ := io.ReadAll(conn)
    fmt.Println(string(resp)) // {"busId":1}\n
}
```

For a higher-level experience, see the Go client in `/apiclient/`.

## How this relates to USBIP

The VIIPER API controls which virtual devices exist and exposes a device stream for live input/feedback.  
Separately, the USBIP server (default `:3241`) makes these devices attachable from USBIP clients.

Typical flow:

1. Create a bus
2. Add a device
3. Connect the device stream
4. Attach a USBIP client to the device

If auto-attach is enabled step 4 is attempted automatically for the local host; you still must perform step 3 to keep the device alive.
