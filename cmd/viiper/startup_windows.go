//go:build windows

package main

import (
	"log/slog"
	"os"

	"github.com/Alia5/VIIPER/internal/util"
)

func init() {
	if util.IsRunFromGUI() {
		args := os.Args
		if len(args) < 2 || args[1] != "server" {
			slog.Info("Detected GUI startup, injecting 'server' argument")
			slog.Warn("Run from a CLI for more options!")
			newArgs := make([]string, 0, len(args)+1)
			newArgs = append(newArgs, args[0], "server")
			newArgs = append(newArgs, args[1:]...)
			os.Args = newArgs
		}
	}
}
