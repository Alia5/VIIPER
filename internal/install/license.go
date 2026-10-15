package install

import (
	"github.com/Alia5/VIIPER/tui"

	tea "charm.land/bubbletea/v2"
)

const licenseText = `VIIPER is free software licensed under the terms of the GPLv3 license.
It uses third-party dependencies subject to their own respective licenses.

You can find the licenses below.

`

func (w *wizard) license() tea.Cmd {
	return tui.Show(tui.Page{
		Title: "License Agreement",
		Text:  licenseText + w.licenses,
		Hint:  "PgUp/PgDn: scroll   ←/→: switch   Enter: confirm   Esc: quit",
		Buttons: []tui.Button{
			{
				Label:     "Ok",
				OnPressed: w.options,
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
