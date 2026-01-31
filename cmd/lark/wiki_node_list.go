package main

import (
	"errors"

	"github.com/spf13/cobra"
)

func newWikiNodeListCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Wiki nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("not implemented")
		},
	}
	return cmd
}
