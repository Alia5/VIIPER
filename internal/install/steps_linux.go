package install

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	exeName         = "viiper"
	serviceName     = "viiper.service"
	servicePath     = "/etc/systemd/system/viiper.service"
	modulesConfPath = "/etc/modules-load.d/viiper.conf"
)

var (
	ErrNeedsRoot = errors.New("install needs root")

	installSteps = []step{
		{
			label: "Check for root",
			run:   requireRoot,
		},
		{
			label: "Check for the latest release",
			run:   checkLatest,
		},
		{
			label:    "Update to the latest release",
			optional: true,
			enabled:  true,
			run:      downloadLatest,
		},
		{
			label: "Remove old systemd service",
			run:   removeService,
		},
		{
			label: "Copy VIIPER",
			run: func(t *Target) error {
				return copyFile(t.Source, t.Exe)
			},
		},
		{
			label: "Copy licenses",
			run: func(t *Target) error {
				licenses, err := bundle.ReadFile("bundle/licenses.txt")
				if err != nil {
					return err
				}

				return writeFile(filepath.Join(filepath.Dir(t.Exe), "licenses.txt"), licenses)
			},
		},
		{
			label:    "Add systemd service",
			optional: true,
			enabled:  true,
			run:      addService,
		},
		{
			label:    "Enable service",
			optional: true,
			enabled:  true,
			run:      enableService,
		},
		{
			label:    "Install USBIP",
			optional: true,
			enabled:  true,
			run:      installUsbip,
		},
		{
			label:    "Start the server now",
			optional: true,
			enabled:  true,
			run:      startServer,
		},
	}
	uninstallSteps = []step{
		{
			label: "Check for root",
			run:   requireRoot,
		},
		{
			label: "Remove systemd service",
			run:   removeService,
		},
		{
			label: "Remove module autoload",
			run: func(*Target) error {
				return os.RemoveAll(modulesConfPath)
			},
		},
		{
			label: "Remove licenses",
			run: func(t *Target) error {
				return os.RemoveAll(filepath.Join(filepath.Dir(t.Exe), "licenses.txt"))
			},
		},
		{
			label: "Remove VIIPER",
			run: func(t *Target) error {
				return os.RemoveAll(t.Exe)
			},
		},
	}
)

func DefaultExe() (string, error) {
	return filepath.Join("/usr/local/bin", exeName), nil
}

func requireRoot(*Target) error {
	if os.Geteuid() != 0 {
		return ErrNeedsRoot
	}

	return nil
}

func addService(t *Target) error {
	err := os.WriteFile(servicePath, []byte(systemdUnitContent(t.Exe)), 0o644)
	if err != nil {
		return err
	}

	return runSystemctl("daemon-reload")
}

func enableService(*Target) error {
	return runSystemctl("enable", serviceName)
}

func startServer(t *Target) error {
	_, err := os.Stat(servicePath)
	if err == nil {
		return runSystemctl("restart", serviceName)
	}

	return exec.Command(t.Exe, "server").Start()
}

func removeService(*Target) error {
	_, err := os.Stat(servicePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return errors.Join(
		runSystemctl("stop", serviceName),
		runSystemctl("disable", serviceName),
		os.Remove(servicePath),
		runSystemctl("daemon-reload"),
	)
}

func systemdUnitContent(exePath string) string {
	return fmt.Sprintf(`[Unit]
Description=VIIPER server
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=%q server
WorkingDirectory=%s
Restart=on-failure

[Install]
WantedBy=multi-user.target
`, exePath, filepath.Dir(exePath))
}

func runSystemctl(args ...string) error {
	output, err := exec.Command("systemctl", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl %s failed: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}

	return nil
}
