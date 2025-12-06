<img src="docs/viiper.svg" align="right" width="128"/>
<br />

[![Build Status](https://github.com/alia5/VIIPER/actions/workflows/snapshots.yml/badge.svg)](https://github.com/alia5/VIIPER/actions/workflows/snapshots.yml)
[![License: GPL-3.0](https://img.shields.io/github/license/alia5/VIIPER)](https://github.com/alia5/VIIPER/blob/main/LICENSE.txt)
[![Client Libraries: MIT](https://img.shields.io/badge/Client_Libraries-MIT-green)](https://github.com/alia5/VIIPER/blob/main/internal/codegen/common/license.go)
[![Release](https://img.shields.io/github/v/release/alia5/VIIPER?include_prereleases&sort=semver)](https://github.com/alia5/VIIPER/releases)
[![Downloads](https://img.shields.io/github/downloads/alia5/VIIPER/total?logo=github)](https://github.com/alia5/VIIPER/releases)
[![Issues](https://img.shields.io/github/issues/alia5/VIIPER)](https://github.com/alia5/VIIPER/issues)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/alia5/VIIPER/pulls)
[![npm version](https://img.shields.io/npm/v/viiperclient?logo=npm)](https://www.npmjs.com/package/viiperclient)
[![npm downloads](https://img.shields.io/npm/dm/viiperclient?logo=npm&label=downloads)](https://www.npmjs.com/package/viiperclient)
[![NuGet version](https://img.shields.io/nuget/v/Viiper.Client?logo=nuget)](https://www.nuget.org/packages/Viiper.Client/)
[![NuGet downloads](https://img.shields.io/nuget/dt/Viiper.Client?logo=nuget&label=downloads)](https://www.nuget.org/packages/Viiper.Client/)
[![crates.io version](https://img.shields.io/crates/v/viiper-client?logo=rust)](https://crates.io/crates/viiper-client)
[![crates.io downloads](https://img.shields.io/crates/d/viiper-client?logo=rust&label=downloads)](https://crates.io/crates/viiper-client)
[![C Client Library](https://img.shields.io/badge/C_Client_Library-artifact-blueviolet)](https://github.com/Alia5/VIIPER/releases)
[![C++ Client Library](https://img.shields.io/badge/C++_Client_Library-Header_Only-blue)](https://github.com/Alia5/VIIPER/releases)

# VIIPER 🐍

**Virtual** **I**nput over **IP** **E**mulato**R**

VIIPER is a tool to create virtual input devices using USBIP.

## ℹ️ About VIIPER

VIIPER creates virtual USB input devices using the USBIP protocol.  
These virtual devices appear as real hardware to the operating system and applications, allowing you to emulate controllers, keyboards, and other input devices without physical hardware.

- VIIPER abstracts away all USB / USBIP details.  
- Device emulation happens in userspace code instead of kernel drivers, so no kernel programming is required to add new device types.  
- Users need USBIP installed once (built into Linux, usbip-win2 for Windows), after that VIIPER runs without additional dependencies or system-wide installation.  

VIIPER _currently_ comes in a single flavor:

- a self-contained, (no dependencies) portable, standalone executable.  
  providing a lightweight TCP based API for feeder application development.  
- There will eventually be a library version (libVIIPER) that you can link against directly from your application.  
For more information, see [FAQ](#why-is-this-a-standalone-executable-that-i-have-to-interface-via-tcp-and-not-a-shared-object-library-in-itself)  

Beyond device emulation, VIIPER can proxy real USB devices for traffic inspection and reverse engineering.

### ✨ Features

- ✅ Virtual input device emulation over IP using USBIP
  - ✅ Xbox 360 controller emulation (virtual device); see [Devices › Xbox 360 Controller](docs/devices/xbox360.md)
  - ✅ HID Keyboard with N-key rollover and LED feedback; see [Devices › Keyboard](docs/devices/keyboard.md)
  - ✅ HID Mouse with 5 buttons and horizontal/vertical wheel; see [Devices › Mouse](docs/devices/mouse.md)
  - 🔜 ???  
    🔜 Future plugin system allows for more device types (other gamepads, specialized HID)
- ✅ USBIP server mode: expose virtual devices to remote clients
- ✅ **Automatic local attachment**: automatically controls usbip client on localhost to attach devices (enabled by default)
- ✅ Proxy mode: forward real USB devices and inspect/record traffic (for reversing)
- ✅ Cross-platform: works on Linux and Windows
- ✅ Flexible logging (including raw USB packet logs)
- ✅ API server for device/bus management and controlling virtual devices programmatically
- ✅ Multiple client libraries for easy integration; see [Client Libraries](docs/api/overview.md)  
  MIT Licensed
- 🔜 _libVIIPER_ to link against, directly incoporating VIIPER into your feeder application.  

## 🔌 Requirements

**Linux:**

- **Arch Linux:**
  - Install: `sudo pacman -S usbip`
  - Docs: [Arch Wiki: USBIP](https://wiki.archlinux.org/title/USB/IP)

- **Ubuntu:**  
  - Install: `sudo apt install linux-tools-generic`
  - Docs: [Ubuntu USBIP Manual](https://manpages.ubuntu.com/manpages/noble/man8/usbip.8.html)

**Windows:**

- [usbip-win2](https://github.com/vadimgrn/usbip-win2) is by far the most complete implementation of USBIP for Windows (comes with a **SIGNED** kernel mode driver).

---

## 🥫 Feeder application development

VIIPER _currently_ comes in a single flavor:

- a standalone executable that exposes an API over TCP.
- There will eventually be a library version (libVIIPER) that you can link against directly from your application.  
For more information, see [FAQ](#why-is-this-a-standalone-executable-that-i-have-to-interface-via-tcp-and-not-a-shared-object-library-in-itself)  

### 🔌 API

VIIPER includes a lightweight TCP based API for device and bus management, as well as streaming device control.  
It's designed to be trivial to drive from any language that can open a TCP socket and send null-byte-terminated commands.  

> ⚠️ Most of the time, you don't need to implement that raw protocol yourself, as client libraries are available.  
See [Client Libraries Available](docs/api/overview.md).

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

See the [API documentation](./docs/api) for details

---

## 🛠️ VIIPER development

### 🧰 Prerequisites

- [Go](https://go.dev/) 1.25 or newer
- USBIP installed
- (Optional) [Make](https://www.gnu.org/software/make/)
  - Linux/macOS: Usually pre-installed
  - Windows: `winget install ezwinports.make`

### 🔄 Building from Source

```bash
git clone https://github.com/Alia5/VIIPER.git
cd VIIPER
make build
```

The binary will be in `dist/viiper` (or `dist/viiper.exe` on Windows).

For more build options:

```bash
make help              # Show all available targets
make test              # Run tests
```

---

## 🤝 Contributing

Contributions are welcome!  
Please open issues or pull requests on GitHub.  
See the [issues page](https://github.com/Alia5/VIIPER/issues) for bugs and feature requests.

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
Check the [xbox360 device implementation](./device/xbox360/) as a reference for creating new device types.  
In the future there will be a plugin system to load and expose device types dynamically.

### What about the proxy mode?

Proxy mode sits between a USBIP client and a USBIP server (like a Linux machine sharing real USB devices).  
VIIPER intercepts and logs all USB traffic passing through, without handling the devices directly.  
Useful for reverse engineering USB protocols and understanding how devices communicate.

### What about TCP overhead or input latency performance?

End-to-end input latency for virtual devices created with VIIPER is typically well below 1 millisecond on a modern desktop (e.g. Windows / Ryzen 3900X test machine).  
Detailed methodology and sample runs can be found in [E2E Latency Benchmarks](docs/testing/e2e_latency.md).

---

## 📄 License

```license
VIIPER - Virtual Input over IP EmulatoR

Copyright (C) 2025 Peter Repukat

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
```

## Credits / Inspiration

[REDACTED-Bus aka ViGEmBus](https://github.com/nefarius/ViGEmBus) - (Retired) Windows kernel-mode driver emulating well-known USB game controllers.
