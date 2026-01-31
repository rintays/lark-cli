package main

import (
	"errors"

	"github.com/spf13/cobra"
)

func newWikiMemberCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member",
		Short: "Manage Wiki members",
	}
	cmd.AddCommand(newWikiMemberListCmd(state))
	return cmd
}

func newWikiMemberListCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Wiki members",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("not implemented")
		},
	}
	return cmd
}
