# C SDK Documentation

The VIIPER C SDK provides a lightweight, dependency-free client library for interacting with VIIPER servers and controlling virtual devices.

## Overview

The C SDK features:

- **Device-agnostic streaming API**: Uniform interface for all device types
- **Zero dependencies**: Pure C99, no external libraries required
- **Cross-platform**: Windows (MSVC) and POSIX (GCC/Clang)
- **Type-safe**: Generated headers with packed structs and constants
- **Thread-safe**: Recommended: one `viiper_client_t` per thread

!!! note "License"
    The C SDK is licensed under the **MIT License**, providing maximum flexibility for integration into your projects.  
    The core VIIPER server remains under its original license.

## Installation

### Building from Source

The C SDK is generated from the VIIPER server codebase:

```bash
go run ./cmd/viiper codegen --lang=c
```

Build the SDK:

```bash
cd ../clients/c
cmake -B build -G "Visual Studio 17 2022"  # Windows
cmake -B build                              # POSIX
cmake --build build --config Release
```

### Linking to Your Project

**CMake:**

```cmake
# Add viiper SDK
add_subdirectory(path/to/clients/c)
target_link_libraries(your_target PRIVATE viiper)

# Copy DLL on Windows (post-build)
if(WIN32)
    add_custom_command(TARGET your_target POST_BUILD
        COMMAND ${CMAKE_COMMAND} -E copy_if_different
        $<TARGET_FILE:viiper>
        $<TARGET_FILE_DIR:your_target>
    )
endif()
```

**Manual:**

- Include: `clients/c/include/viiper/viiper.h`
- Link: `clients/c/build/Release/viiper.lib` (Windows) or `libviiper.a` (POSIX)
- Runtime: Copy `viiper.dll` next to your executable (Windows)

## Quick Start

```c
#include <viiper/viiper.h>
#include <viiper/viiper_keyboard.h>
#include <stdio.h>

int main(void) {
    // Create new Viiper client
    viiper_client_t* client = NULL;
    int err = viiper_client_create("127.0.0.1", 3242, &client);
    if (err != 0) {
        fprintf(stderr, "Failed to connect: %s\n", viiper_strerror(err));
        return 1;
    }

    // Create or find a bus
    viiper_bus_list_response_t buses = {0};
    err = viiper_bus_list(client, &buses);
    uint32_t bus_id = (buses.BusesCount > 0) ? buses.Buses[0] : 0;
    
    if (bus_id == 0) {
        viiper_bus_create_response_t resp = {0};
        uint32_t desired_id = 1;
        err = viiper_bus_create(client, &desired_id, &resp); // NULL for auto-assign
        bus_id = resp.BusID;
    }

    // Add device and connect (convenience function)
    const char* device_type = "keyboard";
    viiper_device_create_request_t req = {
        .Type = &device_type,
        .IdVendor = NULL,
        .IdProduct = NULL
    };
    viiper_device_info_t dev_info = {0};
    viiper_device_t* device = NULL;
    err = viiper_add_device_and_connect(client, bus_id, &req, &dev_info, &device);

    // Send keyboard input
    viiper_keyboard_input_t input = {
        .modifiers = 0,
        .count = 1
    };
    uint8_t keys[] = {VIIPER_KEYBOARD_KEY_A};
    input.keys = keys;
    input.keys_count = 1;

    err = viiper_device_send(device, &input, sizeof(input.modifiers) + sizeof(input.count) + input.keys_count);

    // Cleanup
    viiper_device_close(device);
    
    char bus_id_str[32];
    snprintf(bus_id_str, sizeof(bus_id_str), "%u", bus_id);
    viiper_device_remove_response_t remove_resp = {0};
    viiper_bus_device_remove(client, bus_id_str, dev_info.DevId, &remove_resp);
    
    viiper_client_free(client);
    return 0;
}
```

## Device Stream API

### Creating a Device Stream

Manual approach (add device, then connect):

```c
// Add device first
const char* device_type = "keyboard";
viiper_device_create_request_t req = {
    .Type = &device_type,
    .IdVendor = NULL,  // Optional: set to specify custom VID
    .IdProduct = NULL  // Optional: set to specify custom PID
};

char bus_id_str[32];
snprintf(bus_id_str, sizeof(bus_id_str), "%u", bus_id);

viiper_device_info_t dev_info = {0};
viiper_error_t err = viiper_bus_device_add(client, bus_id_str, &req, &dev_info);
if (err != VIIPER_OK) {
    fprintf(stderr, "Failed to add device: %s\n", viiper_get_error(client));
}

// Then connect to its stream
viiper_device_t* device = NULL;
err = viiper_open_stream(client, bus_id, dev_info.DevId, &device);
if (err != VIIPER_OK) {
    fprintf(stderr, "Failed to open device stream: %s\n", viiper_get_error(client));
}
```

Convenience approach (add and connect in one call):

```c
const char* device_type = "xbox360";
viiper_device_create_request_t req = {
    .Type = &device_type,
    .IdVendor = NULL,
    .IdProduct = NULL
};

viiper_device_info_t dev_info = {0};
viiper_device_t* device = NULL;
viiper_error_t err = viiper_add_device_and_connect(client, bus_id, &req, &dev_info, &device);
if (err != VIIPER_OK) {
    fprintf(stderr, "Failed to add and connect device: %s\n", viiper_get_error(client));
}
```

### Sending Input

```c
viiper_mouse_input_t input = {
    .buttons = VIIPER_MOUSE_BTN_LEFT,
    .dx = 10,
    .dy = -5,
    .wheel = 0,
    .pan = 0
};

int err = viiper_device_send(device, &input, sizeof(input));
```

### Receiving Output (Callbacks)

```c
void on_led_update(void* user_data, const void* data, size_t len) {
    if (len < 1) return;
    uint8_t leds = ((uint8_t*)data)[0];
    printf("LEDs: NumLock=%d CapsLock=%d ScrollLock=%d\n",
           !!(leds & VIIPER_KEYBOARD_LED_NUM_LOCK),
           !!(leds & VIIPER_KEYBOARD_LED_CAPS_LOCK),
           !!(leds & VIIPER_KEYBOARD_LED_SCROLL_LOCK));
}

viiper_device_on_output(device, on_led_update, NULL);
```

### Closing a Stream

```c
viiper_device_close(device);
```

## Device-Specific Notes

Each device type has specific packet formats, constants, and wire protocols. For wire format details and usage patterns, see the [Devices](../devices/) section of the documentation.

The C SDK provides generated structs and constants in device-specific headers (e.g., `viiper_keyboard.h`, `viiper_mouse.h`, `viiper_xbox360.h`).

## Struct Packing

All device I/O structs use `#pragma pack(1)` to ensure wire compatibility (no padding).

```c
#pragma pack(push, 1)
typedef struct {
    uint8_t buttons;
    int8_t dx;
    // ...
} viiper_mouse_input_t;
#pragma pack(pop)
```

**Important:** Always ensure your compiler respects packing directives. MSVC and GCC/Clang handle this correctly by default.

## Troubleshooting

### Missing DLL on Windows

**Symptom:** Application crashes immediately with "viiper.dll not found"

**Solution:** Copy `viiper.dll` to the same directory as your executable:

```cmake
add_custom_command(TARGET your_target POST_BUILD
    COMMAND ${CMAKE_COMMAND} -E copy_if_different
    $<TARGET_FILE:viiper>
    $<TARGET_FILE_DIR:your_target>
)
```

### Repeated Keys Not Working

**Symptom:** Typing "Hello" outputs "Helo" (missing duplicate letter)

**Solution:** Add sufficient delays between key press, release, and next action:

```c
press_and_release(dev, VIIPER_KEYBOARD_KEY_L, 0);
Sleep(100);
press_and_release(dev, VIIPER_KEYBOARD_KEY_L, 0);
```

### Struct Padding Issues

**Symptom:** Device input is corrupted or "spazzing"

**Solution:** Verify `#pragma pack(1)` is applied to device structs. All generated headers include this by default.

## Examples

Full working examples are available in the repository:

- **Virtual Keyboard**: `examples/c/virtual_keyboard/main.c`
  - Types "Hello!" every 5 seconds
  - Reads LED state (NumLock, CapsLock, ScrollLock)

- **Virtual Xbox360 Controller**: `examples/c/virtual_x360_pad/main.c`
  - Simulates button presses and stick movements
  - Receives rumble feedback

Build and run:

```bash
cd examples/c
cmake -B build -G "Visual Studio 17 2022"
cmake --build build --config Release
./build/Release/virtual_keyboard.exe 127.0.0.1:3242
```

## See Also

- [Generator Documentation](generator.md): How the C SDK is generated
- [Go Client Documentation](go.md): Reference implementation
- [C++ SDK Documentation](cpp.md): Header-only C++ SDK
- [C# SDK Documentation](csharp.md): .NET SDK
- [Rust SDK Documentation](rust.md): Rust SDK
- [TypeScript SDK Documentation](typescript.md): Node.js SDK
- [API Overview](../api/overview.md): Management API reference
