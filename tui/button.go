package tui

import tea "charm.land/bubbletea/v2"

type Button struct {
	Label     string
	OnPressed func() tea.Cmd
	HelpText
}

func (b Button) View(focused bool) string {
	if focused {
		return selectedStyle.Render(b.Label)
	}

	return buttonStyle.Render(b.Label)
}

func (b Button) Update(msg tea.Msg) tea.Cmd {
	if !shouldActivate(msg) {
		return nil
	}

	return b.OnPressed()
}
