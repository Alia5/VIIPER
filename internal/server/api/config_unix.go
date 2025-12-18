//go:build !windows

package api

type platformOpts struct {
	AutoAttachWindowsNative bool `kong:"-"`
}
