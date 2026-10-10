package main

import (
	"os"

	"golang.org/x/term"
)

func init() {
	if len(os.Args) >= 2 || !term.IsTerminal(int(os.Stdout.Fd())) {
		return
	}
	launch()
}
