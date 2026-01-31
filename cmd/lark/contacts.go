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
		Use:   "user",
		Short: "Manage contact user",
	}
	cmd.AddCommand(newUserInfoCmd(state))
	return cmd
}
