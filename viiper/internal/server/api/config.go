package api

import "time"

// ServerConfig represents the server subcommand configuration.
type ServerConfig struct {
	Addr                        string        `help:"API server listen address" default:":3242" env:"VIIPER_API_ADDR"`
	ConnectionTimeout           time.Duration `kong:"-"`
	DeviceHandlerConnectTimeout time.Duration `help:"Time before auto-cleanup occurs when device handler has no active connection" default:"5s" env:"VIIPER_API_DEVICE_HANDLER_TIMEOUT"`
}
