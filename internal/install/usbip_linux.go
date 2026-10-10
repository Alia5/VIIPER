package install

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	vhciModule  = "vhci_hcd"
	procModules = "/proc/modules"
)

var (
	ErrNoPackageManager = errors.New("no supported package manager (pacman, apt, dnf)")

	usbipInstallCommands = map[string]string{
		"pacman": "pacman -S --noconfirm usbip",
		"apt":    "apt update && apt install -y linux-tools-generic",
		"dnf":    "dnf install -y usbip",
	}
)

func installUsbip(*Target) error {
	_, err := exec.LookPath("usbip")
	if err != nil {
		err = installUsbipPackage()
		if err != nil {
			return err
		}
	}
	content, err := os.ReadFile(procModules)
	if err != nil {
		return err
	}
	if strings.Contains(string(content), vhciModule+" ") {
		return nil
	}
	out, err := exec.Command("modprobe", vhciModule).CombinedOutput()
	if err != nil {
		return fmt.Errorf("modprobe %s failed: %w: %s", vhciModule, err, strings.TrimSpace(string(out)))
	}

	return os.WriteFile(modulesConfPath, []byte(vhciModule+"\n"), 0o644)
}

func installUsbipPackage() error {
	for pm, command := range usbipInstallCommands {
		_, err := exec.LookPath(pm)
		if err != nil {
			continue
		}
		out, err := exec.Command("sh", "-c", command).CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s failed: %w: %s", command, err, strings.TrimSpace(string(out)))
		}

		return nil
	}

	return ErrNoPackageManager
}
