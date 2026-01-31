package main

import (
	"errors"

	"github.com/spf13/cobra"
)

func newWikiTaskCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Query Wiki tasks",
	}
	cmd.AddCommand(newWikiTaskListCmd(state))
	return cmd
}

func newWikiTaskListCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Wiki tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("not implemented")
		},
	}
	return cmd
}
