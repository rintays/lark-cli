package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/version"
)

func newVersionCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]string{
				"version": strings.TrimSpace(version.Version),
				"commit":  strings.TrimSpace(version.Commit),
				"date":    strings.TrimSpace(version.Date),
			}
			return state.Printer.Print(payload, fmt.Sprintf("%s", version.String()))
		},
	}
	return cmd
}
