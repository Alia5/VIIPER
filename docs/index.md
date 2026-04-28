<img src="viiper.svg" align="right" width="200"/>
<br />

# VIIPER 🐍

**Virtual** **I**nput over **IP** **E**mulato**R**

A **cross-platform virtual USB input framework** for creating virtual USB input devices (game controllers, keyboards, mice and more)
that are indistinguishable from real hardware to the operating system and applications.

## Quick Links

- [Installation (VIIPER Server)](getting-started/installation.md)  
    - [CLI Reference](cli/overview.md)
    - [API Reference](api/overview.md)
- [libVIIPER](libviiper/overview.md)
- [GitHub Repository](https://github.com/Alia5/VIIPER)

## What is VIIPER?

VIIPER lets developers create and programmatically control virtual USB input devices (using USBIP under the hood),
enabling seamless integration for gaming, automation, testing and remote control scenarios.

These virtual devices are indistinguishable from real hardware to the operating system and applications.

- Runs on Linux and Windows.  
- _(Optional)_ network support built in: control devices over a network with lower overhead than raw USBIP alone.  
- VIIPER abstracts away all USB / USBIP details.  
- VIIPER is portable and runs entirely in userspace.  
    - Utilizes a generic USBIP kernel mode driver  
      (built into Linux; on Windows [usbip-win2](https://github.com/vadimgrn/usbip-win2) provides a signed kernel mode driver)  
      New device types never require touching kernel code.  
- After installing USBIP once, VIIPER can run without additional dependencies or system-wide installation.  

VIIPER comes in two distinct flavors:

- **VIIPER server**  
  a self-contained, (no dependencies, statically linked) portable, standalone executable  
    - exposing a lightweight TCP-API
    - control devices from any language or machine on the network  
- **libVIIPER**  
  a single shared library to embed device emulation directly into your application  
  See Examples for C and C# [here](./examples/libVIIPER)  
  or the [libVIIPER documentation](libviiper/overview.md) for details and examples.  

For why you should pick one over the other see the [FAQ](#why-choose-the-the-standalone-executable-and-interfacing-via-tcp-over-and-the-shared-object-libviiper-library)

Beyond device emulation, VIIPER can proxy real USB devices for traffic inspection and reverse engineering.

---

## 🥫 Feeder application development

You have two options for developing feeder applications that control the virtual devices created by VIIPER:

- Use the standalone VIIPER server and interface via the exposed TCP-API (preferably using one of the available client libraries)
- Integrate libVIIPER directly into your application.  
  See [libVIIPER documentation](libviiper/overview.md) for details and examples.

### 🔌 API

VIIPER includes a lightweight TCP based API for device and bus management, as well as streaming device control.  
It's designed to be trivial to drive from any language that can open a TCP socket and send null-byte-terminated commands.  

!!! tip "Client Libraries Available"
    Most of the time, you don't need to implement that raw protocol yourself, as client libraries are available.  
    See [Client Libraries Available](api/overview.md).

- The TCP API uses a string-based request/response protocol
  terminated by null bytes (`\0`) for device and bus management.  
    - Requests have a "_path_" and optional payload (sometimes  JSON).  
    eg. `bus/{id}/add {"type": "keyboard", "idVendor": "0x6969"}\0`  
    - Responses are often JSON as well!
    - Errors are reported using JSON objectes similar to
    - [RFC 7807 Problem Details](https://datatracker.ietf.org/doc/html/rfc7807)  
 <sup>The use of JSON allows for future extenability without breaking compatibility ;)<sup>
- For controlling, or feeding, a device a long lived TCP stream is used, with a wire-protocol specific to each device type.  
  After an initial "_handshake_" (`bus/{busId}/{deviceId}\0`) a _device-specific **binary protocol**_ is used to send input reports and receive output reports (e.g., rumble commands).

VIIPER takes care of all USBIP protocol details, so you can focus on implementing the device logic only.  
On `localhost` VIIPER also automatically attached the USBIP client, so you don't have to worry about USBIP details at all.

!!! info "Security: Authentication & Encryption"
    VIIPER **requires authentication for remote connections**
    to prevent unauthorized device creation.  
    All authenticated connections use fast **ChaCha20-Poly1305 encryption**
    to protect against man-in-the-middle attacks.  
    Localhost connections are exempt from authentication by default for convenience.

See the [API documentation](api/overview) for details

---

## ❓ FAQ

### What is USBIP and why does VIIPER use it?

USBIP is a protocol that allows USB devices to be shared over a network.  
VIIPER uses it because it's already built into Linux and available for Windows, making virtual device emulation possible without writing custom kernel drivers yourself.

### Why choose the standalone executable and interfacing via TCP over, and the (shared-object) libVIIPER library

- Flexibility
    - allows one to use VIIPER as a service on the same host as the USBIP-Client and use the feeder on a different, remote machine.
    - allows for software written utilizing VIIPER to **not be** licensed under the terms of the GPLv3
    - **_future versions_**: Users can enhance VIIPER with device plugins, sharing a common wire-protocol, which can be dynamically incorporated.

### Can I use VIIPER for gaming?

Yes! VIIPER can create virtual input devices that appear as real hardware to games and applications.

This works with Steam, native Windows games and any other application that supports the emulated device types.

### How is VIIPER different from other controller emulators?

Many controller emulation approaches require writing a custom kernel driver for every device type you want to support.  
VIIPER uses USBIP to handle the USB protocol layer, so device emulation code lives entirely in userspace.  

USBIP itself does require a kernel driver.  
On Linux, the USBIP driver is built into the kernel.  
On Windows, [usbip-win2](https://github.com/vadimgrn/usbip-win2) provides a signed kernel mode driver.  
That driver is generic and does not need to know anything about specific device types.  
All device-type logic stays in userspace.  

This makes VIIPER portable, easier to extend and simpler to bundle with applications.  
Adding a new device type never requires touching kernel code.

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
