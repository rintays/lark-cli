package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func confirmDestructive(cmd *cobra.Command, state *appState, action string) error {
	if state != nil && state.Force {
		return nil
	}
	if state != nil && state.NoInput {
		return usageErrorWithUsage(cmd, "confirmation required; use --force to proceed", "", cmd.UsageString())
	}
	if !isInteractiveInput(cmd) {
		return usageErrorWithUsage(cmd, "confirmation required; use --force to proceed", "", cmd.UsageString())
	}
	prompt := "Proceed"
	if strings.TrimSpace(action) != "" {
		prompt = fmt.Sprintf("Proceed to %s", strings.TrimSpace(action))
	}
	fmt.Fprintf(errWriter(state), "%s? [y/N]: ", prompt)
	reader := bufio.NewReader(cmd.InOrStdin())
	line, _ := reader.ReadString('\n')
	answer := strings.TrimSpace(line)
	if strings.EqualFold(answer, "y") || strings.EqualFold(answer, "yes") {
		return nil
	}
	return errors.New("canceled")
}

func isInteractiveInput(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	in := cmd.InOrStdin()
	if in == nil {
		return false
	}
	return isTerminal(in)
}

func isTerminal(r io.Reader) bool {
	type fdReader interface {
		Fd() uintptr
	}
	fr, ok := r.(fdReader)
	if !ok {
		return false
	}
	fd := fr.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}
