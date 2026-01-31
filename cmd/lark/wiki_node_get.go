package main

import (
	"errors"

	"github.com/spf13/cobra"
)

func newWikiNodeGetCmd(state *appState) *cobra.Command {
	var nodeToken string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a Wiki node",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = nodeToken
			return errors.New("not implemented")
		},
	}

	cmd.Flags().StringVar(&nodeToken, "node-token", "", "wiki node token")
	_ = cmd.MarkFlagRequired("node-token")
	return cmd
}
