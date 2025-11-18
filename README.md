<img src="docs/viiper.svg" align="right" width="128"/>
<br />


[![Build Status](https://github.com/alia5/VIIPER/actions/workflows/snapshots.yml/badge.svg)](https://github.com/alia5/VIIPER/actions/workflows/snapshots.yml)
[![License: GPL-3.0](https://img.shields.io/github/license/alia5/VIIPER)](https://github.com/alia5/VIIPER/blob/main/LICENSE.txt)
[![Client SDKs: MIT](https://img.shields.io/badge/Client_SDKs-MIT-green)](https://github.com/alia5/VIIPER/blob/main/viiper/internal/codegen/common/license.go)
[![Release](https://img.shields.io/github/v/release/alia5/VIIPER?include_prereleases&sort=semver)](https://github.com/alia5/VIIPER/releases)
[![Issues](https://img.shields.io/github/issues/alia5/VIIPER)](https://github.com/alia5/VIIPER/issues)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/alia5/VIIPER/pulls)
[![npm version](https://img.shields.io/npm/v/viiperclient?logo=npm)](https://www.npmjs.com/package/viiperclient)
[![npm downloads](https://img.shields.io/npm/dm/viiperclient?logo=npm&label=downloads)](https://www.npmjs.com/package/viiperclient)
[![NuGet version](https://img.shields.io/nuget/v/Viiper.Client?logo=nuget)](https://www.nuget.org/packages/Viiper.Client/)
[![NuGet downloads](https://img.shields.io/nuget/dt/Viiper.Client?logo=nuget&label=downloads)](https://www.nuget.org/packages/Viiper.Client/)
[![C SDK](https://img.shields.io/badge/C_SDK-artifact-blueviolet)](https://github.com/Alia5/VIIPER/releases)



# VIIPER 🐍

**Virtual** **I**nput over **IP** **E**mulato**R**

VIIPER is a tool to create virtual input devices using USBIP.

## ℹ️ About VIIPER

VIIPER creates virtual USB input devices using the USBIP protocol.  
These virtual devices appear as real hardware to the operating system and applications, allowing you to emulate controllers, keyboards, and other input devices without physical hardware.

VIIPER uses USBIP to handle the USB protocol layer, so device emulation happens in userspace code instead of kernel drivers.  
This means you install USBIP once (built into Linux, usbip-win2 for Windows), and VIIPER can emulate any device type without installing additional drivers.  
New device types can be added with pure Go code, no kernel programming required.

Beyond device emulation, VIIPER can proxy real USB devices for traffic inspection and reverse engineering. 

All devices _**can and must be**_ controlled programmatically via an API.

### ✨ Features

- ✅ Virtual input device emulation over IP using USBIP
  - ✅ Xbox 360 controller emulation (virtual device); see [Devices › Xbox 360 Controller](docs/devices/xbox360.md)
  - ✅ HID Keyboard with N-key rollover and LED feedback; see [Devices › Keyboard](docs/devices/keyboard.md)
  - ✅ HID Mouse with 5 buttons and horizontal/vertical wheel; see [Devices › Mouse](docs/devices/mouse.md)
  - 🔜 ???  
    🚧 Extensible architecture allows for more device types (other gamepads, specialized HID)
- ✅ USBIP server mode: expose virtual devices to remote clients
- ✅ Proxy mode: forward real USB devices and inspect/record traffic (for reversing)
- ✅ Cross-platform: works on Linux and Windows
- ✅ Flexible logging (including raw USB packet logs)
- ✅ API server for device/bus management and controlling virtual devices programmatically
- ✅ Multiple client SDKs (C, C#, Typescript/Javascript) for easy integration; see [Client SDKs](docs/api/overview.md)  
  MIT Licensed

## 🔌 Requirements

VIIPER relies on USBIP.  
You must have USBIP installed on your system.

**Linux:**

- **Arch Linux:**
  - Install: `sudo pacman -S usbip`
  - Docs: [Arch Wiki: USBIP](https://wiki.archlinux.org/title/USB/IP)

- **Ubuntu:**  
  - Install: `sudo apt install linux-tools-generic`
  - Docs: [Ubuntu USBIP Manual](https://manpages.ubuntu.com/manpages/noble/man8/usbip.8.html)

**Windows:**

- [usbip-win2](https://github.com/vadimgrn/usbip-win2) is by far the most complete implementation of USBIP for Windows (comes with a **SIGNED** kernel mode driver).

## 🔌 API

VIIPER includes an  API for device and bus management, as well as streaming device control.  
Each device type exposes its own control interface via the API.

See the [API documentation](./docs/api) for details (🚧 in progress 🚧).

## 🛠️ Development

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

## 🤝 Contributing

Contributions are welcome!  
Please open issues or pull requests on GitHub.  
See the [issues page](https://github.com/Alia5/VIIPER/issues) for bugs and feature requests.

## ❓ FAQ

### What is USBIP and why does VIIPER use it?

USBIP is a protocol that allows USB devices to be shared over a network.  
VIIPER uses it because it's already built into Linux and available for Windows, making virtual device emulation possible without writing custom kernel drivers yourself.

### Can I use VIIPER for gaming?

Yes! VIIPER can create virtual controllers (currently only Xbox360) that appear as real hardware to games and applications.
This works with Steam, native Windows games, and any other application supporting controllers.

### How is VIIPER different from other controller emulators?

VIIPER uses USBIP to handle the USB protocol layer, so device emulation happens in  userspace code instead of kernel drivers.  
This means you install USBIP once (built into Linux, usbip-win2 for Windows), and VIIPER can emulate any device type without installing additional drivers.  
New device types can be added with pure Go code, no kernel programming required.

### Can I add support for other device types?

Yes! VIIPER's architecture is designed to be extensible.  
Check the [xbox360 device implementation](./viiper/pkg/device/xbox360/) as a reference for creating new device types.

### What about the proxy mode?

Proxy mode sits between a USBIP client and a USBIP server (like a Linux machine sharing real USB devices).  
VIIPER intercepts and logs all USB traffic passing through, without handling the devices directly.  
Useful for reverse engineering USB protocols and understanding how devices communicate.

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
