# Rust Client Library Documentation

The VIIPER Rust client library provides a type-safe, zero-cost abstraction client library for interacting with VIIPER servers and controlling virtual devices.

## Overview

The Rust client library features:

- **Sync and Async APIs**: Choose between blocking `ViiperClient` or async `AsyncViiperClient` (with `async` feature)
- **Type-safe**: Generated structs with constants, helper maps, and `DeviceInput` trait implementations
- **Callback-based output**: Register closures for device feedback (LEDs, rumble)
- **Zero external dependencies** (sync): Uses only `std` for the synchronous client
- **Tokio-based async**: Optional `async` feature for async/await support with Tokio runtime

!!! note "License"
    The Rust client library is licensed under the **MIT License**, providing maximum flexibility for integration into your projects.  
    The core VIIPER server remains under its original license.

## Installation

### 1. Using the Published Crate (Recommended)

Install the client library using Cargo:

```bash
cargo add viiper-client
```

For async support:

```bash
cargo add viiper-client --features async
cargo add tokio --features full
```

Package page: [viiper-client on crates.io](https://crates.io/crates/viiper-client)

> Pre-release / snapshot builds are **not** published to crates.io. They are only available as GitHub Release artifacts (e.g. `dev-latest`) or by building from source.

### 2. Path Dependency (For Local Development Against Source)

Use this when modifying the generator or contributing new device types:

```toml
[dependencies]
viiper-client = { path = "../../clients/rust" }
```

### 3. Generating from Source (Advanced / Contributors)

Only required when enhancing VIIPER itself:

```bash
go run ./cmd/viiper codegen --lang=rust
cd clients/rust
cargo build --release
```

## Quick Start (Sync)

```rust
use viiper_client::{ViiperClient, devices::keyboard::*};
use std::net::ToSocketAddrs;

fn main() {
    // Create new Viiper client
    let addr = "localhost:3242"
        .to_socket_addrs()
        .expect("Invalid address")
        .next()
        .expect("No address resolved");
    let client = ViiperClient::new(addr);

    // Find or create a bus
    let bus_id = match client.bus_list() {
        Ok(resp) if resp.buses.is_empty() => {
            client.bus_create(None).expect("Failed to create bus").bus_id
        }
        Ok(resp) => *resp.buses.first().unwrap(),
        Err(e) => panic!("BusList error: {}", e),
    };

    // Add device
    let device_info = client.bus_device_add(
        bus_id,
        &viiper_client::types::DeviceCreateRequest {
            r#type: Some("keyboard".to_string()),
            id_vendor: None,
            id_product: None,
        },
    ).expect("Failed to add device");

    // Connect to device stream
    let mut stream = client
        .connect_device(device_info.bus_id, &device_info.dev_id)
        .expect("Failed to connect");

    println!("Connected to device {} on bus {}", device_info.dev_id, device_info.bus_id);

    // Send keyboard input
    let input = KeyboardInput {
        modifiers: MOD_LEFT_SHIFT,
        count: 1,
        keys: vec![KEY_H],
    };
    stream.send(&input).expect("Failed to send input");

    // Cleanup
    let _ = client.bus_device_remove(device_info.bus_id, Some(&device_info.dev_id));
}
```

## Quick Start (Async)

```rust
use tokio::time::{sleep, Duration};
use viiper_client::{AsyncViiperClient, devices::keyboard::*};
use std::net::ToSocketAddrs;

#[tokio::main]
async fn main() {
    // Create new Viiper client
    let addr = "localhost:3242"
        .to_socket_addrs()
        .expect("Invalid address")
        .next()
        .expect("No address resolved");
    let client = AsyncViiperClient::new(addr);

    // Find or create a bus
    let bus_id = match client.bus_list().await {
        Ok(resp) if resp.buses.is_empty() => {
            client.bus_create(None).await.expect("Failed to create bus").bus_id
        }
        Ok(resp) => *resp.buses.first().unwrap(),
        Err(e) => panic!("BusList error: {}", e),
    };

    // Add device
    let device_info = client.bus_device_add(
        bus_id,
        &viiper_client::types::DeviceCreateRequest {
            r#type: Some("keyboard".to_string()),
            id_vendor: None,
            id_product: None,
        },
    ).await.expect("Failed to add device");

    // Connect to device stream
    let mut stream = client
        .connect_device(device_info.bus_id, &device_info.dev_id)
        .await
        .expect("Failed to connect");

    println!("Connected to device {} on bus {}", device_info.dev_id, device_info.bus_id);

    // Send keyboard input
    let input = KeyboardInput {
        modifiers: MOD_LEFT_SHIFT,
        count: 1,
        keys: vec![KEY_H],
    };
    stream.send(&input).await.expect("Failed to send input");

    // Cleanup
    let _ = client.bus_device_remove(device_info.bus_id, Some(&device_info.dev_id)).await;
}
```

## Device Stream API

### Creating a Device Stream (Sync)

```rust
use viiper_client::{ViiperClient, types::DeviceCreateRequest};
use std::net::ToSocketAddrs;

let addr = "localhost:3242"
    .to_socket_addrs()
    .expect("Invalid address")
    .next()
    .expect("No address resolved");
let client = ViiperClient::new(addr);

// Add device first
let device_info = client.bus_device_add(
    bus_id,
    &DeviceCreateRequest {
        r#type: Some("xbox360".to_string()),
        id_vendor: None,
        id_product: None,
    },
).expect("Failed to add device");

// Then connect to its stream
let mut stream = client
    .connect_device(device_info.bus_id, &device_info.dev_id)
    .expect("Failed to connect");
```

With custom VID/PID:

```rust
let device_info = client.bus_device_add(
    bus_id,
    &DeviceCreateRequest {
        r#type: Some("keyboard".to_string()),
        id_vendor: Some(0x1234),
        id_product: Some(0x5678),
    },
).expect("Failed to add device");
```

### Sending Input

Device input is sent using generated structs that implement the `DeviceInput` trait:

```rust
use viiper_client::devices::xbox360::*;

let input = Xbox360Input {
    buttons: BUTTON_A as u32,
    lt: 255,
    rt: 0,
    lx: -32768,  // Left stick left
    ly: 32767,   // Left stick up
    rx: 0,
    ry: 0,
};
stream.send(&input).expect("Failed to send");
```

### Receiving Output (Callbacks)

For devices that send feedback (rumble, LEDs), register a callback with `on_output`:

**Sync API:**

```rust
use viiper_client::devices::keyboard::OUTPUT_SIZE;

stream.on_output(|reader| {
    let mut buf = [0u8; OUTPUT_SIZE];
    reader.read_exact(&mut buf)?;
    let leds = buf[0];
    
    let num_lock = (leds & 0x01) != 0;
    let caps_lock = (leds & 0x02) != 0;
    let scroll_lock = (leds & 0x04) != 0;
    
    println!("LEDs: Num={} Caps={} Scroll={}", num_lock, caps_lock, scroll_lock);
    Ok(())
}).expect("Failed to register callback");
```

**Async API:**

```rust
use tokio::io::AsyncReadExt;
use viiper_client::devices::keyboard::OUTPUT_SIZE;

stream.on_output(|stream| async move {
    let mut buf = [0u8; OUTPUT_SIZE];
    let mut guard = stream.lock().await;
    guard.read_exact(&mut buf).await?;
    drop(guard);
    
    let leds = buf[0];
    let num_lock = (leds & 0x01) != 0;
    let caps_lock = (leds & 0x02) != 0;
    
    println!("LEDs: Num={} Caps={}", num_lock, caps_lock);
    Ok(())
}).expect("Failed to register callback");
```

For Xbox360 rumble:

```rust
stream.on_output(|reader| {
    let mut buf = [0u8; 2];
    reader.read_exact(&mut buf)?;
    let left_motor = buf[0];
    let right_motor = buf[1];
    println!("Rumble: Left={} Right={}", left_motor, right_motor);
    Ok(())
}).expect("Failed to register callback");
```

## Generated Constants and Maps

The Rust client library generates constants and lazy-static maps for each device type.

### Keyboard Constants

**Key Constants:**

```rust
use viiper_client::devices::keyboard::*;

let key = KEY_A;           // 0x04
let f1 = KEY_F1;           // 0x3A
let enter = KEY_ENTER;     // 0x28
```

**Modifier Flags:**

```rust
use viiper_client::devices::keyboard::*;

let mods = MOD_LEFT_SHIFT | MOD_LEFT_CTRL;  // 0x03
```

**LED Flags:**

```rust
use viiper_client::devices::keyboard::*;

let num_lock = (leds & LED_NUM_LOCK) != 0;
let caps_lock = (leds & LED_CAPS_LOCK) != 0;
```

### Helper Maps

The client library generates useful lookup maps for working with keyboard input:

**CHAR_TO_KEY** - Convert ASCII characters to key codes:

```rust
use viiper_client::devices::keyboard::CHAR_TO_KEY;

if let Some(&key) = CHAR_TO_KEY.get(&b'a') {
    println!("'a' maps to key code {}", key);  // KEY_A
}
```

**KEY_NAME** - Get human-readable key names:

```rust
use viiper_client::devices::keyboard::KEY_NAME;

if let Some(name) = KEY_NAME.get(&KEY_F1) {
    println!("Key name: {}", name);  // "F1"
}
```

**SHIFT_CHARS** - Check if a character requires shift:

```rust
use viiper_client::devices::keyboard::SHIFT_CHARS;

let needs_shift = SHIFT_CHARS.contains(&b'A');  // true for uppercase
```

## Practical Example: Typing Text

Using the generated maps to type a string:

```rust
use std::thread;
use std::time::Duration;
use viiper_client::{DeviceStream, devices::keyboard::*};

fn type_string(stream: &mut DeviceStream, text: &str) -> Result<(), viiper_client::ViiperError> {
    for ch in text.chars() {
        let byte = ch as u8;
        let key = match CHAR_TO_KEY.get(&byte) {
            Some(&k) => k,
            None => continue,
        };

        let mods = if SHIFT_CHARS.contains(&byte) {
            MOD_LEFT_SHIFT
        } else {
            0
        };

        // Press
        let down = KeyboardInput {
            modifiers: mods,
            count: 1,
            keys: vec![key],
        };
        stream.send(&down)?;
        thread::sleep(Duration::from_millis(50));

        // Release
        let up = KeyboardInput {
            modifiers: 0,
            count: 0,
            keys: vec![],
        };
        stream.send(&up)?;
        thread::sleep(Duration::from_millis(50));
    }
    Ok(())
}

// Usage
type_string(&mut stream, "Hello, World!")?;
```

## Device-Specific Wire Formats

### Keyboard Input

```rust
pub struct KeyboardInput {
    pub modifiers: u8,    // Modifier flags (Ctrl, Shift, Alt, GUI)
    pub count: u8,        // Number of keys in keys vec
    pub keys: Vec<u8>,    // Key codes (max 6 for HID compliance)
}
```

**Wire format:** 1 byte modifiers + 1 byte count + N bytes keys (variable-length)

### Keyboard Output (LEDs)

```rust
// Single byte with LED flags
let leds = buf[0];
let num_lock = (leds & LED_NUM_LOCK) != 0;
```

### Xbox360 Input

```rust
pub struct Xbox360Input {
    pub buttons: u32,   // Button flags
    pub lt: u8,         // Left trigger (0-255)
    pub rt: u8,         // Right trigger (0-255)
    pub lx: i16,        // Left stick X (-32768 to 32767)
    pub ly: i16,        // Left stick Y (-32768 to 32767)
    pub rx: i16,        // Right stick X (-32768 to 32767)
    pub ry: i16,        // Right stick Y (-32768 to 32767)
}
```

**Wire format:** Fixed 14 bytes, packed structure (little-endian)

### Xbox360 Output (Rumble)

```rust
// Two bytes: left motor + right motor (0-255 each)
let left_motor = buf[0];
let right_motor = buf[1];
```

### Mouse Input

```rust
pub struct MouseInput {
    pub buttons: u8,   // Button flags
    pub dx: i8,        // Relative X movement (-128 to 127)
    pub dy: i8,        // Relative Y movement (-128 to 127)
    pub wheel: i8,     // Vertical scroll
    pub pan: i8,       // Horizontal scroll
}
```

**Wire format:** Fixed 5 bytes, packed structure

## Error Handling

The client library uses a custom `ViiperError` type for all errors:

```rust
use viiper_client::ViiperError;

match client.bus_list() {
    Ok(buses) => println!("Found {} buses", buses.buses.len()),
    Err(ViiperError::Io(e)) => eprintln!("I/O error: {}", e),
    Err(ViiperError::Protocol(problem)) => eprintln!("API error: {}", problem),
    Err(e) => eprintln!("Other error: {}", e),
}
```

The server returns errors as RFC 7807 Problem JSON. The client parses these into `ProblemJson`:

```rust
use viiper_client::ProblemJson;

if let Err(ViiperError::Protocol(problem)) = result {
    println!("Status: {}", problem.status);
    println!("Title: {}", problem.title);
    println!("Detail: {}", problem.detail);
}
```

## Features

The Rust client library supports optional features:

| Feature | Description | Dependencies |
|---------|-------------|--------------|
| (default) | Synchronous blocking client | None |
| `async` | Async client with Tokio runtime | `tokio`, `tokio-util` |

Enable async support:

```toml
[dependencies]
viiper-client = { version = "0.1", features = ["async"] }
```

## Examples

Full working examples are available in the repository:

- **Virtual Keyboard (sync)**: `examples/rust/sync/virtual_keyboard/`
  - Types "Hello!" every 5 seconds using generated maps
  - Displays LED feedback in console
  
- **Virtual Keyboard (async)**: `examples/rust/async/virtual_keyboard/`
  - Async version using Tokio runtime

- **Virtual Mouse (sync/async)**: `examples/rust/sync/virtual_mouse/`, `examples/rust/async/virtual_mouse/`
  - Moves cursor diagonally
  - Demonstrates button clicks and scroll wheel

- **Virtual Xbox360 Controller (sync/async)**: `examples/rust/sync/virtual_x360_pad/`, `examples/rust/async/virtual_x360_pad/`
  - Cycles through buttons
  - Handles rumble feedback

### Running Examples

```bash
cd examples/rust

# Sync examples
cargo run --release -p virtual_keyboard_sync -- localhost:3242
cargo run --release -p virtual_mouse_sync -- localhost:3242
cargo run --release -p virtual_x360_pad_sync -- localhost:3242

# Async examples
cargo run --release -p virtual_keyboard_async -- localhost:3242
cargo run --release -p virtual_mouse_async -- localhost:3242
cargo run --release -p virtual_x360_pad_async -- localhost:3242
```

## Project Structure

Generated client library layout:

```text
clients/rust/
├── Cargo.toml
├── src/
│   ├── lib.rs                 # Re-exports
│   ├── client.rs              # Sync ViiperClient + DeviceStream
│   ├── async_client.rs        # Async ViiperClient (feature = "async")
│   ├── error.rs               # ViiperError, ProblemJson
│   ├── types.rs               # API request/response types
│   ├── wire.rs                # DeviceInput/DeviceOutput traits
│   └── devices/
│       ├── mod.rs
│       ├── keyboard/
│       │   ├── mod.rs
│       │   ├── input.rs       # KeyboardInput struct
│       │   ├── output.rs      # Output parsing
│       │   └── constants.rs   # Keys, mods, LEDs, maps
│       ├── mouse/
│       │   └── ...
│       └── xbox360/
│           └── ...
```

## Troubleshooting

**Connection refused:**

Verify VIIPER server is running and listening on the expected API port (default 3242).

```rust
use std::net::SocketAddr;
let addr: SocketAddr = "127.0.0.1:3242".parse().expect("Invalid address");
let client = ViiperClient::new(addr);
```

**Feature not found errors:**

Make sure to enable the `async` feature if using `AsyncViiperClient`:

```toml
viiper-client = { version = "0.1", features = ["async"] }
```

## See Also

- [Generator Documentation](generator.md): How generated client libraries work
- [Go Client Documentation](go.md): Reference implementation patterns
- [C# Client Library Documentation](csharp.md): Alternative managed language client library
- [TypeScript Client Library Documentation](typescript.md): Node.js client library
- [C Client Library Documentation](c.md): Native C client library
- [C++ Client Library Documentation](cpp.md): Header-only C++ client library
- [API Overview](../api/overview.md): Management API reference
- [Device Documentation](../devices/): Wire formats and device-specific details

---

For questions or contributions, see the main VIIPER repository.
