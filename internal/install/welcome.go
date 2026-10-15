package install

import (
	"fmt"

	"github.com/Alia5/VIIPER/tui"

	tea "charm.land/bubbletea/v2"
)

const (
	welcomeTitle = "Welcome to the VIIPER Setup Wizard"
	welcomeText  = `This will install VIIPER %s on your computer.

Select Next to continue, or Cancel to exit Setup.`

	uninstallTitle = "VIIPER Uninstall"
	uninstallText  = "Are you sure you want to completely remove VIIPER %s and all of its components?"
)

func (w *wizard) welcome() tea.Cmd {
	if w.uninstall {
		return tui.Show(tui.Page{
			Title: uninstallTitle,
			Text:  fmt.Sprintf(uninstallText, w.target.Version),
			Hint:  "←/→: switch   Enter: confirm   Esc: quit",
			Buttons: []tui.Button{
				{
					Label:     "Yes",
					OnPressed: w.start,
				},
				{
					Label: "No",
					OnPressed: func() tea.Cmd {
						return tea.Quit
					},
				},
			},
		})
	}

	return tui.Show(tui.Page{
		Title: welcomeTitle,
		Text:  fmt.Sprintf(welcomeText, w.target.Version),
		Hint:  "←/→: switch   Enter: confirm   Esc: quit",
		Buttons: []tui.Button{
			{
				Label:     "Next",
				OnPressed: w.license,
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
