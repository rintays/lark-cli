package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
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
	var spaceID string
	var limit int
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Wiki space members (v2)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			if limit <= 0 {
				limit = 50
			}
			accessToken, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}

			items := make([]larksdk.WikiSpaceMember, 0, limit)
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
				result, err := state.SDK.ListWikiSpaceMembersV2(context.Background(), accessToken, larksdk.ListWikiSpaceMembersRequest{
					SpaceID:   strings.TrimSpace(spaceID),
					PageSize:  ps,
					PageToken: pageToken,
				})
				if err != nil {
					return err
				}
				items = append(items, result.Members...)
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

			payload := map[string]any{"members": items}
			lines := make([]string, 0, len(items))
			for _, m := range items {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s", m.MemberType, m.MemberID, m.MemberRole, m.Type))
			}
			text := "no members found"
			if len(lines) > 0 {
				text = strings.Join(lines, "\n")
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spaceID, "space-id", "", "Wiki space ID")
	cmd.Flags().IntVar(&limit, "limit", 50, "max number of members to return")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "page size (default: auto)")
	_ = cmd.MarkFlagRequired("space-id")
	return cmd
}
