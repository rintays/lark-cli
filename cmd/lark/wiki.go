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
	cmd.AddCommand(newWikiSpaceCmd(state))
	cmd.AddCommand(newWikiMemberCmd(state))
	cmd.AddCommand(newWikiTaskCmd(state))
	return cmd
}

func newWikiNodeCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Manage Wiki nodes",
	}
	cmd.AddCommand(newWikiNodeSearchCmd(state))
	cmd.AddCommand(newWikiNodeInfoCmd(state))
	cmd.AddCommand(newWikiNodeListCmd(state))
	return cmd
}
