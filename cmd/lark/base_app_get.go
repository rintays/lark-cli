package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func newBaseAppGetCmd(state *appState) *cobra.Command {
	var appToken string

	cmd := &cobra.Command{
		Use:     "info",
		Aliases: []string{"get"},
		Short:   "Get a Bitable app",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenant)
			if err != nil {
				return err
			}
			app, err := state.SDK.GetBitableApp(context.Background(), token, appToken)
			if err != nil {
				return err
			}
			payload := map[string]any{"app": app}
			text := fmt.Sprintf("%s\t%s", app.AppToken, app.Name)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
