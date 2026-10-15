package tui

import (
	"strings"

	"github.com/Alia5/VIIPER/meta"
)

const (
	bigLogoMinWidth   = 140
	smallLogoMinWidth = 110
)

var (
	bigLogo   = strings.ReplaceAll(meta.ASCIIBrailleColoredBig, "\r", "")
	smallLogo = strings.ReplaceAll(meta.ASCIIBrailleColoredSmall, "\r", "")
)

func logoFor(width int) string {
	switch {
	case width >= bigLogoMinWidth:
		return bigLogo
	case width >= smallLogoMinWidth:
		return smallLogo
	default:
		return ""
	}
}
