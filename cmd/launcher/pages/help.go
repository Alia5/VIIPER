package pages

import (
	"bytes"
	"strings"

	"github.com/Alia5/VIIPER/tui"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/kong"
)

func help(parser *kong.Kong, launchCommand LaunchCommand) tea.Cmd {
	usage, err := kong.Trace(parser, []string{
		"--help",
	})
	if err != nil {
		return tui.Fail(err)
	}
	var text bytes.Buffer
	app := *usage.Kong
	app.Stdout = &text
	usage.Kong = &app
	err = kong.DefaultHelpPrinter(kong.HelpOptions{
		ValueFormatter: kong.DefaultHelpValueFormatter,
	}, usage)
	if err != nil {
		return tui.Fail(err)
	}

	return tui.Show(tui.Page{
		Text: strings.TrimRight(text.String(), "\n"),
		Hint: "PgUp/PgDn: scroll   Enter: back   Esc: quit",
		Buttons: []tui.Button{
			{
				Label: "Back",
				OnPressed: func() tea.Cmd {
					return Menu(parser, launchCommand)
				},
			},
		},
	})
}
