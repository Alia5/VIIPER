<img src="viiper.svg" align="right" width="200"/>
<br />

# VIIPER 🐍

**Virtual** **I**nput over **IP** **E**mulato**R**

## Quick Links

- [Installation](getting-started/installation.md)
- [CLI Reference](cli/overview.md)
- [API Reference](api/overview.md)
- [GitHub Repository](https://github.com/Alia5/VIIPER)

## What is VIIPER?

VIIPER lets developers create virtual USB input devices (like game controllers, keyboards, and mice) that can be controlled programmatically (even over a network!) (using USBIP under the hood).  
These virtual devices are indistinguishable from real hardware to the operating system and applications, enabling seamless integration for testing, automation, and remote control scenarios.

- VIIPER abstracts away all USB / USBIP details.  
- Device emulation happens in userspace code instead of kernel drivers, so no kernel programming is required to add new device types.  
- Users need USBIP installed once (built into Linux, usbip-win2 for Windows), after that VIIPER can run without additional dependencies or system-wide installation.  

VIIPER _currently_ comes in a single flavor:

- a self-contained, (no dependencies) portable, standalone executable.  
  providing a lightweight TCP based API for feeder application development.  
- There will eventually be a library version (libVIIPER) that you can link against directly from your application.  
For more information, see [FAQ](#why-is-this-a-standalone-executable-that-i-have-to-interface-via-tcp-and-not-a-shared-object-library-in-itself)  

Beyond device emulation, VIIPER can proxy real USB devices for traffic inspection and reverse engineering.

---

## 🥫 Feeder application development

VIIPER _currently_ comes in a single flavor:

- a standalone executable that exposes an API over TCP.
- There will eventually be a shared-library version (libVIIPER) that you can link against directly from your application.  
For more information, see [FAQ](#why-is-this-a-standalone-executable-that-i-have-to-interface-via-tcp-and-not-a-shared-object-library-in-itself)  

### 🔌 API

VIIPER includes a lightweight TCP based API for device and bus management, as well as streaming device control.  
It's designed to be trivial to drive from any language that can open a TCP socket and send null-byte-terminated commands.  

!!! tip "Client Libraries Available"
    Most of the time, you don't need to implement that raw protocol yourself, as client libraries are available.  
    See [Client Libraries Available](api/overview.md).

- The TCP API uses a string-based request/response protocol terminated by null bytes (`\0`) for device and bus management.  
  - Requests have a "_path_" and optional payload (sometimes  JSON).  
    eg. `bus/{id}/add {"type": "keyboard", "idVendor": "0x6969"}\0`  
  - Responses are often JSON as well!
  - Errors are reported using JSON objectes similar to [RFC 7807 Problem Details](https://datatracker.ietf.org/doc/html/rfc7807)  
 <sup>The use of JSON allows for future extenability without breaking compatibility ;)<sup>
- For controlling, or feeding, a device a long lived TCP stream is used, with a wire-protocol specific to each device type.  
  After an initial "_handshake_" (`bus/{busId}/{deviceId}\0`) a _device-specific **binary protocol**_ is used to send input reports and receive output reports (e.g., rumble commands).

VIIPER takes care of all USBIP protocol details, so you can focus on implementing the device logic only.  
On `localhost` VIIPER also automatically attached the USBIP client, so you don't have to worry about USBIP details at all.

See the [API documentation](api/overview) for details

---

## ❓ FAQ

### What is USBIP and why does VIIPER use it?

USBIP is a protocol that allows USB devices to be shared over a network.  
VIIPER uses it because it's already built into Linux and available for Windows, making virtual device emulation possible without writing custom kernel drivers yourself.

### Why is this a standalone executable that I have to interface via TCP, and not a (shared-object) library in itself

- Flexibility
  - allows one to use VIIPER as a service on the same host as the USBIP-Client and use the feeder on a different, remote machine.
  - allows for software written utilizing VIIPER to **not be** licensed under the terms of the GPLv3
  - **_future versions_**: Users can enhance VIIPER with device plugins, sharing a common wire-protocol, which can be dynamically incorporated.
- **That said**, there **will be** a _libVIIPER_  that you can link against, eleminating multi-process and potential firewall issues.  
  Note that this **will require** your application to be licensed under the terms of the GPLv3 (or comptible license)

### Can I use VIIPER for gaming?

Yes! VIIPER can create virtual controllers (currently only Xbox360) that appear as real hardware to games and applications.
This works with Steam, native Windows games, and any other application supporting controllers.

### How is VIIPER different from other controller emulators?

Most controller emulators require custom kernel drivers for each device type.  
VIIPER uses USBIP to handle the USB protocol layer, allowing device emulation in userspace without kernel drivers.  
This makes VIIPER portable, easier to extend, and simpler to bundle with applications.

### Can I add support for other device types?

Yes! VIIPER's architecture is designed to be extensible.  
In the future there will be a plugin system to load and expose device types dynamically.

### What about the proxy mode?

Proxy mode sits between a USBIP client and a USBIP server (like a Linux machine sharing real USB devices).  
VIIPER intercepts and logs all USB traffic passing through, without handling the devices directly.  
Useful for reverse engineering USB protocols and understanding how devices communicate.

### What about TCP overhead or input latency performance?

End-to-end input latency for virtual devices created with VIIPER is typically well below 1 millisecond on a modern desktop (e.g. Windows / Ryzen 3900X test machine).  
Detailed methodology and sample runs can be found in [E2E Latency Benchmarks](testing/e2e_latency.md).  
However, to not stress the CPU excessively, reports get batched and sent every millisecond. So the best you will achive is a 1000Hz update rate, which is more than enough and more than what most real hardware devices provide.  
_Note_: Actual device polling rates may be lower depending on the device type and configuration.
