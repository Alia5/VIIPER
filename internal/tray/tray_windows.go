//go:build windows

package tray

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"

	"fyne.io/systray"
	"golang.org/x/sys/windows/registry"
)

//go:embed icon.ico
var trayIcon []byte

const (
	runKeyPath  = `Software\Microsoft\Windows\CurrentVersion\Run`
	runValueKey = "VIIPER"
)

func Run(shutdown func()) {
	systray.Run(func() {
		systray.SetIcon(trayIcon)
		systray.SetTooltip("VIIPER")

		version := readVersion()
		infoStr := fmt.Sprintf("VIIPER - %s", version)
		versionItem := systray.AddMenuItem(infoStr, infoStr)
		versionItem.Disable()

		systray.AddSeparator()

		autoStartItem := systray.AddMenuItemCheckbox("Run at startup", "", autoStartEnabled())

		systray.AddSeparator()

		exitItem := systray.AddMenuItem("Quit", "Exit VIIPER")

		go func() {
			for {
				select {
				case <-autoStartItem.ClickedCh:
					if toggleAutoStart() {
						autoStartItem.Check()
					} else {
						autoStartItem.Uncheck()
					}
				case <-exitItem.ClickedCh:
					systray.Quit()
					shutdown()
					return
				}
			}
		}()
	}, func() {})
}

func readVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		v := info.Main.Version
		if v != "" && v != "(devel)" {
			return v
		}
	}
	return "dev"
}

func autoStartEnabled() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()
	_, _, err = key.GetStringValue(runValueKey)
	return err == nil
}

func toggleAutoStart() bool {
	if autoStartEnabled() {
		key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
		if err != nil {
			slog.Error("Failed to open registry key", "error", err)
			return true
		}
		defer key.Close()
		_ = key.DeleteValue(runValueKey)
		slog.Info("Auto-start disabled")
		return false
	}
	exe, err := os.Executable()
	if err != nil {
		slog.Error("Failed to get executable path", "error", err)
		return false
	}
	selfPath, err := filepath.EvalSymlinks(exe)
	if err != nil {
		slog.Error("Failed to evaluate symlinks", "error", err)
		return false
	}
	key, _, err := registry.CreateKey(registry.CURRENT_USER, runKeyPath, registry.ALL_ACCESS)
	if err != nil {
		slog.Error("Failed to create registry key", "error", err)
		return false
	}
	defer key.Close()
	value := fmt.Sprintf("\"%s\" server", selfPath)
	if err := key.SetStringValue(runValueKey, value); err != nil {
		slog.Error("Failed to set registry value", "error", err)
		return false
	}
	slog.Info("Auto-start enabled")
	return true
}
