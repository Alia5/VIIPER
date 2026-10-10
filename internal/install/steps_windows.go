package install

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/Alia5/VIIPER/internal/configpaths"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	exeName            = "viiper.exe"
	runKeyPath         = `Software\Microsoft\Windows\CurrentVersion\Run`
	runValueKey        = "VIIPER"
	appsAndFeaturesKey = `Software\Microsoft\Windows\CurrentVersion\Uninstall\VIIPER`
	createShortcut     = `$s = (New-Object -ComObject WScript.Shell).CreateShortcut("%s"); $s.TargetPath = "%s"; $s.Save()`
	removeAfterExit    = `Wait-Process -Id %d -ErrorAction SilentlyContinue; Remove-Item -LiteralPath "%s"; Remove-Item -LiteralPath "%s" -ErrorAction SilentlyContinue`
)

var (
	installSteps = []step{
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
			label: "Stop running VIIPER",
			run:   stopServer,
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
			label: "Clear old autostart entry",
			run: func(*Target) error {
				return deleteRegistryValue(runKeyPath, runValueKey)
			},
		},
		{
			label: "Register in Apps & Features",
			run: func(t *Target) error {
				return setRegistryValues(
					appsAndFeaturesKey,
					map[string]string{
						"DisplayName":          "VIIPER",
						"DisplayVersion":       strings.TrimPrefix(t.Version, "v"),
						"Publisher":            "Peter Repukat",
						"InstallLocation":      filepath.Dir(t.Exe),
						"UninstallString":      fmt.Sprintf("\"%s\" uninstall", t.Exe),
						"QuietUninstallString": fmt.Sprintf("\"%s\" uninstall --quiet", t.Exe),
						"ModifyPath":           fmt.Sprintf("\"%s\" install", t.Exe),
					},
				)
			},
		},
		{
			label:    "Autostart",
			optional: true,
			enabled:  true,
			run:      enableAutostart,
		},
		{
			label:    "Start menu shortcut",
			optional: true,
			enabled:  true,
			run: func(t *Target) error {
				return writeShortcut(windows.FOLDERID_Programs, t.Exe)
			},
		},
		{
			label:    "Desktop shortcut",
			optional: true,
			run: func(t *Target) error {
				return writeShortcut(windows.FOLDERID_Desktop, t.Exe)
			},
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
			label: "Stop running VIIPER",
			run:   stopServer,
		},
		{
			label: "Remove autostart entry",
			run: func(*Target) error {
				return deleteRegistryValue(runKeyPath, runValueKey)
			},
		},
		{
			label: "Remove from Apps & Features",
			run: func(*Target) error {
				err := registry.DeleteKey(registry.CURRENT_USER, appsAndFeaturesKey)
				if errors.Is(err, registry.ErrNotExist) {
					return nil
				}

				return err
			},
		},
		{
			label: "Remove start menu shortcut",
			run: func(*Target) error {
				return removeShortcut(windows.FOLDERID_Programs)
			},
		},
		{
			label: "Remove desktop shortcut",
			run: func(*Target) error {
				return removeShortcut(windows.FOLDERID_Desktop)
			},
		},
		{
			label: "Remove licenses",
			run: func(t *Target) error {
				return os.RemoveAll(filepath.Join(filepath.Dir(t.Exe), "licenses.txt"))
			},
		},
		{
			label: "Remove VIIPER after exit",
			run: func(t *Target) error {
				cleanup := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", fmt.Sprintf(removeAfterExit, os.Getpid(), t.Exe, filepath.Dir(t.Exe)))
				cleanup.SysProcAttr = &syscall.SysProcAttr{
					HideWindow:    true,
					CreationFlags: windows.DETACHED_PROCESS,
				}

				return cleanup.Start()
			},
		},
	}
)

func DefaultExe() (string, error) {
	local, err := os.UserCacheDir()

	return filepath.Join(local, "VIIPER", exeName), err
}

func enableAutostart(t *Target) error {
	cfgDir, err := configpaths.DefaultConfigDir()
	if err != nil {
		return err
	}

	return setRegistryValues(
		runKeyPath,
		map[string]string{
			runValueKey: fmt.Sprintf("\"%s\" server --log.file \"%s\"", t.Exe, filepath.Join(cfgDir, "viiper.log")),
		},
	)
}

func startServer(t *Target) error {
	cfgDir, err := configpaths.DefaultConfigDir()
	if err != nil {
		return err
	}
	err = os.MkdirAll(cfgDir, 0o755)
	if err != nil {
		return err
	}

	return exec.Command(t.Exe, "server", "--log.file", filepath.Join(cfgDir, "viiper.log")).Start()
}

func stopServer(t *Target) error {
	previous, err := currentAutorunExe()
	if err != nil {
		return err
	}
	if previous != "" && previous != t.Exe {
		err = Kill(previous)
		if err != nil {
			return err
		}
	}

	return Kill(t.Exe)
}

func writeShortcut(folder *windows.KNOWNFOLDERID, target string) error {
	dir, err := windows.KnownFolderPath(folder, 0)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "VIIPER.lnk")
	output, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", fmt.Sprintf(createShortcut, path, target)).CombinedOutput()
	if err != nil {
		return fmt.Errorf("shortcut %s failed: %w: %s", path, err, strings.TrimSpace(string(output)))
	}

	return nil
}

func removeShortcut(folder *windows.KNOWNFOLDERID) error {
	dir, err := windows.KnownFolderPath(folder, 0)
	if err != nil {
		return err
	}

	return os.RemoveAll(filepath.Join(dir, "VIIPER.lnk"))
}

func currentAutorunExe() (string, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if errors.Is(err, registry.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	defer key.Close() //nolint:errcheck
	value, _, err := key.GetStringValue(runValueKey)
	if errors.Is(err, registry.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	value, quoted := strings.CutPrefix(strings.TrimSpace(value), "\"")
	if quoted {
		value, _, _ = strings.Cut(value, "\"")
	}
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return "", nil
	}

	return filepath.Clean(fields[0]), nil
}

func Running(exe string) ([]int, error) {
	script := fmt.Sprintf(
		"$ErrorActionPreference='SilentlyContinue';$t='%s';Get-CimInstance Win32_Process | Where-Object { $_.ExecutablePath -eq $t } | Select-Object -ExpandProperty ProcessId",
		strings.ReplaceAll(filepath.Clean(exe), "'", "''"),
	)
	output, err := exec.Command("powershell", "-NoProfile", "-Command", script).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("process query failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	var pids []int
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		pid, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil || pid == os.Getpid() {
			continue
		}
		pids = append(pids, pid)
	}

	return pids, scanner.Err()
}

func Kill(exe string) error {
	pids, err := Running(exe)
	if err != nil {
		return err
	}
	for _, pid := range pids {
		output, err := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/T", "/F").CombinedOutput()
		if err != nil {
			return fmt.Errorf("taskkill pid %d failed: %w: %s", pid, err, strings.TrimSpace(string(output)))
		}
		slog.Info("terminated running instance", "pid", pid)
	}

	return nil
}
