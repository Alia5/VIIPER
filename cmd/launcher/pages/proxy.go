package pages

import (
	"github.com/Alia5/VIIPER/tui"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/kong"
)

func proxy(parser *kong.Kong, launchCommand LaunchCommand) tea.Cmd {
	values, err := commandValues(parser, "proxy")
	if err != nil {
		return tui.Fail(err)
	}

	options := []*launchOption{
		valueToOption(values["upstream-addr"]),
		valueToOption(values["listen-addr"]),
	}

	return tui.Show(tui.Page{
		Title:    "Run VIIPER USB-IP proxy",
		Hint:     "↑/↓ Tab: move   Space: toggle   Enter: confirm   Esc: quit",
		Elements: optionsToElements(options),
		Buttons: []tui.Button{
			{
				Label: "OK",
				OnPressed: func() tea.Cmd {
					return launchCommand(commandLine("proxy", options)...)
				},
			},
			{
				Label: "Back",
				OnPressed: func() tea.Cmd {
					return Menu(parser, launchCommand)
				},
			},
		},
	})
}
