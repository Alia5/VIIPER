//go:build !windows

package tray

import "context"

func Run(ctx context.Context, shutdown func()) {}
