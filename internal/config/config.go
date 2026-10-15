// Package config defines the CLI structure and configuration for VIIPER.
package config

import (
	"github.com/Alia5/VIIPER/cmd/install"
	"github.com/Alia5/VIIPER/cmd/uninstall"
	"github.com/Alia5/VIIPER/internal/cmd"
)

type UpdateNotify string

const (
	UpdateNotifyNone       UpdateNotify = "none"
	UpdateNotifyStable     UpdateNotify = "stable"
	UpdateNotifyPrerelease UpdateNotify = "prerelease"
)

type Log struct {
	Level   string `aliases:"l" help:"Log level: trace, debug, info, warn, error" default:"info" env:"VIIPER_LOG_LEVEL"`
	File    string `help:"Log file path (default: none; logs only to console)" env:"VIIPER_LOG_FILE"`
	RawFile string `help:"Raw packet log file path (default: none)" env:"VIIPER_LOG_RAW_FILE"`
}

// CLI is the root command structure for Kong CLI parsing.
type CLI struct {
	// Global
	ConfigPath   string       `help:"Path to configuration file (json|yaml|toml)" name:"config" env:"VIIPER_CONFIG"`
	UpdateNotify UpdateNotify `help:"Update notification level: none, stable, prerelease" default:"stable" env:"VIIPER_UPDATE_NOTIFY"`
	Log          `embed:"" prefix:"log."`
	codegenCommand

	TUI    cmd.TUI    `cmd:"" help:"Run an interactive TUI"`
	Server cmd.Server `cmd:"" help:"Start the VIIPER USB-IP server"`
	Proxy  cmd.Proxy  `cmd:"" help:"Start the VIIPER USB-IP proxy"`

	Config    cmd.ConfigCommand   `cmd:"" help:"Manage configuration files"`
	Install   install.Installer   `cmd:"" help:"Add the current VIIPER executable to system startup and runs it (creates a Systemd service on Linux)"`
	Uninstall uninstall.Uninstall `cmd:"" help:"Remove any VIIPER system startup configuration / Systemd service"`
}
