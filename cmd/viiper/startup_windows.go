package main

import (
	"os"
	"strings"

	"github.com/Alia5/VIIPER/internal/util"
)

func init() {
	if !util.IsRunFromGUI() {
		return
	}
	if len(os.Args) >= 2 && !strings.HasPrefix(os.Args[1], "-") {
		return
	}
	launch()
}
