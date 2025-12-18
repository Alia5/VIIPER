//go:build windows

package api

type platformOpts struct {
	AutoAttachWindowsNative bool `help:"Use native IOCTL instead of usbip.exe for auto-attach" default:"true" env:"VIIPER_API_AUTO_ATTACH_WINDOWS_NATIVE"`
}
