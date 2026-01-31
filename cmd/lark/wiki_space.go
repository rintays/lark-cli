package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newWikiSpaceCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "space",
		Short: "Manage Wiki spaces",
	}
	cmd.AddCommand(newWikiSpaceCreateCmd(state))
	cmd.AddCommand(newWikiSpaceListCmd(state))
	cmd.AddCommand(newWikiSpaceInfoCmd(state))
	return cmd
}

func newWikiSpaceListCmd(state *appState) *cobra.Command {
	var limit int
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Wiki spaces (v2)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			if limit <= 0 {
				limit = 50
			}
			tenantToken, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}

			items := make([]larksdk.WikiSpace, 0, limit)
			pageToken := ""
			remaining := limit
			for {
				ps := pageSize
				if ps <= 0 {
					ps = remaining
					if ps > 200 {
						ps = 200
					}
				}
				result, err := state.SDK.ListWikiSpacesV2(context.Background(), tenantToken, larksdk.ListWikiSpacesRequest{
					PageSize:  ps,
					PageToken: pageToken,
				})
				if err != nil {
					return err
				}
				items = append(items, result.Items...)
				if len(items) >= limit || !result.HasMore {
					break
				}
				remaining = limit - len(items)
				pageToken = result.PageToken
				if pageToken == "" {
					break
				}
			}
			if len(items) > limit {
				items = items[:limit]
			}

			payload := map[string]any{"spaces": items}
			lines := make([]string, 0, len(items))
			for _, s := range items {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s", s.SpaceID, s.Name, s.SpaceType, s.Visibility))
			}
			text := tableText([]string{"space_id", "name", "space_type", "visibility"}, lines, "no spaces found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 50, "max number of spaces to return")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "page size (default: auto)")
	return cmd
}

func newWikiSpaceInfoCmd(state *appState) *cobra.Command {
	var spaceID string

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show a Wiki space (v2)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			tenantToken, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			space, err := state.SDK.GetWikiSpaceV2(context.Background(), tenantToken, larksdk.GetWikiSpaceRequest{SpaceID: strings.TrimSpace(spaceID)})
			if err != nil {
				return err
			}
			payload := map[string]any{"space": space}
			text := tableTextRow(
				[]string{"space_id", "name", "space_type", "visibility"},
				[]string{space.SpaceID, space.Name, space.SpaceType, space.Visibility},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spaceID, "space-id", "", "Wiki space ID")
	_ = cmd.MarkFlagRequired("space-id")
	return cmd
}
