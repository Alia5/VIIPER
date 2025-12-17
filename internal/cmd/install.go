package cmd

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Install sets up VIIPER to run automatically.
type Install struct{}

// Uninstall removes VIIPER startup configuration.
type Uninstall struct{}

func (c *Install) Run(logger *slog.Logger) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	if strings.Contains(exe, "go-build") {
		return errors.New("cannot install from 'go run'")
	}

	return install(logger)
}

func (c *Uninstall) Run(logger *slog.Logger) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	if strings.Contains(exe, "go-build") {
		return errors.New("cannot uninstall from 'go run'")
	}

	return uninstall(logger)
}

func currentExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Abs(exe)
}
