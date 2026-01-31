package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

const maxMinutesPageSize = 50

func newMinutesCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "minutes",
		Short: "Manage Minutes",
	}
	cmd.AddCommand(newMinutesInfoCmd(state))
	cmd.AddCommand(newMinutesListCmd(state))
	return cmd
}

func newMinutesInfoCmd(state *appState) *cobra.Command {
	var minuteToken string
	var userIDType string

	cmd := &cobra.Command{
		Use:   "info <minute-token>",
		Short: "Show Minutes details",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) > 0 {
				if minuteToken != "" && minuteToken != args[0] {
					return errors.New("minute-token provided twice")
				}
				if err := cmd.Flags().Set("minute-token", args[0]); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
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
	_ = cmd.MarkFlagRequired("minute-token")
	return cmd
}

func newMinutesListCmd(state *appState) *cobra.Command {
	var limit int
	var userIDType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Minutes",
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				return errors.New("limit must be greater than 0")
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			minutes := make([]larksdk.Minute, 0, limit)
			pageToken := ""
			remaining := limit
			for {
				pageSize := remaining
				if pageSize > maxMinutesPageSize {
					pageSize = maxMinutesPageSize
				}
				result, err := state.SDK.ListMinutes(context.Background(), token, larksdk.ListMinutesRequest{
					PageSize:   pageSize,
					PageToken:  pageToken,
					UserIDType: userIDType,
				})
				if err != nil {
					return err
				}
				minutes = append(minutes, result.Items...)
				if len(minutes) >= limit || !result.HasMore {
					break
				}
				remaining = limit - len(minutes)
				pageToken = result.PageToken
				if pageToken == "" {
					break
				}
			}
			if len(minutes) > limit {
				minutes = minutes[:limit]
			}
			payload := map[string]any{"minutes": minutes}
			lines := make([]string, 0, len(minutes))
			for _, minute := range minutes {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s", minute.Token, minute.Title, minute.URL))
			}
			text := "no minutes found"
			if len(lines) > 0 {
				text = strings.Join(lines, "\n")
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "max number of minutes to return")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user ID type (user_id, union_id, open_id)")
	return cmd
}
