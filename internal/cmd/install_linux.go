//go:build linux

package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	serviceName = "viiper.service"
	servicePath = "/etc/systemd/system/viiper.service"
)

func install(logger *slog.Logger) error {
	exePath, err := currentExecutable()
	if err != nil {
		return err
	}

	unit := systemdUnitContent(exePath)
	if err := os.WriteFile(servicePath, []byte(unit), 0o644); err != nil {
		return err
	}

	steps := [][]string{
		{"daemon-reload"},
		{"enable", serviceName},
		{"restart", serviceName},
	}

	for _, args := range steps {
		if err := runSystemctl(args...); err != nil {
			return err
		}
	}

	logger.Info("VIIPER systemd service installed", "path", servicePath, "exe", exePath)
	return nil
}

func uninstall(logger *slog.Logger) error {
	var errs []error

	if err := runSystemctl("stop", serviceName); err != nil {
		errs = append(errs, err)
	}
	if err := runSystemctl("disable", serviceName); err != nil {
		errs = append(errs, err)
	}

	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		errs = append(errs, err)
	}

	if err := runSystemctl("daemon-reload"); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	logger.Info("VIIPER systemd service removed", "path", servicePath)
	return nil
}

func systemdUnitContent(exePath string) string {
	workingDir := filepath.Dir(exePath)
	return fmt.Sprintf(`[Unit]
Description=VIIPER server
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=%q server
WorkingDirectory=%q
Restart=on-failure

[Install]
WantedBy=multi-user.target
`, exePath, workingDir)
}

func runSystemctl(args ...string) error {
	cmd := exec.Command("systemctl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl %s failed: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return nil
}
