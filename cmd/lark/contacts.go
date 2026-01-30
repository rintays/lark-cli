package main

import (
	"github.com/spf13/cobra"
)

func newContactsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contacts",
		Short: "Manage contacts",
	}
	cmd.AddCommand(newContactsUserCmd(state))
	return cmd
}

func newContactsUserCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "users",
		Aliases: []string{"user"},
		Short:   "Manage contact users",
	}
	cmd.AddCommand(newUserGetCmd(state))
	return cmd
}
