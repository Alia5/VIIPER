package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/Alia5/VIIPER/internal/tui"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type field struct {
	label string
	flag  string
	input textinput.Model
}

type command struct {
	label   string
	args    []string
	confirm string
	fields  []field
	before  func() error
}

var (
	configInit = command{
		label: "Generate a config template",
		args:  []string{"config", "init"},
		fields: []field{
			textField("Command (server or proxy)", "", "server"),
			textField("Format (json, yaml or toml)", "--format", "json"),
			textField("Output file (empty for the current directory)", "--output", ""),
		},
	}
	proxy = command{
		label: "Run VIIPER USB-IP proxy",
		args:  []string{"proxy"},
		fields: []field{
			textField("Upstream USB-IP server address", "--upstream-addr", ""),
			textField("Listen address", "--listen-addr", ":3241"),
		},
	}
)

func textField(label, flag, value string) field {
	input := textinput.New()
	input.SetValue(value)
	input.Focus()

	return field{
		label: label,
		flag:  flag,
		input: input,
	}
}

func launch() {
	commands, err := commands()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var chosen *command
	ui := &tui.Wizard{}
	form := func(c *command) tea.Cmd {
		if len(c.fields) == 0 {
			chosen = c

			return tea.Quit
		}
		var rows []tui.Row
		for i := range c.fields {
			rows = append(rows, tui.Text(c.fields[i].label, &c.fields[i].input))
		}
		ui.Pages = append(ui.Pages, tui.Page{
			Title: c.label,
			Text:  "Fill in the fields, then select OK.",
			Hint:  "↑/↓ Tab: move   ←/→: switch   Enter: confirm   Esc: quit",
			Rows: append(rows, tui.Buttons(tui.Button{
				Label: "OK",
				Press: func() tea.Cmd {
					chosen = c

					return tea.Quit
				},
			}, tui.Button{
				Label: "Back",
				Press: ui.Back,
			})),
		})

		return ui.Next()
	}
	var items []tui.Button
	for i := range commands {
		c := &commands[i]
		items = append(items, tui.Button{
			Label: c.label,
			Press: func() tea.Cmd {
				if c.confirm == "" {
					return form(c)
				}
				ui.Pages = append(ui.Pages, tui.Page{
					Title: c.label,
					Text:  c.confirm,
					Hint:  "←/→: switch   Enter: confirm   Esc: quit",
					Rows: []tui.Row{
						tui.Buttons(tui.Button{
							Label: "No",
							Press: ui.Back,
						}, tui.Button{
							Label: "Yes",
							Press: func() tea.Cmd {
								return form(c)
							},
						}),
					},
				})

				return ui.Next()
			},
		})
	}
	menu := tui.Menu(items...)
	ui.Pages = []tui.Page{
		{
			Hint: "↑/↓ j/k: move   Enter: select   Esc: quit",
			Rows: []tui.Row{
				{
					Items: menu.Items,
					View: func(focus int) string {
						return mergeArtWithHelp(normalizeLineEndings(asciiBrailleColoredSmall), normalizeLineEndings(Description()+"\n"+menu.View(focus)))
					},
					Key: menu.Key,
				},
			},
		},
	}
	ui.Forward = func(msg tea.Msg) tea.Cmd {
		var cmds []tea.Cmd
		for i := range commands {
			for j := range commands[i].fields {
				var cmd tea.Cmd
				commands[i].fields[j].input, cmd = commands[i].fields[j].input.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

		return tea.Batch(cmds...)
	}
	_, err = tea.NewProgram(ui).Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if chosen == nil {
		os.Exit(0)
	}
	if chosen.before != nil {
		err = chosen.before()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	args := slices.Clone(chosen.args)
	for _, f := range chosen.fields {
		value := f.input.Value()
		switch {
		case f.flag == "":
			args = append(args, value)
		case value != "":
			args = append(args, f.flag+"="+value)
		}
	}
	os.Args = slices.Insert(os.Args, 1, args...)
}
