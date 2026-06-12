//go:build !windows

package api

type PlatformOpts struct {
	AutoAttachWindowsNative bool `kong:"-"`
}
