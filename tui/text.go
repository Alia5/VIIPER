package tui

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type Text struct {
	label string
	value *string
	input textinput.Model
	HelpText
}

func NewText(label, help string, value *string) *Text {
	input := textinput.New()
	input.Prompt = ""
	input.SetValue(*value)

	return &Text{
		label:    label,
		value:    value,
		input:    input,
		HelpText: HelpText(help),
	}
}

func (t *Text) View(focused bool) string {
	if focused {
		t.input.Focus()

		return focusStyle.Render(focusArrow+t.label+": ") + t.input.View()
	}
	t.input.Blur()

	return focusBlank + t.label + ": " + t.input.View()
}

func (t *Text) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	t.input, cmd = t.input.Update(msg)
	*t.value = t.input.Value()

	return cmd
}
