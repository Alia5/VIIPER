package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Row struct {
	Items int
	View  func(focus int) string
	Key   func(msg tea.KeyPressMsg, focus int) (tea.Cmd, bool)
}

type Button struct {
	Label string
	Press func() tea.Cmd
}

type Page struct {
	Title string
	Text  string
	Hint  string
	Rows  []Row
}

type Wizard struct {
	Pages   []Page
	Index   int
	Cursor  int
	Focus   int
	Width   int
	Height  int
	Forward func(tea.Msg) tea.Cmd
	Resize  func()
}

var (
	KeyQuit    = key.NewBinding(key.WithKeys("esc", "ctrl+c"))
	KeyNav     = key.NewBinding(key.WithKeys("tab", "shift+tab", "up", "down", "enter"))
	KeyNext    = key.NewBinding(key.WithKeys("tab", "down", "j"))
	KeyPrev    = key.NewBinding(key.WithKeys("shift+tab", "up", "k"))
	KeyRight   = key.NewBinding(key.WithKeys("right", "l"))
	KeyLeft    = key.NewBinding(key.WithKeys("left", "h"))
	KeyConfirm = key.NewBinding(key.WithKeys("enter", "space"))
)

func Quit() tea.Cmd {
	return tea.Quit
}

func Buttons(all ...Button) Row {
	return choice("  ", all)
}

func Menu(all ...Button) Row {
	return choice("\n", all)
}

func choice(separator string, all []Button) Row {
	return Row{
		Items: len(all),
		View: func(focus int) string {
			var parts []string
			for i, b := range all {
				parts = append(parts, lipgloss.NewStyle().Reverse(i == focus).Render("[ "+b.Label+" ]"))
			}

			return strings.Join(parts, separator)
		},
		Key: func(msg tea.KeyPressMsg, focus int) (tea.Cmd, bool) {
			if !key.Matches(msg, KeyConfirm) {
				return nil, false
			}

			return all[focus].Press(), true
		},
	}
}

func Text(label string, input *textinput.Model) Row {
	return Row{
		Items: 1,
		View: func(focus int) string {
			return lipgloss.NewStyle().Reverse(focus >= 0).Render(label + ": " + input.View())
		},
		Key: func(msg tea.KeyPressMsg, _ int) (tea.Cmd, bool) {
			if key.Matches(msg, KeyNav) {
				return nil, false
			}
			var cmd tea.Cmd
			*input, cmd = input.Update(msg)

			return cmd, true
		},
	}
}

func (w *Wizard) Next() tea.Cmd {
	w.Index++
	w.first()
	if w.Resize != nil {
		w.Resize()
	}

	return nil
}

func (w *Wizard) Back() tea.Cmd {
	w.Pages = w.Pages[:w.Index]
	w.Index--
	w.first()

	return nil
}

func (w *Wizard) first() {
	rows := w.Pages[w.Index].Rows
	w.Cursor = len(rows) - 1
	w.Focus = rows[w.Cursor].Items
	w.move(1)
}

func (w *Wizard) move(step int) {
	rows := w.Pages[w.Index].Rows
	items := 0
	for _, r := range rows {
		items += r.Items
	}
	if items == 0 {
		return
	}
	w.Focus += step
	for w.Focus < 0 || w.Focus >= rows[w.Cursor].Items {
		w.Cursor = (w.Cursor + step + len(rows)) % len(rows)
		w.Focus = max(-step, 0) * (rows[w.Cursor].Items - 1)
	}
}

func (w *Wizard) Init() tea.Cmd {
	return nil
}

func (w *Wizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.Width = msg.Width
		w.Height = msg.Height
		if w.Resize != nil {
			w.Resize()
		}

		return w, nil
	case tea.KeyPressMsg:
		if key.Matches(msg, KeyQuit) {
			return w, tea.Quit
		}
		row := w.Pages[w.Index].Rows[w.Cursor]
		if row.Items > 0 {
			cmd, consumed := row.Key(msg, w.Focus)
			if consumed {
				return w, cmd
			}
		}
		switch {
		case key.Matches(msg, KeyNext):
			w.move(1)
		case key.Matches(msg, KeyPrev):
			w.move(-1)
		case key.Matches(msg, KeyRight):
			w.Focus = min(w.Focus+1, row.Items-1)
		case key.Matches(msg, KeyLeft):
			w.Focus = max(w.Focus-1, 0)
		}

		return w, nil
	}
	if w.Forward != nil {
		return w, w.Forward(msg)
	}

	return w, nil
}

func (w *Wizard) View() tea.View {
	p := w.Pages[w.Index]
	var b strings.Builder
	if p.Title != "" {
		fmt.Fprintf(&b, "%s\n\n%s\n\n", p.Title, p.Text)
	}
	for i, r := range p.Rows {
		focus := -1
		if i == w.Cursor {
			focus = w.Focus
		}
		fmt.Fprintln(&b, r.View(focus))
	}
	fmt.Fprintf(&b, "\n%s", p.Hint)
	v := tea.NewView(b.String())
	v.AltScreen = true

	return v
}
