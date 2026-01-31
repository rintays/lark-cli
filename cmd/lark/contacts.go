package main

import (
	"github.com/spf13/cobra"
)

func newContactsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contacts",
		Short: "Manage contacts",
		Long: `Contacts expose the organization directory.

- Contact users are address-book entries for tenant users.
- Use contacts user info to resolve basic profile data for directory lookups.`,
	}
	cmd.AddCommand(newContactsUserCmd(state))
	return cmd
}

func newContactsUserCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage contact user",
	}
	cmd.AddCommand(newUserInfoCmd(state))
	return cmd
}
