package pages

import (
	"os"

	"github.com/Alia5/VIIPER/internal/install"
	"github.com/Alia5/VIIPER/tui"

	tea "charm.land/bubbletea/v2"
)

func serverButton(launchCommand LaunchCommand) (tui.Button, error) {
	exe, err := os.Executable()
	if err != nil {
		return tui.Button{}, err
	}
	running, err := install.Running(exe)
	if err != nil {
		return tui.Button{}, err
	}
	label := "Run VIIPER"
	if len(running) > 0 {
		label = "Restart VIIPER"
	}

	return tui.Button{
		Label:    label,
		HelpText: "Start the VIIPER USB-IP server",
		OnPressed: func() tea.Cmd {
			err := install.Kill(exe)
			if err != nil {
				return tui.Fail(err)
			}

			return launchCommand("server")
		},
	}, nil
}
