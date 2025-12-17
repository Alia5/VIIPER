//go:build windows

package cmd

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

	"golang.org/x/sys/windows/registry"
)

const (
	runKeyPath  = `Software\Microsoft\Windows\CurrentVersion\Run`
	runValueKey = "VIIPER"
)

func install(logger *slog.Logger) error {
	exePath, err := currentExecutable()
	if err != nil {
		return err
	}

	previousExe, err := currentAutorunExe()
	if err != nil {
		return err
	}

	value := fmt.Sprintf("%q server", exePath)
	key, _, err := registry.CreateKey(registry.CURRENT_USER, runKeyPath, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer key.Close()

	if err := key.SetStringValue(runValueKey, value); err != nil {
		return err
	}

	if previousExe != "" {
		if err := killProcessesByExe(previousExe, logger); err != nil {
			return fmt.Errorf("failed to stop previous autorun instance: %w", err)
		}
	}

	if err := exec.Command(exePath, "server").Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	logger.Info("VIIPER install completed for Windows autorun", "exe", exePath)
	return nil
}

func uninstall(logger *slog.Logger) error {
	autorunExe, err := currentAutorunExe()
	if err != nil {
		return err
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		if !errors.Is(err, registry.ErrNotExist) {
			return err
		}
	} else {
		defer key.Close()

		if err := key.DeleteValue(runValueKey); err != nil {
			if !errors.Is(err, registry.ErrNotExist) {
				return err
			}
		}
	}

	if autorunExe != "" {
		if err := killProcessesByExe(autorunExe, logger); err != nil {
			return fmt.Errorf("failed to stop autorun instance: %w", err)
		}
	}

	logger.Info("VIIPER autorun entry removed")
	return nil
}

func currentAutorunExe() (string, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	defer key.Close()

	val, _, err := key.GetStringValue(runValueKey)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return "", nil
		}
		return "", err
	}

	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return "", nil
	}

	if strings.HasPrefix(trimmed, "\"") {
		trimmed = strings.TrimPrefix(trimmed, "\"")
		if end := strings.Index(trimmed, "\""); end >= 0 {
			trimmed = trimmed[:end]
		}
	}

	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return "", nil
	}

	path := fields[0]
	if path == "" {
		return "", nil
	}
	return filepath.Clean(path), nil
}

func killProcessesByExe(target string, logger *slog.Logger) error {
	target = filepath.Clean(target)
	if target == "" {
		return nil
	}

	script := fmt.Sprintf(
		"$ErrorActionPreference='SilentlyContinue';$t='%s';Get-CimInstance Win32_Process | Where-Object { $_.ExecutablePath -eq $t } | Select-Object -ExpandProperty ProcessId",
		strings.ReplaceAll(target, "'", "''"),
	)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("process query failed: %w: %s", err, strings.TrimSpace(string(output)))
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	var pids []int
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		if err == nil {
			pids = append(pids, pid)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if len(pids) == 0 {
		return nil
	}

	self := os.Getpid()
	for _, pid := range pids {
		if pid == self {
			continue
		}
		cmd := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/T", "/F")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("taskkill pid %d failed: %w: %s", pid, err, strings.TrimSpace(string(output)))
		}
		logger.Info("terminated autorun instance", "pid", pid)
	}

	return nil
}
