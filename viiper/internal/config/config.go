// Package config defines the CLI structure and configuration for VIIPER.
package config

import (
	"viiper/internal/cmd"
)

type Log struct {
	Level   string `help:"Log level: trace, debug, info, warn, error" default:"info" env:"VIIPER_LOG_LEVEL"`
	File    string `help:"Log file path (default: none; logs only to console)" env:"VIIPER_LOG_FILE"`
	RawFile string `help:"Raw packet log file path (default: none)" env:"VIIPER_LOG_RAW_FILE"`
}

// CLI is the root command structure for Kong CLI parsing.
type CLI struct {
	Log `embed:"" prefix:"log."`

	Server cmd.Server `cmd:"" help:"Start the VIIPER USB-IP server"`
	Proxy  cmd.Proxy  `cmd:"" help:"Start the VIIPER USB-IP proxy"`
}
