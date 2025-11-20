package api

import "time"

// ServerConfig represents the server subcommand configuration.
type ServerConfig struct {
	Addr                        string        `help:"API server listen address" default:":3242" env:"VIIPER_API_ADDR"`
	DeviceHandlerConnectTimeout time.Duration `help:"Time before auto-cleanup occurs when device handler has no active connection" default:"5s" env:"VIIPER_API_DEVICE_HANDLER_TIMEOUT"`
	AutoAttachLocalClient       bool          `help:"Controls usbip-client on localhost to auto-attach devices added to the virtual bus" default:"true" env:"VIIPER_API_AUTO_ATTACH_LOCAL_CLIENT"`
	ConnectionTimeout           time.Duration `kong:"-"`
}
