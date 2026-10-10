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

VIIPERs installation will install the necessary USBIP components/driver for you.  
It is only listed here for informational purposes.

=== "Windows"

    [usbip-win2](https://github.com/vadimgrn/usbip-win2) is by far the most complete implementation of USBIP for Windows (comes with a **SIGNED** kernel mode driver).

    **Install and done 😉**

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

    ### Installer

    The VIIPER executable is its own installer.  
    Run the release binary and you will be prompted to install VIIPER or run the sever portably. 

    === "Windows"

        Double-click `viiper-windows-x64.exe` (or run it without arguments from a terminal) and choose **Install/Update VIIPER**, or run:

        ```powershell
        .\viiper-windows-x64.exe install
        ```

        - **Update to the latest release**: checks GitHub and installs the newest release instead of the binary you ran
        - **Autostart**: starts the server when you log in
        - **Start menu shortcut** / **Desktop shortcut**
        - **Install USBIP**: downloads and runs the [usbip-win2](https://github.com/vadimgrn/usbip-win2) installer when it is missing or outdated (needs an elevated terminal; usbip-win2 asks for a reboot)
        - **Start the server now**

        Installs to `%LOCALAPPDATA%\VIIPER\viiper.exe` by default and registers VIIPER in **Apps & Features**, where it can be modified or uninstalled.

        Unattended install, for scripts and package managers (defaults for every choice, no output):

        ```powershell
        .\viiper-windows-x64.exe install --quiet
        ```

        Installs to: `%LOCALAPPDATA%\VIIPER\viiper.exe`

        Uninstall from **Apps & Features** or with:

        ```powershell
        viiper.exe uninstall
        ```

    === "Linux"

        ```bash
        chmod +x viiper-linux-amd64
        sudo ./viiper-linux-amd64 install
        ```

        - **Update to the latest release**: checks GitHub and installs the newest release instead of the binary you ran
        - **Add systemd service** / **Enable service**
        - **Install USBIP**: installs the distribution package (pacman, apt or dnf), loads `vhci_hcd` and configures it to autoload
        - **Start the server now**

        Installs to `/usr/local/bin/viiper` by default.

        Unattended install (defaults for every choice, no output):

        ```bash
        sudo ./viiper-linux-amd64 install --quiet
        ```

        Uninstall with:

        ```bash
        sudo viiper uninstall
        ```

        #### Manual installation

        For package maintainers, or if you prefer not to use the built-in installer.  
        This is what the installer does:

        1. Copy the binary and the bundled licenses:

            ```bash
            sudo install -m 755 viiper-linux-amd64 /usr/local/bin/viiper
            sudo install -m 644 licenses.txt /usr/local/bin/licenses.txt
            ```

        2. Install USBIP and load `vhci_hcd` on boot (see [USBIP](#usbip) above):

            ```bash
            echo "vhci_hcd" | sudo tee /etc/modules-load.d/viiper.conf
            sudo modprobe vhci_hcd
            ```

        3. Create `/etc/systemd/system/viiper.service`:

            ```ini
            [Unit]
            Description=VIIPER server
            After=network-online.target
            Wants=network-online.target

            [Service]
            Type=simple
            ExecStart="/usr/local/bin/viiper" server
            WorkingDirectory=/usr/local/bin
            Restart=on-failure

            [Install]
            WantedBy=multi-user.target
            ```

        4. Enable and start it:

            ```bash
            sudo systemctl daemon-reload
            sudo systemctl enable --now viiper.service
            ```

        To remove: stop and disable the service, delete the unit file, `/etc/modules-load.d/viiper.conf`, the binary and `licenses.txt`.

    ### One-line install

    Downloads the release binary and runs `install --quiet`.

    === "Windows"

        ```powershell
        irm https://alia5.github.io/VIIPER/stable/install.ps1 | iex
        ```

    === "Linux"

        ```bash
        curl -fsSL https://alia5.github.io/VIIPER/stable/install.sh | sh
        ```

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
    just build-libVIIPER
    ```

    The output will be in `dist/libVIIPER/`.

    !!! info "CGO Required"
        Building libVIIPER requires CGO (`CGO_ENABLED=1`) and a C compiler (GCC / MSVC / Clang) in `PATH`.
        On Windows, [llvm-mingw](https://github.com/mstorsjo/llvm-mingw) is recommended.

    See [libVIIPER documentation](../libviiper/overview.md) for integration guides and examples.
