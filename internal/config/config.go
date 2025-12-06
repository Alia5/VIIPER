// Package config defines the CLI structure and configuration for VIIPER.
package config

import (
	"github.com/Alia5/VIIPER/internal/cmd"
)

type Log struct {
	Level   string `help:"Log level: trace, debug, info, warn, error" default:"info" env:"VIIPER_LOG_LEVEL"`
	File    string `help:"Log file path (default: none; logs only to console)" env:"VIIPER_LOG_FILE"`
	RawFile string `help:"Raw packet log file path (default: none)" env:"VIIPER_LOG_RAW_FILE"`
}

// CLI is the root command structure for Kong CLI parsing.
type CLI struct {
	// Global
	ConfigPath string `help:"Path to configuration file (json|yaml|toml)" name:"config" env:"VIIPER_CONFIG"`
	Log        `embed:"" prefix:"log."`

	Server cmd.Server `cmd:"" help:"Start the VIIPER USB-IP server"`
	Proxy  cmd.Proxy  `cmd:"" help:"Start the VIIPER USB-IP proxy"`

	Config  cmd.ConfigCommand `cmd:"" help:"Manage configuration files"`
	Codegen cmd.Codegen       `cmd:"" help:"Generate client libraries from server code"`
}
