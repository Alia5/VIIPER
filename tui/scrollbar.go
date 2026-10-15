package tui

import (
	"math"
	"strings"
)

const (
	scrollbarGap   = " "
	scrollbarThumb = "█"
	scrollbarTrack = "░"
	minThumbHeight = 1
)

func (s *screen) scrollbar() string {
	height := s.text.Height()
	thumbHeight := max(height*height/s.text.TotalLineCount(), minThumbHeight)
	above := int(math.Round(s.text.ScrollPercent() * float64(height-thumbHeight)))
	track := trackStyle.Render(scrollbarTrack) + "\n"
	thumb := thumbStyle.Render(scrollbarThumb) + "\n"

	return strings.TrimSuffix(
		strings.Repeat(track, above)+strings.Repeat(thumb, thumbHeight)+strings.Repeat(track, height-above-thumbHeight),
		"\n",
	)
}
