package install

import (
	"fmt"
	"strings"

	"github.com/Alia5/VIIPER/internal/tui"

	"charm.land/lipgloss/v2"
)

const (
	welcomeText = `Welcome to the VIIPER Setup Wizard

This will install VIIPER %s on your computer.

Select Next to continue, or Cancel to exit Setup.`

	uninstallText = `VIIPER Uninstall

Are you sure you want to completely remove VIIPER %s and all of its components?`
)

func (w *wizard) welcome(banner string) tui.Page {
	choice := tui.Buttons(tui.Button{
		Label: "Next",
		Press: w.ui.Next,
	}, tui.Button{
		Label: "Cancel",
		Press: tui.Quit,
	})
	text := fmt.Sprintf(welcomeText, w.target.Version)
	if w.uninstall {
		choice = tui.Buttons(tui.Button{
			Label: "Yes",
			Press: w.start,
		}, tui.Button{
			Label: "No",
			Press: tui.Quit,
		})
		text = fmt.Sprintf(uninstallText, w.target.Version)
	}

	return tui.Page{
		Hint: "←/→: switch   Enter: confirm   Esc: quit",
		Rows: []tui.Row{
			{
				Items: choice.Items,
				View: func(focus int) string {
					return lipgloss.JoinHorizontal(lipgloss.Top, banner, "  "+strings.ReplaceAll(text+"\n\n"+choice.View(focus), "\n", "\n  "))
				},
				Key: choice.Key,
			},
		},
	}
}
