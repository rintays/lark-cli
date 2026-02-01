package main

import (
	"io"
	"os"
)

func errWriter(state *appState) io.Writer {
	if state != nil && state.ErrWriter != nil {
		return state.ErrWriter
	}
	return os.Stderr
}
