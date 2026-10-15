package pages

import (
	"github.com/Alia5/VIIPER/tui"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/kong"
)

func uninstall(parser *kong.Kong, launchCommand LaunchCommand) tea.Cmd {
	return tui.Show(tui.Page{
		Title: "Uninstall VIIPER",
		Text:  "Really uninstall VIIPER?",
		Hint:  "←/→: switch   Enter: confirm   Esc: quit",
		Buttons: []tui.Button{
			{
				Label: "No",
				OnPressed: func() tea.Cmd {
					return Menu(parser, launchCommand)
				},
			},
			{
				Label: "Yes",
				OnPressed: func() tea.Cmd {
					return launchCommand("uninstall")
				},
			},
		},
	})
}
