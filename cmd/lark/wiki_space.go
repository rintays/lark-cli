package main

import (
	"errors"

	"github.com/spf13/cobra"
)

func newWikiSpaceCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "space",
		Short: "Manage Wiki spaces",
	}
	cmd.AddCommand(newWikiSpaceListCmd(state))
	cmd.AddCommand(newWikiSpaceGetCmd(state))
	return cmd
}

func newWikiSpaceListCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Wiki spaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("not implemented")
		},
	}
	return cmd
}

func newWikiSpaceGetCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a Wiki space",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("not implemented")
		},
	}
	return cmd
}
