# libVIIPER Documentation

libVIIPER is a shared library (`libVIIPER.dll` on Windows, `libVIIPER.so` on Linux) that embeds the full VIIPER USB/USBIP stack directly into your application.

- Single shared library (`libVIIPER.dll` / `libVIIPER.so`)
- Pure C API callable from any language with C FFI support
- In-process + threadsafe  
  the USBIP server runs in a background thread inside your application
- Optional auto-attach to the local USBIP client on the same machine

!!! warning "License"
    libVIIPER is licensed under **GPL-3.0**.  
    Linking against it **requires your application to be GPL-3.0 compatible**.  
    If your project cannot comply with the GPL-3.0, use the standalone VIIPER executable and the [TCP API](../api/overview.md) instead. All TCP client libraries are **MIT licensed**.

!!! info "USBIP Required"
    libVIIPER uses USBIP internally. A USBIP client must be installed on the target machine.  
    See [Installation › Requirements](../getting-started/installation.md#usbip) for setup instructions.

## API Overview

The libVIIPER C API is declared in `libVIIPER.h`.  
All functions return `bool` (`true` on success, `false` on failure).  
Handles (`USBServerHandle`, `Xbox360DeviceHandle`, …) are opaque `uintptr_t` values.

### Server lifecycle

| Function                               | Description                               |
| -------------------------------------- | ----------------------------------------- |
| `NewUSBServer(config, &handle, logCb)` | Start a USB server in a background thread |
| `CloseUSBServer(handle)`               | Stop the server and free all resources    |

### Bus management

| Function                             | Description                                       |
| ------------------------------------ | ------------------------------------------------- |
| `CreateUSBBus(serverHandle, &busID)` | Create a new USB bus (pass `0` to auto-assign ID) |
| `RemoveUSBBus(serverHandle, busID)`  | Remove a bus and all its devices                  |

## Examples

Full working examples are in [`examples/libVIIPER/`](https://github.com/Alia5/VIIPER/tree/main/examples/libVIIPER).

=== "C"

    ```c
    USBServerConfig conf = { .addr = "localhost:3245" };
    USBServerHandle serverHandle = 0;
    NewUSBServer(&conf, &serverHandle, logCallback);

    uint32_t busID = 0;
    CreateUSBBus(serverHandle, &busID);

    Xbox360DeviceHandle deviceHandle = 0;
    CreateXbox360Device(serverHandle, &deviceHandle, busID, /*autoAttach=*/true, 0, 0, 0);

    SetXbox360RumbleCallback(deviceHandle, rumbleCallback);

    Xbox360DeviceState state = {0};
    while (running) {
        // only required when an actual change occurs
        state.Buttons = XBOX360_BUTTON_A;
        state.LT      = 128;
        state.LX      = 20000;
        SetXbox360DeviceState(deviceHandle, state);
        _sleep(16);
    }

    CloseUSBServer(serverHandle);
    ```

=== "C#"

    ```csharp
    USBServerConfig conf = new() { addr = "localhost:3245" };
    LibVIIPER.NewUSBServer(ref conf, out nuint serverHandle, logCb);

    uint busID = 0;
    LibVIIPER.CreateUSBBus(serverHandle, ref busID);

    LibVIIPER.CreateXbox360Device(serverHandle, out nuint deviceHandle, busID, autoAttachLocalhost: true, 0, 0, 0);

    Xbox360RumbleCallbackDelegate rumbleCb = RumbleCallback;
    LibVIIPER.SetXbox360RumbleCallback(deviceHandle, rumbleCb);

    Xbox360DeviceState state = new();
    while (running) {
        // only required when an actual change occurs
        state.Buttons = Xbox360Buttons.A;
        state.LT      = 128;
        state.LX      = 20000;
        LibVIIPER.SetXbox360DeviceState(deviceHandle, state);
        Thread.Sleep(16);
    }

    LibVIIPER.CloseUSBServer(serverHandle);
    ```

    See [`examples/libVIIPER/C#/`](https://github.com/Alia5/VIIPER/tree/main/examples/libVIIPER/C%23) for the full project including P/Invoke declarations.

## Devices

- [Xbox 360 Controller](../devices/xbox360.md)
- [DualShock 4](../devices/dualshock4.md)
- [Keyboard](../devices/keyboard.md)
- [Mouse](../devices/mouse.md)

### Logging

Pass a `VIIPERLogCallback` to `NewUSBServer` to receive log messages from the library.  
Pass `NULL` to discard all log output.

```c
typedef enum {
    VIIPER_LOG_DEBUG = -4,
    VIIPER_LOG_INFO  = 0,
    VIIPER_LOG_WARN  = 4,
    VIIPER_LOG_ERROR = 8,
} VIIPERLogLevel;

typedef void (*VIIPERLogCallback)(VIIPERLogLevel level, const char* message);
```


