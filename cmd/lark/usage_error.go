package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/output"
)

func usageError(cmd *cobra.Command, message string, hint string) error {
	return usageErrorWithUsage(cmd, message, hint, "")
}

func usageErrorWithUsage(cmd *cobra.Command, message string, hint string, usage string) error {
	useText := strings.TrimSpace(usage)
	if useText == "" && cmd != nil {
		useLine := strings.TrimSpace(cmd.UseLine())
		if useLine != "" && !strings.HasPrefix(useLine, "lark ") {
			useLine = fmt.Sprintf("lark %s", useLine)
		}
		useText = useLine
	}
	return output.UsageError{
		Message: strings.TrimSpace(message),
		Usage:   useText,
		Hint:    strings.TrimSpace(hint),
	}
}

func usagef(cmd *cobra.Command, format string, args ...any) error {
	return usageErrorWithUsage(cmd, fmt.Sprintf(format, args...), "", cmd.UsageString())
}

func flagUsage(cmd *cobra.Command, message string) error {
	return usageErrorWithUsage(cmd, message, "", cmd.UsageString())
}

func argsUsageError(cmd *cobra.Command, err error) error {
	if err == nil {
		return nil
	}
	usage := ""
	if cmd != nil {
		usage = cmd.UsageString()
	}
	hint := ""
	if cmd != nil {
		example := strings.TrimSpace(cmd.Example)
		if example != "" {
			hint = fmt.Sprintf("Example:\n%s", example)
		} else {
			command := strings.TrimSpace(cmd.CommandPath())
			if command != "" {
				hint = fmt.Sprintf("Run: %s --help for details.", command)
			}
		}
	}
	return usageErrorWithUsage(cmd, err.Error(), hint, usage)
}

func flagErrorHint(cmd *cobra.Command, err error) string {
	if cmd == nil || err == nil {
		return ""
	}
	command := strings.TrimSpace(cmd.CommandPath())
	if command == "" {
		return ""
	}
	message := err.Error()
	switch {
	case strings.Contains(message, "unknown flag"):
		return fmt.Sprintf("Run: %s --help to see available flags.", command)
	case strings.Contains(message, "required flag(s)"):
		return fmt.Sprintf("Run: %s --help to see required flags and examples.", command)
	default:
		return fmt.Sprintf("Run: %s --help for details.", command)
	}
}
