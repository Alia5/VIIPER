package install

import (
	"strings"

	"github.com/Alia5/VIIPER/internal/tui"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	barTrack = " ░"
	barThumb = " █"
)

var keyScroll = key.NewBinding(key.WithKeys("up", "down", "pgup", "pgdown", "home", "end"))

func (w *wizard) licensePage() tui.Page {
	text := `VIIPER is free software licensed under the terms of the GPLv3 license.
It uses third-party dependencies subject to their own respective licenses.

You can find the licenses below.`
	w.license.SoftWrap = true
	w.license.SetWidth(lipgloss.Width(text))

	return tui.Page{
		Title: "License Agreement",
		Text:  text,
		Hint:  "↑/↓: scroll   Tab: move   ←/→: switch   Enter: confirm   Esc: quit",
		Rows: []tui.Row{
			{
				Items: 1,
				View: func(int) string {
					height := w.license.Height()
					thumb := max(height*height/max(w.license.TotalLineCount(), 1), 1)
					top := int(float64(height-thumb) * w.license.ScrollPercent())
					var bar []string
					for i := range height {
						if i >= top && i < top+thumb {
							bar = append(bar, barThumb)

							continue
						}
						bar = append(bar, barTrack)
					}

					return lipgloss.JoinHorizontal(lipgloss.Top, w.license.View(), strings.Join(bar, "\n"))
				},
				Key: func(msg tea.KeyPressMsg, _ int) (tea.Cmd, bool) {
					if !key.Matches(msg, keyScroll) {
						return nil, false
					}
					var cmd tea.Cmd
					w.license, cmd = w.license.Update(msg)

					return cmd, true
				},
			},
			tui.Buttons(tui.Button{
				Label: "Ok",
				Press: w.ui.Next,
			}, tui.Button{
				Label: "Cancel",
				Press: tui.Quit,
			}),
		},
	}
}
