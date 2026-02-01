package main

import (
	"errors"
	"io"
	"os"
	"strings"
)

func readInputFile(path string) ([]byte, error) {
	normalized := strings.TrimSpace(path)
	if normalized == "" {
		return nil, errors.New("file path is required")
	}
	if normalized == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(normalized)
}
