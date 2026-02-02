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
			if limit <= 0 {
				limit = 50
			}
			spaceID = strings.TrimSpace(spaceID)
			return runWithToken(cmd, state, nil, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
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
					req := larksdk.ListWikiSpaceMembersRequest{
						SpaceID:   spaceID,
						PageSize:  ps,
						PageToken: pageToken,
					}
					var result larksdk.ListWikiSpaceMembersResult
					var err error
					switch tokenType {
					case tokenTypeTenant:
						result, err = sdk.ListWikiSpaceMembersV2(ctx, token, req)
					case tokenTypeUser:
						result, err = sdk.ListWikiSpaceMembersV2WithUserToken(ctx, token, req)
					default:
						return nil, "", fmt.Errorf("unsupported token type %s", tokenType)
					}
					if err != nil {
						return nil, "", err
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
				return payload, text, nil
			})
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
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			memberType = strings.TrimSpace(args[0])
			memberID = strings.TrimSpace(args[1])
			if memberType == "" {
				return argsUsageError(cmd, errors.New("member-type is required"))
			}
			if memberID == "" {
				return argsUsageError(cmd, errors.New("member-id is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := confirmDestructive(cmd, state, fmt.Sprintf("delete wiki member %s", memberID)); err != nil {
				return err
			}
			spaceID = strings.TrimSpace(spaceID)
			memberType = strings.TrimSpace(memberType)
			memberID = strings.TrimSpace(memberID)
			memberRole = strings.TrimSpace(memberRole)
			needNotificationSet := cmd.Flags().Changed("need-notification")
			req := larksdk.CreateWikiSpaceMemberRequest{
				SpaceID:             spaceID,
				MemberType:          memberType,
				MemberID:            memberID,
				MemberRole:          memberRole,
				NeedNotification:    needNotification,
				NeedNotificationSet: needNotificationSet,
			}
			return runWithToken(cmd, state, nil, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				var member larksdk.WikiSpaceMember
				var err error
				switch tokenType {
				case tokenTypeTenant:
					member, err = sdk.CreateWikiSpaceMemberV2(ctx, token, req)
				case tokenTypeUser:
					member, err = sdk.CreateWikiSpaceMemberV2WithUserToken(ctx, token, req)
				default:
					return nil, "", fmt.Errorf("unsupported token type %s", tokenType)
				}
				if err != nil {
					return nil, "", err
				}

				payload := map[string]any{"member": member}
				text := tableTextRow(
					[]string{"member_type", "member_id", "member_role", "type"},
					[]string{member.MemberType, member.MemberID, member.MemberRole, member.Type},
				)
				return payload, text, nil
			})
		},
	}

	cmd.Flags().StringVar(&spaceID, "space-id", "", "Wiki space ID")
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
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			memberType = strings.TrimSpace(args[0])
			memberID = strings.TrimSpace(args[1])
			if memberType == "" {
				return errors.New("member-type is required")
			}
			if memberID == "" {
				return errors.New("member-id is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceID = strings.TrimSpace(spaceID)
			memberType = strings.TrimSpace(memberType)
			memberID = strings.TrimSpace(memberID)
			req := larksdk.DeleteWikiSpaceMemberRequest{
				SpaceID:    spaceID,
				MemberType: memberType,
				MemberID:   memberID,
			}
			return runWithToken(cmd, state, nil, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				var member larksdk.WikiSpaceMember
				var err error
				switch tokenType {
				case tokenTypeTenant:
					member, err = sdk.DeleteWikiSpaceMemberV2(ctx, token, req)
				case tokenTypeUser:
					member, err = sdk.DeleteWikiSpaceMemberV2WithUserToken(ctx, token, req)
				default:
					return nil, "", fmt.Errorf("unsupported token type %s", tokenType)
				}
				if err != nil {
					return nil, "", err
				}

				payload := map[string]any{"deleted": true, "member": member}
				text := "deleted"
				if member.MemberID != "" {
					text = member.MemberID
				}
				return payload, text, nil
			})
		},
	}

	cmd.Flags().StringVar(&spaceID, "space-id", "", "Wiki space ID")
	_ = cmd.MarkFlagRequired("space-id")
	return cmd
}
