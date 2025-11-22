<img src="viiper.svg" align="right" width="200"/>
<br />

# VIIPER Documentation

Welcome to the VIIPER documentation!

VIIPER is a tool to create virtual input devices using USBIP.

## Quick Links

- [Installation](getting-started/installation.md)
- [CLI Reference](cli/overview.md)
- [API Reference](api/overview.md)
- [GitHub Repository](https://github.com/Alia5/VIIPER)

## What is VIIPER?

VIIPER creates virtual USB input devices using the USBIP protocol.  
These virtual devices appear as real hardware to the operating system and applications, allowing you to emulate controllers, keyboards, and other input devices without physical hardware.

VIIPER is a self-contained, standalone binary that uses USBIP to handle the USB protocol layer.  
Device emulation happens in userspace code instead of kernel drivers, so no kernel programming is required to add new device types.  
Users need USBIP installed once (built into Linux, usbip-win2 for Windows), after that VIIPER runs without additional dependencies or system-wide installation.  
The binary is portable and can be bundled with your application.

Beyond device emulation, VIIPER can proxy real USB devices for traffic inspection and reverse engineering.

All devices _**can and must be**_ controlled programmatically via an API.

## Key Features

- ✅ Virtual input device emulation over IP using USBIP
    - ✅ Xbox 360 controller emulation (virtual device);  see [Devices › Xbox 360 Controller](devices/xbox360.md)
    - ✅ HID Keyboard with N-key rollover and LED feedback; see [Devices › Keyboard](devices/keyboard.md)
    - ✅ HID Mouse with 5 buttons and horizontal/vertical wheel; see [Devices › Mouse](devices/mouse.md)
    - 🚧 Extensible architecture allows for more device types (other gamepads, specialized HID)
- ✅ USBIP server mode: expose virtual devices to remote clients
- ✅ **Automatic local attachment**: automatically controls usbip client on localhost to attach devices (enabled by default)
- ✅ Proxy mode: forward real USB devices and inspect/record traffic
- ✅ Cross-platform: works on Linux and Windows
- ✅ Flexible logging (including raw USB packet logs)
- ✅ API server for device/bus management and controlling virtual devices programmatically
- ✅ Multiple client SDKs for easy integration; see [Client SDKs](api/overview.md)  
  MIT Licensed

---

### Additional Resources / Misc

- [E2E Latency Benchmarks](testing/e2e_latency.md)
