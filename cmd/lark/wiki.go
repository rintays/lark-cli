package main

import (
	"github.com/spf13/cobra"
)

func newWikiCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wiki",
		Short: "Manage Wiki",
	Long: `Wiki (Knowledge Base) organizes content into spaces and a tree of nodes.

- Space: a knowledge space (space_id) with visibility (public/private) and type (team/personal).
- Node: a tree entry in a space identified by node_token; it links to a Drive object via obj_type + obj_token (doc/sheet/mindnote) and can be an origin or shortcut node.
- Members define access roles for a space.
- Use lark drive permissions to manage collaborators for the underlying Drive object (use obj_token as FILE_TOKEN).
- Tasks track async operations (e.g. moving Drive docs into a wiki space).

Relationships: space -> nodes (tree) -> Drive objects; members + tasks belong to a space.`,
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
