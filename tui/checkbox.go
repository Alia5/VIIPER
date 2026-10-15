package tui

import tea "charm.land/bubbletea/v2"

type Checkbox struct {
	Label   string
	Checked *bool
	HelpText
}

func (c Checkbox) View(focused bool) string {
	mark := "🔳 "
	if *c.Checked {
		mark = "✅ "
	}
	if focused {
		return focusStyle.Render(focusArrow + mark + c.Label)
	}

	return focusBlank + mark + c.Label
}

func (c Checkbox) Update(msg tea.Msg) tea.Cmd {
	if shouldActivate(msg) {
		*c.Checked = !*c.Checked
	}

	return nil
}
