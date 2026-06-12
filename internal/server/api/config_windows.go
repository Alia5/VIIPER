//go:build windows

package api

type PlatformOpts struct {
	AutoAttachWindowsNative bool `default:"true" help:"Use native IOCTL instead of usbip.exe for auto-attach"  env:"VIIPER_API_AUTO_ATTACH_WINDOWS_NATIVE"`
}
