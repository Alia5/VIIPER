package tui

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

const (
	optionIndent = "  "
	labelRows    = 1
)

type Select struct {
	Label    string
	Options  []string
	Selected *string
	HelpText

	open        bool
	highlighted int
}

func (s *Select) View(focused bool) string {
	if !focused {
		return focusBlank + s.Label + ": " + *s.Selected + closedArrow
	}
	if !s.open {
		return focusStyle.Render(focusArrow + s.Label + ": " + *s.Selected + closedArrow)
	}
	lines := []string{
		focusStyle.Render(focusArrow + s.Label + ": " + *s.Selected + openArrow),
	}
	for i, option := range s.Options {
		style := buttonStyle
		if i == s.highlighted {
			style = selectedStyle
		}
		lines = append(lines, optionIndent+style.Render(option))
	}

	return strings.Join(lines, "\n")
}

func (s *Select) Update(msg tea.Msg) tea.Cmd {
	if !s.open {
		if shouldActivate(msg) {
			s.open = true
			s.highlighted = max(slices.Index(s.Options, *s.Selected), 0)
		}

		return nil
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keyNext):
			s.highlighted = (s.highlighted + 1) % len(s.Options)
		case key.Matches(msg, keyPrevious):
			s.highlighted = (s.highlighted + len(s.Options) - 1) % len(s.Options)
		case key.Matches(msg, keyConfirm):
			*s.Selected = s.Options[s.highlighted]
			s.open = false
		case key.Matches(msg, keyCancel):
			s.open = false
		}
	case tea.MouseMotionMsg:
		if option := msg.Y - labelRows; option >= 0 && option < len(s.Options) {
			s.highlighted = option
		}
	case tea.MouseClickMsg:
		if msg.Button != tea.MouseLeft {
			return nil
		}
		if option := msg.Y - labelRows; option >= 0 && option < len(s.Options) {
			*s.Selected = s.Options[option]
		}
		s.open = false
	}

	return nil
}
