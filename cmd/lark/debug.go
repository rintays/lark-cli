package main

import "fmt"

func debugf(state *appState, format string, args ...any) {
	if state == nil || !state.Verbose {
		return
	}
	fmt.Fprintf(errWriter(state), format, args...)
}
