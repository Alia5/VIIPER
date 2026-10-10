package install

import (
	"path/filepath"
	"strings"

	"github.com/Alia5/VIIPER/internal/tui"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type wizard struct {
	ui        tui.Wizard
	target    Target
	uninstall bool
	steps     []step
	current   int
	license   viewport.Model
	dir       textinput.Model
	spin      spinner.Model
}

func TUI(t Target, uninstall bool, banner string) error {
	w := &wizard{
		target:    t,
		uninstall: uninstall,
		steps:     installSteps,
		license:   viewport.New(),
		dir:       textinput.New(),
		spin:      spinner.New(spinner.WithSpinner(spinner.Dot)),
	}
	w.ui.Forward = w.forward
	w.ui.Resize = w.resize
	if uninstall {
		w.steps = uninstallSteps
		w.ui.Pages = []tui.Page{
			w.welcome(banner),
		}
	} else {
		w.dir.SetValue(filepath.Dir(t.Exe))
		w.dir.Focus()
		licenses, err := bundle.ReadFile("bundle/licenses.txt")
		if err != nil {
			return err
		}
		w.license.SetContent(string(licenses))
		w.ui.Pages = []tui.Page{
			w.welcome(banner),
			w.licensePage(),
			w.options(),
		}
	}
	_, err := tea.NewProgram(&w.ui).Run()
	if err != nil {
		return err
	}
	if w.current < len(w.steps) {
		return w.steps[w.current].err
	}

	return nil
}

func (w *wizard) resize() {
	w.license.SetWidth(min(w.license.Width(), w.ui.Width-lipgloss.Width(barTrack)))
	w.license.SetHeight(0)
	w.license.SetHeight(w.ui.Height - strings.Count(w.ui.View().Content, "\n") - 1)
}

func (w *wizard) forward(msg tea.Msg) tea.Cmd {
	s, ok := msg.(*step)
	if ok {
		if s.err != nil && !s.optional {
			return w.finish(failedTitle, failedText)
		}
		w.current++

		return w.advance()
	}
	var cmds [3]tea.Cmd
	w.license, cmds[0] = w.license.Update(msg)
	w.dir, cmds[1] = w.dir.Update(msg)
	w.spin, cmds[2] = w.spin.Update(msg)

	return tea.Batch(cmds[:]...)
}

func (w *wizard) start() tea.Cmd {
	title := installingTitle
	text := installingText
	if w.uninstall {
		title = uninstallingTitle
		text = uninstallingText
	} else {
		w.target.Exe = filepath.Join(w.dir.Value(), exeName)
	}
	w.ui.Pages = append(w.ui.Pages, w.progress(title, text))
	w.ui.Next()

	return tea.Batch(w.spin.Tick, w.advance())
}

func (w *wizard) advance() tea.Cmd {
	for w.current < len(w.steps) && w.steps[w.current].optional && !w.steps[w.current].enabled {
		w.current++
	}
	if w.current < len(w.steps) {
		return w.runStep
	}

	return w.finish(finishedTitle, finishedText)
}

func (w *wizard) finish(title, text string) tea.Cmd {
	w.ui.Pages = append(w.ui.Pages, w.progress(title, text, tui.Buttons(tui.Button{
		Label: "Finish",
		Press: tui.Quit,
	})))
	w.ui.Next()

	return nil
}

func (w *wizard) runStep() tea.Msg {
	s := &w.steps[w.current]
	s.err = s.run(&w.target)

	return s
}
