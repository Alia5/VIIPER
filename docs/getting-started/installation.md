# Installation

VIIPER comes in two distinct flavors:

- **VIIPER Server** 
A self-contained, portable, standalone executable providing a lightweight TCP-based API for feeder application development. All client libraries are MIT licensed.
- **libVIIPER**  
 a single shared library that lets you emulate devices using USBIP directly from within your application. Requires your application to be GPL-3.0 licensed.

Regardless of the flavour you choose, VIIPER requires USBIP.

## Requirements

### USBIP

VIIPER relies on USBIP.
You must have a USBIP-Client implementation available on your system to use VIIPER's virtual devices.

=== "Windows"

    [usbip-win2](https://github.com/vadimgrn/usbip-win2) is by far the most complete implementation of USBIP for Windows (comes with a **SIGNED** kernel mode driver).

    **Install and done 😉**

    !!! warning "USBIP-Win2 security issue"
        The releases of usbip-win2 **currently** (at the time of writing) install the publicly available test signing CA as a _trusted root CA_ on your system.
        You can safely remove this CA after installation using `certmgr.msc` (run as admin) and removing the "USBIP" from the "Trusted Root Certification Authorities" -> "Certificates" list.

        **Alternatively**, you can download and install the **latest pre-release** driver manually from the
        [OSSign repository](https://github.com/OSSign/vadimgrn--usbip-win2/releases), which has this issue fixed already.
        _Note_ that the installer does not work, only the driver `.cat,.inf,.sys` files.

=== "Linux"

    #### Ubuntu/Debian

    ```bash
    sudo apt install linux-tools-generic
    ```

    [Ubuntu USBIP Manual](https://manpages.ubuntu.com/manpages/noble/man8/usbip.8.html)

    #### Arch Linux

    ```bash
    sudo pacman -S usbip
    ```

    [Arch Wiki: USBIP](https://wiki.archlinux.org/title/USB/IP)

    ### Linux Kernel Module Setup

    !!! info "USBIP Client Requirement"
        USBIP requires the `vhci-hcd` (Virtual Host Controller Interface) kernel module on Linux for client operations. This includes VIIPER's auto-attach feature and manual device attachment.

    Most Linux distributions include this module but do not load it automatically.

    #### One-Time Setup

    To load the module automatically on boot:

    ```bash
    echo "vhci-hcd" | sudo tee /etc/modules-load.d/vhci-hcd.conf
    sudo modprobe vhci-hcd
    ```

    #### Manual Loading

    To load the module for the current session only:

    ```bash
    sudo modprobe vhci-hcd
    ```

    #### Verification

    ```bash
    lsmod | grep vhci_hcd
    ```

---

=== "VIIPER Server"

    A standalone executable that exposes an API over TCP.

    ## Installing VIIPER

    VIIPER does not require system-wide installation.
    The `viiper` executable is completely self-contained (fully portable, no dependencies except USBIP) and can be:

    - Placed in any directory
    - Shipped alongside your application
    - Run directly without installation
    - Bundled with your application's distribution

    !!! warning "Daemon/Service Conflicts"
        If VIIPER is already running as a system service or daemon on the target machine, be aware of potential port conflicts. Applications should:
        - Check if VIIPER is already running before starting their own instance
          - Use the `ping` API endpoint to check for VIIPER presence and version
        - Connect to the existing VIIPER instance (if accessible)
        - Use a custom port via `--api.addr` flag to run a separate instance

    !!! info "Linux Permissions"
        On Linux, attaching devices via USBIP requires root permissions.
        You can run VIIPER with `sudo`, or configure appropriate udev rules to allow non-root users to attach devices.

    ### Pre-built Binaries

    Download the latest release from the [GitHub Releases](https://github.com/Alia5/VIIPER/releases) page. Pre-built binaries are available for:

    - Windows (x64, ARM64)
    - Linux (x64, ARM64)

    ### Automated Install Script

    The following scripts will download a VIIPER release, install it to a system location and configure it to start automatically on boot.

    !!! info "For Application Developers"
        The installation scripts are intended for **end-users** setting up a permanent VIIPER service on their system.

        If you are developing an application that uses VIIPER, I **strongly** encourage you to **not** install a permanent VIIPER service on your users' machines.

        Instead, bundle the (no dependencies, portable) VIIPER binary with your application and start/stop the server directly from your application as needed.
        You may need to check for existing VIIPER instances or use a custom port via `--api.addr` to avoid conflicts.

    !!! info "USBIP installed by scripts"
        The install scripts install and configure USBIP for you:

        - **Windows:** installs the usbip-win2 driver (admin prompt) and prompts for a reboot when drivers were added.
        - **Linux:** installs USBIP via the distro package manager (when available), loads `vhci_hcd` and configures it to autoload.

        If the automated USBIP setup fails, follow the [USBIP guide](usbip.md) to finish manually.

    === "Windows"

        ```powershell
        irm https://alia5.github.io/VIIPER/stable/install.ps1 | iex
        ```

        Installs to: `%LOCALAPPDATA%\VIIPER\viiper.exe`

        The scripts will:

        1. Download the specified VIIPER binary version
        2. Install it to the system location
        3. Install and configure USBIP (driver on Windows; packages/modules on Linux)
        4. Configure automatic startup (Registry RunKey on Windows, systemd service on Linux)
        5. Start/restart the VIIPER service

    === "Linux"

        ```bash
        curl -fsSL https://alia5.github.io/VIIPER/stable/install.sh | sh
        ```

        Installs to: `/usr/local/bin/viiper`

        The scripts will:

        1. Download the specified VIIPER binary version
        2. Install it to the system location
        3. Attempt to install and configure USBIP
        4. Load the `vhci_hcd` kernel module and configure it to autoload on boot
        5. Configure and run a systemd service

=== "libVIIPER"

    libVIIPER is a shared library (`libVIIPER.dll` on Windows, `libVIIPER.so` on Linux) that you link against directly from your application, eliminating the need for a separate VIIPER server process.

    !!! warning "License"
        Linking against libVIIPER requires your application to be licensed under the **GPL-3.0** (or a compatible license).
        If you cannot comply with the GPL-3.0, use the standalone executable and the [TCP API](../api/overview.md) instead. All client libraries are **MIT licensed**.

    ## Pre-built Binaries

    Download the latest `libVIIPER` release artifact from the [GitHub Releases](https://github.com/Alia5/VIIPER/releases) page.
    The archive contains:

    - `libVIIPER.dll` / `libVIIPER.so`: the shared library
    - `libVIIPER.h`: the C header
    - `libVIIPER.def`: Windows import definition (for generating `.lib`/`.dll.a` import libraries)

    ## Building from Source

    ```bash
    git clone https://github.com/Alia5/VIIPER.git
    cd VIIPER
    make libVIIPER
    ```

    The output will be in `dist/libVIIPER/`.

    !!! info "CGO Required"
        Building libVIIPER requires CGO (`CGO_ENABLED=1`) and a C compiler (GCC / MSVC / Clang) in `PATH`.
        On Windows, [mingw-w64](https://www.mingw-w64.org/) or MSVC is required.

    See [libVIIPER documentation](../libviiper/overview.md) for integration guides and examples.
