package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

const maxChatMembersPageSize = 50

func newChatsGetCmd(state *appState) *cobra.Command {
	var chatID string
	var userIDType string
	var membersLimit int
	var membersPageSize int

	cmd := &cobra.Command{
		Use:   "get <chat-id>",
		Short: "Get chat information",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) > 0 {
				if chatID != "" && chatID != args[0] {
					return errors.New("chat-id provided twice")
				}
				if err := cmd.Flags().Set("chat-id", args[0]); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			chat, err := state.SDK.GetChatInfo(context.Background(), token, larksdk.GetChatRequest{
				ChatID:     chatID,
				UserIDType: userIDType,
			})
			if err != nil {
				return err
			}
			members, membersTotal, membersTruncated, membersErr := listChatMembers(context.Background(), state, token, chatID, userIDType, membersLimit, membersPageSize)
			payload := map[string]any{"chat": chat}
			if membersLimit > 0 {
				if membersErr != nil {
					payload["members_error"] = membersErr.Error()
				} else {
					payload["members"] = members
					payload["members_total"] = membersTotal
					if membersTruncated {
						payload["members_truncated"] = true
					}
				}
			}
			text := formatChatInfoText(chat, members, membersTotal, membersTruncated, membersErr, membersLimit)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&chatID, "chat-id", "", "chat ID (or provide as positional argument)")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "open_id", "user ID type (open_id, union_id, user_id)")
	cmd.Flags().IntVar(&membersLimit, "members-limit", 20, "max number of chat members to return (0 to skip)")
	cmd.Flags().IntVar(&membersPageSize, "members-page-size", 0, "chat members page size (default: auto)")
	_ = cmd.MarkFlagRequired("chat-id")
	return cmd
}

func listChatMembers(ctx context.Context, state *appState, token string, chatID string, memberIDType string, limit int, pageSize int) ([]larksdk.ChatMember, int, bool, error) {
	if limit <= 0 {
		return nil, 0, false, nil
	}
	members := make([]larksdk.ChatMember, 0, limit)
	pageToken := ""
	remaining := limit
	var membersTotal int
	truncated := false

	for {
		ps := pageSize
		if ps <= 0 {
			ps = remaining
			if ps > maxChatMembersPageSize {
				ps = maxChatMembersPageSize
			}
		}
		result, err := state.SDK.ListChatMembers(ctx, token, larksdk.ListChatMembersRequest{
			ChatID:       chatID,
			MemberIDType: memberIDType,
			PageSize:     ps,
			PageToken:    pageToken,
		})
		if err != nil {
			return members, membersTotal, truncated, err
		}
		if membersTotal == 0 && result.MemberTotal != 0 {
			membersTotal = result.MemberTotal
		}
		members = append(members, result.Items...)
		if len(members) >= limit {
			if len(members) > limit {
				members = members[:limit]
			}
			if result.HasMore || result.PageToken != "" {
				truncated = true
			}
			break
		}
		if !result.HasMore {
			break
		}
		remaining = limit - len(members)
		pageToken = result.PageToken
		if pageToken == "" {
			break
		}
	}
	return members, membersTotal, truncated, nil
}

func formatChatInfoText(chat larksdk.ChatInfo, members []larksdk.ChatMember, membersTotal int, membersTruncated bool, membersErr error, membersLimit int) string {
	lines := make([]string, 0, 32)
	lines = append(lines, fmt.Sprintf("chat_id: %s", chat.ChatID))
	appendIf := func(label string, value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		lines = append(lines, fmt.Sprintf("%s: %s", label, value))
	}
	appendIf("name", chat.Name)
	appendIf("description", chat.Description)
	if chat.I18nNames != nil {
		appendIf("i18n_names", formatI18nNames(chat.I18nNames))
	}
	if chat.OwnerID != "" {
		owner := chat.OwnerID
		if chat.OwnerIDType != "" {
			owner = fmt.Sprintf("%s (%s)", owner, chat.OwnerIDType)
		}
		lines = append(lines, fmt.Sprintf("owner: %s", owner))
	}
	if len(chat.UserManagerIDList) > 0 {
		lines = append(lines, fmt.Sprintf("user_managers: %s", strings.Join(chat.UserManagerIDList, ", ")))
	}
	if len(chat.BotManagerIDList) > 0 {
		lines = append(lines, fmt.Sprintf("bot_managers: %s", strings.Join(chat.BotManagerIDList, ", ")))
	}
	appendIf("chat_type", chat.ChatType)
	appendIf("chat_mode", chat.ChatMode)
	appendIf("chat_tag", chat.ChatTag)
	appendIf("chat_status", chat.ChatStatus)
	appendIf("group_message_type", chat.GroupMessageType)
	appendIf("tenant_key", chat.TenantKey)
	appendIf("user_count", chat.UserCount)
	appendIf("bot_count", chat.BotCount)
	appendIf("join_message_visibility", chat.JoinMessageVisibility)
	appendIf("leave_message_visibility", chat.LeaveMessageVisibility)
	appendIf("membership_approval", chat.MembershipApproval)
	appendIf("moderation_permission", chat.ModerationPermission)
	appendIf("add_member_permission", chat.AddMemberPermission)
	appendIf("share_card_permission", chat.ShareCardPermission)
	appendIf("at_all_permission", chat.AtAllPermission)
	appendIf("edit_permission", chat.EditPermission)
	appendIf("urgent_setting", chat.UrgentSetting)
	appendIf("video_conference_setting", chat.VideoConferenceSetting)
	appendIf("pin_manage_setting", chat.PinManageSetting)
	appendIf("hide_member_count_setting", chat.HideMemberCountSetting)
	if chat.RestrictedModeSetting != nil {
		appendIf("restricted_mode", formatRestrictedModeSetting(chat.RestrictedModeSetting))
	}
	lines = append(lines, fmt.Sprintf("external: %t", chat.External))

	if membersLimit > 0 {
		if membersErr != nil {
			lines = append(lines, fmt.Sprintf("members: (failed to load: %s)", membersErr.Error()))
		} else {
			membersLine := fmt.Sprintf("members: %d", len(members))
			if membersTotal > 0 {
				membersLine = fmt.Sprintf("members: %d/%d", len(members), membersTotal)
			}
			if membersTruncated {
				membersLine = fmt.Sprintf("%s (truncated)", membersLine)
			}
			lines = append(lines, membersLine)
			for _, member := range members {
				lines = append(lines, fmt.Sprintf("  - %s", formatChatMember(member)))
			}
		}
	}
	return strings.Join(lines, "\n")
}

func formatChatMember(member larksdk.ChatMember) string {
	display := member.MemberID
	if member.Name != "" {
		display = fmt.Sprintf("%s (%s)", member.Name, member.MemberID)
	}
	if member.MemberIDType != "" {
		display = fmt.Sprintf("%s [%s]", display, member.MemberIDType)
	}
	return display
}

func formatI18nNames(names *larksdk.I18nNames) string {
	if names == nil {
		return ""
	}
	parts := make([]string, 0, 3)
	if names.ZhCn != "" {
		parts = append(parts, fmt.Sprintf("zh_cn=%s", names.ZhCn))
	}
	if names.EnUs != "" {
		parts = append(parts, fmt.Sprintf("en_us=%s", names.EnUs))
	}
	if names.JaJp != "" {
		parts = append(parts, fmt.Sprintf("ja_jp=%s", names.JaJp))
	}
	return strings.Join(parts, ", ")
}

func formatRestrictedModeSetting(setting *larksdk.RestrictedModeSetting) string {
	if setting == nil {
		return ""
	}
	parts := []string{fmt.Sprintf("status=%t", setting.Status)}
	if setting.ScreenshotHasPermissionSetting != "" {
		parts = append(parts, fmt.Sprintf("screenshot=%s", setting.ScreenshotHasPermissionSetting))
	}
	if setting.DownloadHasPermissionSetting != "" {
		parts = append(parts, fmt.Sprintf("download=%s", setting.DownloadHasPermissionSetting))
	}
	if setting.MessageHasPermissionSetting != "" {
		parts = append(parts, fmt.Sprintf("message=%s", setting.MessageHasPermissionSetting))
	}
	return strings.Join(parts, ", ")
}
