# C++ SDK Documentation

The VIIPER C++ SDK provides a modern, header-only C++20 client library for interacting with VIIPER servers and controlling virtual devices.

## Overview

The C++ SDK features:

- **Header-only**: No separate compilation required, just include and use
- **Modern C++20**: Uses concepts, designated initializers, std::optional, smart pointers
- **Type-safe**: Generated structs with constants and helper maps
- **Callback-based output**: Register lambdas for device feedback (LEDs, rumble)
- **Thread-safe**: Separate mutexes for send/recv operations
- **Cross-platform**: Windows (MSVC) and POSIX (GCC/Clang)

!!! warning "JSON Parser Required"
    The C++ SDK requires a JSON library to be provided by the user. You **must** define `VIIPER_JSON_INCLUDE`, `VIIPER_JSON_NAMESPACE`, and `VIIPER_JSON_TYPE` before including the SDK headers.
    
    **Recommended**: [nlohmann/json](https://github.com/nlohmann/json) - a header-only JSON library that can be easily integrated.

!!! note "License"
    The C++ SDK is licensed under the **MIT License**, providing maximum flexibility for integration into your projects.  
    The core VIIPER server remains under its original license.

## Installation

### 1. Header-Only Integration

Copy the `clients/cpp/include/viiper` directory to your project's include path:

```bash
cp -r clients/cpp/include/viiper /path/to/your/project/include/
```

Or add it as an include directory in your build system.

### 2. CMake Integration

```cmake
# Add viiper include directory
target_include_directories(your_target PRIVATE path/to/clients/cpp/include)

# Also ensure nlohmann/json is available
# Option A: FetchContent
include(FetchContent)
FetchContent_Declare(json
    GIT_REPOSITORY https://github.com/nlohmann/json.git
    GIT_TAG v3.11.3
)
FetchContent_MakeAvailable(json)
target_link_libraries(your_target PRIVATE nlohmann_json::nlohmann_json)

# Option B: Find package (if installed system-wide)
find_package(nlohmann_json REQUIRED)
target_link_libraries(your_target PRIVATE nlohmann_json::nlohmann_json)
```

### 3. Generating from Source

```bash
go run ./cmd/viiper codegen --lang=cpp
```

The SDK will be generated in `clients/cpp/include/viiper/`.

## JSON Parser Configuration

Before including the VIIPER SDK, you must configure a JSON parser. The SDK is designed to work with any JSON library that provides a compatible interface.

### Using nlohmann/json (Recommended)

```cpp
// Define these BEFORE including viiper headers
#define VIIPER_JSON_INCLUDE <nlohmann/json.hpp>
#define VIIPER_JSON_NAMESPACE nlohmann
#define VIIPER_JSON_TYPE json

#include <viiper/viiper.hpp>
```

### Using a Custom JSON Library

Your JSON type must support:

- `parse(const std::string&)` → JsonType
- `dump()` → std::string
- `operator[](const std::string&)` → JsonType
- `contains(const std::string&)` → bool
- `is_number()`, `is_string()`, `is_array()`, `is_object()` → bool
- `get<T>()` → T
- `size()` → std::size_t (for arrays)

Example with a custom library:

```cpp
#define VIIPER_JSON_INCLUDE "my_json_lib.hpp"
#define VIIPER_JSON_NAMESPACE myjson
#define VIIPER_JSON_TYPE JsonValue

#include <viiper/viiper.hpp>
```

## Quick Start

```cpp
#define VIIPER_JSON_INCLUDE <nlohmann/json.hpp>
#define VIIPER_JSON_NAMESPACE nlohmann
#define VIIPER_JSON_TYPE json

#include <viiper/viiper.hpp>
#include <iostream>

int main() {
    // Create new Viiper client
    viiper::ViiperClient client("localhost", 3242);

    // Find or create a bus
    auto buses_result = client.buslist();
    if (buses_result.is_error()) {
        std::cerr << "BusList error: " << buses_result.error().to_string() << "\n";
        return 1;
    }

    std::uint32_t bus_id;
    if (buses_result.value().buses.empty()) {
        auto create_result = client.buscreate(std::nullopt);  // Auto-assign ID
        if (create_result.is_error()) {
            std::cerr << "BusCreate error: " << create_result.error().to_string() << "\n";
            return 1;
        }
        bus_id = create_result.value().busid;
    } else {
        bus_id = buses_result.value().buses[0];
    }

    // Add device
    auto device_result = client.busdeviceadd(bus_id, {.type = "keyboard"});
    if (device_result.is_error()) {
        std::cerr << "AddDevice error: " << device_result.error().to_string() << "\n";
        return 1;
    }
    auto device_info = std::move(device_result.value());

    // Connect to device stream
    auto stream_result = client.connectDevice(device_info.busid, device_info.devid);
    if (stream_result.is_error()) {
        std::cerr << "Connect error: " << stream_result.error().to_string() << "\n";
        return 1;
    }
    auto stream = std::move(stream_result.value());

    std::cout << "Connected to device " << device_info.devid
              << " on bus " << device_info.busid << "\n";

    // Send keyboard input
    viiper::keyboard::Input input = {
        .modifiers = viiper::keyboard::ModLeftShift,
        .keys = {viiper::keyboard::KeyH},
    };
    stream->send(input);

    // Cleanup
    client.busdeviceremove(device_info.busid, device_info.devid);

    return 0;
}
```

## Device Stream API

### Creating a Device Stream

The simplest way to add a device and connect:

```cpp
auto [device_info, stream] = client.addDeviceAndConnect(bus_id, {.type = "xbox360"}).value();
```

With custom VID/PID:

```cpp
viiper::Devicecreaterequest req = {
    .type = "keyboard",
    .idvendor = 0x1234,
    .idproduct = 0x5678
};
auto [device_info, stream] = client.addDeviceAndConnect(bus_id, req).value();
```

Or manually add and connect:

```cpp
auto device_result = client.busdeviceadd(bus_id, {.type = "keyboard"});
auto device_info = device_result.value();

auto stream_result = client.connectDevice(device_info.busid, device_info.devid);
auto stream = std::move(stream_result.value());
```

### Sending Input

Device input is sent using generated structs:

**Keyboard:**

```cpp
viiper::keyboard::Input input = {
    .modifiers = viiper::keyboard::ModLeftShift,
    .keys = {viiper::keyboard::KeyH, viiper::keyboard::KeyE},
};
stream->send(input);
```

**Mouse:**

```cpp
viiper::mouse::Input input = {
    .buttons = viiper::mouse::ButtonLeft,
    .x = 10,
    .y = -5,
    .wheel = 0,
};
stream->send(input);
```

**Xbox360 Controller:**

```cpp
viiper::xbox360::Input input = {
    .buttons = viiper::xbox360::ButtonA,
    .lt = 255,           // Left trigger (0-255)
    .rt = 0,             // Right trigger (0-255)
    .lx = -32768,        // Left stick X (-32768 to 32767)
    .ly = 32767,         // Left stick Y
    .rx = 0,             // Right stick X
    .ry = 0,             // Right stick Y
};
stream->send(input);
```

### Receiving Output (Callbacks)

For devices that send feedback (rumble, LEDs), register a callback:

**Keyboard LEDs:**

```cpp
stream->on_output(viiper::keyboard::OUTPUT_SIZE, [](const std::uint8_t* data, std::size_t len) {
    if (len < viiper::keyboard::OUTPUT_SIZE) return;
    auto result = viiper::keyboard::Output::from_bytes(data, len);
    if (result.is_error()) return;
    
    auto& leds = result.value();
    bool num_lock = (leds.leds & viiper::keyboard::LEDNumLock) != 0;
    bool caps_lock = (leds.leds & viiper::keyboard::LEDCapsLock) != 0;
    std::cout << "LEDs: Num=" << num_lock << " Caps=" << caps_lock << "\n";
});
```

**Xbox360 Rumble:**

```cpp
stream->on_output(viiper::xbox360::OUTPUT_SIZE, [](const std::uint8_t* data, std::size_t len) {
    if (len < viiper::xbox360::OUTPUT_SIZE) return;
    auto result = viiper::xbox360::Output::from_bytes(data, len);
    if (result.is_error()) return;
    
    auto& rumble = result.value();
    std::cout << "Rumble: Left=" << static_cast<int>(rumble.left)
              << ", Right=" << static_cast<int>(rumble.right) << "\n";
});
```

### Event Handlers

```cpp
// Called when the server disconnects the device
stream->on_disconnect([]() {
    std::cerr << "Device disconnected by server\n";
});

// Called on stream errors
stream->on_error([](const viiper::Error& err) {
    std::cerr << "Stream error: " << err.to_string() << "\n";
});
```

### Stopping a Device

```cpp
stream->stop();  // Stops the output thread and closes the connection
```

The device is also automatically stopped when the `ViiperDevice` is destroyed.

## Generated Constants and Maps

The C++ SDK automatically generates constants and helper maps for each device type.

### Keyboard Constants

**Key Codes:**

```cpp
auto key = viiper::keyboard::KeyA;           // 0x04
auto f1 = viiper::keyboard::KeyF1;           // 0x3A
auto enter = viiper::keyboard::KeyEnter;     // 0x28
```

**Modifier Flags:**

```cpp
std::uint8_t mods = viiper::keyboard::ModLeftShift | viiper::keyboard::ModLeftCtrl;
```

**LED Flags:**

```cpp
bool num_lock = (leds & viiper::keyboard::LEDNumLock) != 0;
bool caps_lock = (leds & viiper::keyboard::LEDCapsLock) != 0;
```

### Helper Maps

The SDK generates useful lookup maps for working with keyboard input:

**CHAR_TO_KEY Map** - Convert ASCII characters to key codes:

```cpp
auto it = viiper::keyboard::CHAR_TO_KEY.find(static_cast<std::uint8_t>('a'));
if (it != viiper::keyboard::CHAR_TO_KEY.end()) {
    std::uint8_t key = it->second;  // KeyA
}
```

**KEY_NAME Array** - Get human-readable key names:

```cpp
for (const auto& [key, name] : viiper::keyboard::KEY_NAME) {
    if (key == viiper::keyboard::KeyF1) {
        std::cout << "Key name: " << name << "\n";  // "F1"
        break;
    }
}
```

**SHIFT_CHARS Set** - Check if a character requires shift:

```cpp
bool needs_shift = viiper::keyboard::SHIFT_CHARS.contains(static_cast<std::uint8_t>('A'));
```

### Xbox360 Constants

**Button Flags:**

```cpp
std::uint16_t buttons = viiper::xbox360::ButtonA | viiper::xbox360::ButtonB;
```

**All Button Constants:**

```cpp
viiper::xbox360::ButtonDPadUp
viiper::xbox360::ButtonDPadDown
viiper::xbox360::ButtonDPadLeft
viiper::xbox360::ButtonDPadRight
viiper::xbox360::ButtonStart
viiper::xbox360::ButtonBack
viiper::xbox360::ButtonLThumb
viiper::xbox360::ButtonRThumb
viiper::xbox360::ButtonLShoulder
viiper::xbox360::ButtonRShoulder
viiper::xbox360::ButtonGuide
viiper::xbox360::ButtonA
viiper::xbox360::ButtonB
viiper::xbox360::ButtonX
viiper::xbox360::ButtonY
```

## Practical Example: Typing Text

Using the generated maps to type a string:

```cpp
void type_string(viiper::ViiperDevice& stream, const std::string& text) {
    for (char ch : text) {
        auto it = viiper::keyboard::CHAR_TO_KEY.find(static_cast<std::uint8_t>(ch));
        if (it == viiper::keyboard::CHAR_TO_KEY.end()) continue;
        std::uint8_t key = it->second;

        std::uint8_t mods = 0;
        if (viiper::keyboard::SHIFT_CHARS.contains(static_cast<std::uint8_t>(ch))) {
            mods = viiper::keyboard::ModLeftShift;
        }

        // Press key
        viiper::keyboard::Input down = {
            .modifiers = mods,
            .keys = {key},
        };
        stream.send(down);
        std::this_thread::sleep_for(std::chrono::milliseconds(50));

        // Release key
        viiper::keyboard::Input up = {
            .modifiers = 0,
            .keys = {},
        };
        stream.send(up);
        std::this_thread::sleep_for(std::chrono::milliseconds(50));
    }
}

// Usage
type_string(*stream, "Hello, World!");
```

## Error Handling

All API methods return `Result<T>`, which is either a value or an error:

```cpp
auto result = client.buslist();
if (result.is_error()) {
    std::cerr << "Error: " << result.error().to_string() << "\n";
    return 1;
}
auto buses = result.value();
```

Using the value directly (throws on error):

```cpp
// Only use if you're certain the operation succeeded
auto buses = client.buslist().value();
```

### Resource Management

`ViiperDevice` is managed via `std::unique_ptr` and automatically cleans up:

```cpp
{
    auto stream = client.connectDevice(bus_id, device_id).value();
    // ... use stream ...
}  // stream->stop() called automatically
```

## Examples

Full working examples are available in the repository:

- **Virtual Keyboard**: `examples/cpp/virtual_keyboard.cpp`
  - Types "Hello!" every 5 seconds using generated maps
  - Displays LED feedback in console
  
- **Virtual Mouse**: `examples/cpp/virtual_mouse.cpp`
  - Moves cursor in a circle pattern
  - Demonstrates button clicks

- **Virtual Xbox360 Controller**: `examples/cpp/virtual_x360_pad.cpp`
  - Cycles through buttons A, B, X, Y
  - Handles rumble feedback

### Building Examples

```bash
cd examples/cpp
mkdir build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
cmake --build . --config Release
```

### Running Examples

```bash
./virtual_keyboard localhost:3242
./virtual_mouse localhost:3242
./virtual_x360_pad localhost:3242
```

## Project Structure

Generated SDK layout:

```text
clients/cpp/include/viiper/
├── viiper.hpp              # Main include (includes all others)
├── config.hpp              # JSON configuration macros
├── error.hpp               # Result<T> and Error types
├── types.hpp               # API request/response types
├── client.hpp              # ViiperClient management API
├── device.hpp              # ViiperDevice stream wrapper
├── detail/
│   ├── socket.hpp          # Cross-platform socket wrapper
│   └── json.hpp            # JSON parsing helpers
└── devices/
    ├── keyboard.hpp        # Keyboard constants, Input, Output
    ├── mouse.hpp           # Mouse constants, Input
    └── xbox360.hpp         # Xbox360 constants, Input, Output
```

## Requirements

- **C++20** or later
- **JSON library** (nlohmann/json recommended)
- **Platform**: Windows (MSVC 2019+) or POSIX (GCC 10+, Clang 10+)

### Windows-specific

The SDK uses Winsock2 for networking. Link against `Ws2_32.lib` (done automatically via `#pragma comment`).

### POSIX-specific

Standard POSIX sockets are used. No additional libraries required.

## Troubleshooting

**Error: VIIPER_JSON_INCLUDE must be defined**

You must define the JSON macros before including any VIIPER headers:

```cpp
#define VIIPER_JSON_INCLUDE <nlohmann/json.hpp>
#define VIIPER_JSON_NAMESPACE nlohmann
#define VIIPER_JSON_TYPE json

#include <viiper/viiper.hpp>  // Include AFTER the defines
```

**Linker errors on Windows**

Ensure Winsock2 is linked. If not using the auto-link pragma, add:

```cmake
target_link_libraries(your_target PRIVATE Ws2_32)
```

**Connection refused**

Verify the VIIPER server is running:

```bash
viiper server --api-addr localhost:3242
```

## See Also

- [Generator Documentation](generator.md): How generated SDKs work
- [C SDK Documentation](c.md): Pure C alternative
- [Rust SDK Documentation](rust.md): Rust SDK with sync/async support
- [C# SDK Documentation](csharp.md): .NET SDK
- [TypeScript SDK Documentation](typescript.md): Node.js SDK
- [API Overview](../api/overview.md): Management API reference
- [Device Documentation](../devices/): Wire formats and device-specific details

---

For questions or contributions, see the main VIIPER repository.
