package main

import (
	"fmt"
	"os"
)

func debugf(state *appState, format string, args ...any) {
	if state == nil || !state.Verbose {
		return
	}
	fmt.Fprintf(os.Stderr, format, args...)
}
