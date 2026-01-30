package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func newMinutesCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "minutes",
		Short: "Get Minutes details",
	}
	cmd.AddCommand(newMinutesGetCmd(state))
	return cmd
}

func newMinutesGetCmd(state *appState) *cobra.Command {
	var minuteToken string
	var userIDType string

	cmd := &cobra.Command{
		Use:   "get <minute-token>",
		Short: "Get Minutes details",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				if minuteToken != "" && minuteToken != args[0] {
					return errors.New("minute-token provided twice")
				}
				minuteToken = args[0]
			}
			if minuteToken == "" {
				return errors.New("minute-token is required")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			minute, err := state.SDK.GetMinute(context.Background(), token, minuteToken, userIDType)
			if err != nil {
				return err
			}
			payload := map[string]any{"minute": minute}
			text := fmt.Sprintf("%s\t%s\t%s", minute.Token, minute.Title, minute.URL)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&minuteToken, "minute-token", "", "minute token (or provide as positional argument)")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user ID type (user_id, union_id, open_id)")
	return cmd
}
