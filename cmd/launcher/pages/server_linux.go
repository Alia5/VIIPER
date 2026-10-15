package pages

import (
	"github.com/Alia5/VIIPER/tui"

	tea "charm.land/bubbletea/v2"
)

func serverButton(launchCommand LaunchCommand) (tui.Button, error) {
	return tui.Button{
		Label:    "Run VIIPER",
		HelpText: "Start the VIIPER USB-IP server",
		OnPressed: func() tea.Cmd {
			return launchCommand("server")
		},
	}, nil
}
