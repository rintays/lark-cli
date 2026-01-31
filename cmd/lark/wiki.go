package main

import (
	"github.com/spf13/cobra"
)

func newWikiCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wiki",
		Short: "Manage Wiki",
	}
	cmd.AddCommand(newWikiNodeCmd(state))
	return cmd
}

func newWikiNodeCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Manage Wiki nodes",
	}
	cmd.AddCommand(newWikiNodeSearchCmd(state))
	return cmd
}
