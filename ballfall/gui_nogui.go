//go:build nogui

package main

import (
	"fmt"
	"os"
)

func runGUI(_ *Game) {
	fmt.Fprintln(os.Stderr, "this binary was built without GUI support; pass --headless or rebuild with: go build -tags gui")
	os.Exit(1)
}

func runVersusGUI(_ *VersusGame) {
	fmt.Fprintln(os.Stderr, "this binary was built without GUI support; pass --headless or rebuild with: go build -tags gui")
	os.Exit(1)
}
