package install

import (
	"github.com/Alia5/VIIPER/internal/tui"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (w *wizard) options() tui.Page {
	rows := []tui.Row{
		tui.Text("Directory", &w.dir),
	}
	for i := range w.steps {
		s := &w.steps[i]
		if !s.optional {
			continue
		}
		rows = append(rows, tui.Row{
			Items: 1,
			View: func(focus int) string {
				mark := "[ ] "
				if s.enabled {
					mark = "[x] "
				}

				return lipgloss.NewStyle().Reverse(focus >= 0).Render(mark + s.label)
			},
			Key: func(msg tea.KeyPressMsg, _ int) (tea.Cmd, bool) {
				if !key.Matches(msg, tui.KeyConfirm) {
					return nil, false
				}
				s.enabled = !s.enabled

				return nil, true
			},
		})
	}

	return tui.Page{
		Title: "Select Destination Location and Additional Tasks",
		Text:  "Where should VIIPER be installed, and which additional tasks should be performed?",
		Hint:  "↑/↓ Tab: move   Space: toggle   ←/→: switch   Enter: confirm   Esc: quit",
		Rows: append(rows, tui.Buttons(tui.Button{
			Label: "Install",
			Press: w.start,
		}, tui.Button{
			Label: "Cancel",
			Press: tui.Quit,
		})),
	}
}
