package install

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	usbipWin2Version  = "0.9.8.1"
	usbipUninstallKey = `Software\Microsoft\Windows\CurrentVersion\Uninstall\{199505b0-b93d-4521-a8c7-897818e0205a}_is1`
	usbipInstallerURL = "https://github.com/vadimgrn/usbip-win2/releases/download/v." + usbipWin2Version + "/USBip-" + usbipWin2Version + "-%s.exe"
)

var ErrNeedsAdmin = errors.New("installing the USBIP driver needs elevated (administrator) privileges")

func (rebootRequired) ExitCode() int {
	return int(windows.ERROR_SUCCESS_REBOOT_REQUIRED)
}

func installUsbip(t *Target) error {
	installed := ""
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, usbipUninstallKey, registry.QUERY_VALUE|registry.WOW64_64KEY)
	if err == nil {
		installed, _, err = key.GetStringValue("DisplayVersion")
		key.Close() //nolint:errcheck
	}
	if err != nil && !errors.Is(err, registry.ErrNotExist) {
		return err
	}
	if !ShouldUpdate(installed, usbipWin2Version) {
		return nil
	}
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x64"
	}
	installer := filepath.Join(os.TempDir(), "USBip-setup.exe")
	err = downloadFile(fmt.Sprintf(usbipInstallerURL, arch), installer)
	if err != nil {
		return err
	}
	defer os.Remove(installer) //nolint:errcheck
	var switches []string
	if t.Quiet {
		switches = []string{
			"/VERYSILENT",
			"/SUPPRESSMSGBOXES",
			"/NORESTART",
			fmt.Sprintf("/RESTARTEXITCODE=%d", windows.ERROR_SUCCESS_REBOOT_REQUIRED),
		}
	}
	err = exec.Command(installer, switches...).Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == int(windows.ERROR_SUCCESS_REBOOT_REQUIRED) {
		return ErrRebootRequired
	}
	if err != nil && !windows.GetCurrentProcessToken().IsElevated() {
		return fmt.Errorf("%w: %w", ErrNeedsAdmin, err)
	}
	if err != nil {
		return fmt.Errorf("usbip-win2 installer: %w", err)
	}

	return nil
}
