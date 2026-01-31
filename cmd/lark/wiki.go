package main

import (
	"github.com/spf13/cobra"
)

func newWikiCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wiki",
		Short: "Manage Wiki",
		Long: `Wiki organizes content into spaces and nodes.

- A space is the top-level container.
- A node represents an entry in a space and points to a Drive file (doc/sheet/etc).
- Members define access roles; tasks track background operations in a space.`,
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
