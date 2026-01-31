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
	cmd.AddCommand(newWikiMemberAddCmd(state))
	cmd.AddCommand(newWikiMemberDeleteCmd(state))
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
			text := tableText([]string{"member_type", "member_id", "member_role", "type"}, lines, "no members found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spaceID, "space-id", "", "Wiki space ID")
	cmd.Flags().IntVar(&limit, "limit", 50, "max number of members to return")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "page size (default: auto)")
	_ = cmd.MarkFlagRequired("space-id")
	return cmd
}

func newWikiMemberAddCmd(state *appState) *cobra.Command {
	var spaceID string
	var memberType string
	var memberID string
	var memberRole string
	var needNotification bool

	cmd := &cobra.Command{
		Use:   "add <member-type> <member-id>",
		Short: "Add a Wiki space member (v2)",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(2)(cmd, args); err != nil {
				return err
			}
			if len(args) > 0 {
				if memberType != "" && memberType != args[0] {
					return errors.New("member-type provided twice")
				}
				if err := cmd.Flags().Set("member-type", args[0]); err != nil {
					return err
				}
			}
			if len(args) > 1 {
				if memberID != "" && memberID != args[1] {
					return errors.New("member-id provided twice")
				}
				if err := cmd.Flags().Set("member-id", args[1]); err != nil {
					return err
				}
			}
			if strings.TrimSpace(spaceID) == "" {
				return errors.New("space-id is required")
			}
			if strings.TrimSpace(memberType) == "" {
				return errors.New("member-type is required")
			}
			if strings.TrimSpace(memberID) == "" {
				return errors.New("member-id is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			accessToken, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}

			needNotificationSet := cmd.Flags().Changed("need-notification")
			member, err := state.SDK.CreateWikiSpaceMemberV2(context.Background(), accessToken, larksdk.CreateWikiSpaceMemberRequest{
				SpaceID:             strings.TrimSpace(spaceID),
				MemberType:          strings.TrimSpace(memberType),
				MemberID:            strings.TrimSpace(memberID),
				MemberRole:          strings.TrimSpace(memberRole),
				NeedNotification:    needNotification,
				NeedNotificationSet: needNotificationSet,
			})
			if err != nil {
				return err
			}

			payload := map[string]any{"member": member}
			text := tableTextRow(
				[]string{"member_type", "member_id", "member_role", "type"},
				[]string{member.MemberType, member.MemberID, member.MemberRole, member.Type},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spaceID, "space-id", "", "Wiki space ID")
	cmd.Flags().StringVar(&memberType, "member-type", "", "member type (userid, email, openid, unionid, openchat, opendepartmentid) (or provide as positional argument)")
	cmd.Flags().StringVar(&memberID, "member-id", "", "member id (or provide as positional argument)")
	cmd.Flags().StringVar(&memberRole, "role", "member", "member role (member, admin)")
	cmd.Flags().BoolVar(&needNotification, "need-notification", false, "notify the member after adding permissions")
	_ = cmd.MarkFlagRequired("space-id")
	return cmd
}

func newWikiMemberDeleteCmd(state *appState) *cobra.Command {
	var spaceID string
	var memberType string
	var memberID string

	cmd := &cobra.Command{
		Use:   "delete <member-type> <member-id>",
		Short: "Delete a Wiki space member (v2)",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(2)(cmd, args); err != nil {
				return err
			}
			if len(args) > 0 {
				if memberType != "" && memberType != args[0] {
					return errors.New("member-type provided twice")
				}
				if err := cmd.Flags().Set("member-type", args[0]); err != nil {
					return err
				}
			}
			if len(args) > 1 {
				if memberID != "" && memberID != args[1] {
					return errors.New("member-id provided twice")
				}
				if err := cmd.Flags().Set("member-id", args[1]); err != nil {
					return err
				}
			}
			if strings.TrimSpace(spaceID) == "" {
				return errors.New("space-id is required")
			}
			if strings.TrimSpace(memberType) == "" {
				return errors.New("member-type is required")
			}
			if strings.TrimSpace(memberID) == "" {
				return errors.New("member-id is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			accessToken, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}

			member, err := state.SDK.DeleteWikiSpaceMemberV2(context.Background(), accessToken, larksdk.DeleteWikiSpaceMemberRequest{
				SpaceID:    strings.TrimSpace(spaceID),
				MemberType: strings.TrimSpace(memberType),
				MemberID:   strings.TrimSpace(memberID),
			})
			if err != nil {
				return err
			}

			payload := map[string]any{"deleted": true, "member": member}
			text := "deleted"
			if member.MemberID != "" {
				text = member.MemberID
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spaceID, "space-id", "", "Wiki space ID")
	cmd.Flags().StringVar(&memberType, "member-type", "", "member type (userid, email, openid, unionid, openchat, opendepartmentid) (or provide as positional argument)")
	cmd.Flags().StringVar(&memberID, "member-id", "", "member id (or provide as positional argument)")
	_ = cmd.MarkFlagRequired("space-id")
	return cmd
}
