package pages

import (
	"strings"

	"github.com/Alia5/VIIPER/meta"
	"github.com/Alia5/VIIPER/tui"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/kong"
)

type LaunchCommand func(line ...string) tea.Cmd

func Menu(parser *kong.Kong, launchCommand LaunchCommand) tea.Cmd {
	return func() tea.Msg {
		server, err := serverButton(launchCommand)
		if err != nil {
			return err
		}

		return tui.Page{
			Text: strings.Trim(meta.Description(), " \n"),
			Hint: "↑/↓: move   Enter: select   Esc: quit",
			Elements: []tui.Element{
				server,
				tui.Button{
					Label: "Show help information",
					OnPressed: func() tea.Cmd {
						return help(parser, launchCommand)
					},
				},
				tui.Button{
					Label:    "Generate a config template",
					HelpText: "Write a server or proxy configuration file with every default",
					OnPressed: func() tea.Cmd {
						return configInit(parser, launchCommand)
					},
				},
				tui.Button{
					Label:    "Run VIIPER USB-IP proxy",
					HelpText: "Start the VIIPER USB-IP proxy",
					OnPressed: func() tea.Cmd {
						return proxy(parser, launchCommand)
					},
				},
				tui.Button{
					Label:    "Install/Update VIIPER",
					HelpText: "Add VIIPER to system startup and run it (creates a systemd service on Linux)",
					OnPressed: func() tea.Cmd {
						return launchCommand("install")
					},
				},
				tui.Button{
					Label:    "Uninstall VIIPER",
					HelpText: "Remove the VIIPER system startup configuration / systemd service",
					OnPressed: func() tea.Cmd {
						return uninstall(parser, launchCommand)
					},
				},
				tui.Button{
					Label:    "Exit",
					HelpText: "Exit the launcher",
					OnPressed: func() tea.Cmd {
						return tea.Quit
					},
				},
			},
		}
	}
}
