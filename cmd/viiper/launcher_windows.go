package main

import (
	"os"

	"github.com/Alia5/VIIPER/internal/install"
)

func commands() ([]command, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	running, err := install.Running(exe)
	if err != nil {
		return nil, err
	}
	run := "Run VIIPER"
	if len(running) > 0 {
		run = "Restart VIIPER"
	}

	return []command{
		{
			label: run,
			args:  []string{"server"},
			before: func() error {
				return install.Kill(exe)
			},
		},
		configInit,
		proxy,
		{
			label: "Install/Update VIIPER",
			args:  []string{"install"},
		},
		{
			label:   "Uninstall VIIPER",
			args:    []string{"uninstall"},
			confirm: "Really uninstall VIIPER?",
		},
	}, nil
}
