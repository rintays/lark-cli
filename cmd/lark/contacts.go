package main

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"lark/internal/larkapi"
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
		Short: "Manage contact users",
	}
	cmd.AddCommand(newContactsUserGetCmd(state))
	return cmd
}

func newContactsUserGetCmd(state *appState) *cobra.Command {
	var openID string
	var userID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a contact user by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			if openID == "" && userID == "" {
				return errors.New("open-id or user-id is required")
			}
			if openID != "" && userID != "" {
				return errors.New("only one of open-id or user-id can be used at a time")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			request := larkapi.GetContactUserRequest{}
			if openID != "" {
				request.UserID = openID
				request.UserIDType = "open_id"
			} else {
				request.UserID = userID
				request.UserIDType = "user_id"
			}
			user, err := state.Client.GetContactUser(context.Background(), token, request)
			if err != nil {
				return err
			}
			payload := map[string]any{"user": user}
			return state.Printer.Print(payload, formatUserLine(user))
		},
	}

	cmd.Flags().StringVar(&openID, "open-id", "", "open ID")
	cmd.Flags().StringVar(&userID, "user-id", "", "user ID")

	return cmd
}
