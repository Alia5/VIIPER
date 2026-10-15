package install

import "github.com/Alia5/VIIPER/tui"

const (
	installingTitle   = "Installing"
	installingText    = "Please wait while Setup installs VIIPER on your computer."
	uninstallingTitle = "Uninstalling"
	uninstallingText  = "Please wait while VIIPER is removed from your computer."
	finishedTitle     = "Completing the VIIPER Setup Wizard"
	finishedText      = "Setup has finished. Select Finish to exit Setup."
	failedTitle       = "Setup Failed"
	failedText        = "Setup could not complete. See the step marked ❌ below."
)

func (w *wizard) progress(title, text string, buttons ...tui.Button) tui.Page {
	var steps []tui.Step
	for i, s := range w.steps {
		state := tui.StepPending
		switch {
		case !s.shouldRun(w.flags):
			state = tui.StepSkipped
		case s.err != nil:
			state = tui.StepFailed
		case i < w.current:
			state = tui.StepDone
		case i == w.current:
			state = tui.StepRunning
		}
		steps = append(steps, tui.Step{
			Label: s.label,
			State: state,
			Err:   s.err,
		})
	}

	return tui.Page{
		Title:   title,
		Text:    text,
		Hint:    "Enter: confirm   Esc: quit",
		Steps:   steps,
		Buttons: buttons,
	}
}
