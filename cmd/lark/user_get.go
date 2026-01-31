package main

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newUserInfoCmd(state *appState) *cobra.Command {
	var openID string
	var userID string

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show a contact user by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			request := larksdk.GetContactUserRequest{}
			if openID != "" {
				request.UserID = openID
				request.UserIDType = "open_id"
			} else {
				request.UserID = userID
				request.UserIDType = "user_id"
			}
			user, err := state.SDK.GetContactUser(context.Background(), token, request)
			if err != nil {
				return err
			}
			payload := map[string]any{"user": user}
			return state.Printer.Print(payload, formatUserLine(user))
		},
	}

	cmd.Flags().StringVar(&openID, "open-id", "", "open ID")
	cmd.Flags().StringVar(&userID, "user-id", "", "user ID")
	cmd.MarkFlagsOneRequired("open-id", "user-id")
	cmd.MarkFlagsMutuallyExclusive("open-id", "user-id")

	return cmd
}
