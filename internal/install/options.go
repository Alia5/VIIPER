package install

import (
	"github.com/Alia5/VIIPER/tui"

	tea "charm.land/bubbletea/v2"
)

func (w *wizard) options() tea.Cmd {
	elements := []tui.Element{
		tui.NewText("Directory", w.flags["install-dir"].Help, &w.dir),
	}
	for i := range w.steps {
		s := &w.steps[i]
		switch {
		case s.option != "" && s.optional:
			flag := w.flags[s.option]
			elements = append(elements, tui.Checkbox{
				Label:    s.label,
				Checked:  flag.Target.Addr().Interface().(*bool),
				HelpText: tui.HelpText(flag.Help),
			})
		case s.optional:
			elements = append(elements, tui.Checkbox{
				Label:    s.label,
				Checked:  &s.enabled,
				HelpText: tui.HelpText(s.help),
			})
		}
	}

	return tui.Show(tui.Page{
		Title:    "Select Destination Location and Additional Tasks",
		Text:     "Where should VIIPER be installed, and which additional tasks should be performed?",
		Hint:     "↑/↓ Tab: move   Space: toggle   ←/→: switch   Enter: confirm   Esc: quit",
		Elements: elements,
		Buttons: []tui.Button{
			{
				Label:     "Install",
				OnPressed: w.start,
			},
			{
				Label: "Cancel",
				OnPressed: func() tea.Cmd {
					return tea.Quit
				},
			},
		},
	})
}
