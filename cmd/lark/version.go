package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"lark/internal/version"
)

func newVersionCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		RunE: func(cmd *cobra.Command, args []string) error {
			value := version.Version
			return state.Printer.Print(map[string]string{"version": value}, fmt.Sprintf("%s", value))
		},
	}
	return cmd
}
