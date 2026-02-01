package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/output"
)

func usageError(cmd *cobra.Command, message string, hint string) error {
	useLine := ""
	if cmd != nil {
		useLine = strings.TrimSpace(cmd.UseLine())
	}
	if useLine != "" && !strings.HasPrefix(useLine, "lark ") {
		useLine = fmt.Sprintf("lark %s", useLine)
	}
	return output.UsageError{
		Message: strings.TrimSpace(message),
		Usage:   useLine,
		Hint:    strings.TrimSpace(hint),
	}
}
