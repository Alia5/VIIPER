package pages

import (
	"github.com/Alia5/VIIPER/tui"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/kong"
)

func configInit(parser *kong.Kong, launchCommand LaunchCommand) tea.Cmd {
	values, err := commandValues(parser, "config init")
	if err != nil {
		return tui.Fail(err)
	}

	command := valueToOption(values["command"])
	command.text = "server"

	options := []*launchOption{
		command,
		valueToOption(values["format"]),
		valueToOption(values["output"]),
	}

	return tui.Show(tui.Page{
		Title:    "Generate a config template",
		Hint:     "↑/↓ Tab: move   Space: toggle   Enter: confirm   Esc: quit",
		Elements: optionsToElements(options),
		Buttons: []tui.Button{
			{
				Label: "OK",
				OnPressed: func() tea.Cmd {
					return launchCommand(commandLine("config init", options)...)
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
