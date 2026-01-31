package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func newWhoamiCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Show tenant information",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := tokenFor(context.Background(), state, tokenTypesTenant)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			info, err := state.SDK.WhoAmI(context.Background(), token)
			if err != nil {
				return err
			}
			payload := map[string]any{
				"tenant_key": info.TenantKey,
				"name":       info.Name,
			}
			return state.Printer.Print(payload, fmt.Sprintf("%s (%s)", info.Name, info.TenantKey))
		},
	}
	return cmd
}
