package tui

import "fmt"

type StepState int

const (
	StepPending StepState = iota
	StepRunning
	StepDone
	StepSkipped
	StepFailed
)

const (
	pendingMark = "⏳ "
	doneMark    = "✅ "
	skippedMark = "- "
	failedMark  = "❌ "
	spinnerGap  = " "
)

type Step struct {
	Label string
	State StepState
	Err   error
}

func (s *screen) stepView(step Step) string {
	switch step.State {
	case StepRunning:
		return focusBlank + s.spin.View() + spinnerGap + step.Label
	case StepDone:
		return focusBlank + doneMark + step.Label
	case StepSkipped:
		return focusBlank + skippedMark + step.Label
	case StepFailed:
		return focusBlank + failedMark + fmt.Sprintf("%s: %v", step.Label, step.Err)
	default:
		return focusBlank + pendingMark + step.Label
	}
}

func (s *screen) shouldSpin() bool {
	for _, step := range s.page.Steps {
		if step.State == StepRunning {
			return true
		}
	}

	return false
}
