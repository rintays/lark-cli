package main

import (
	"context"
	"errors"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newChatsUpdateCmd(state *appState) *cobra.Command {
	var chatID string
	var userIDType string
	var name string
	var description string
	var avatar string
	var ownerID string
	var addMemberPermission string
	var shareCardPermission string
	var atAllPermission string
	var editPermission string
	var joinMessageVisibility string
	var leaveMessageVisibility string
	var membershipApproval string
	var chatType string
	var groupMessageType string
	var urgentSetting string
	var videoConferenceSetting string
	var hideMemberCountSetting string
	var pinManageSetting string
	var nameZh string
	var nameEn string
	var nameJa string

	cmd := &cobra.Command{
		Use:   "update <chat-id>",
		Short: "Update chat information",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				return errors.New("chat-id is required")
			}
			chatID = strings.TrimSpace(args[0])
			if chatID == "" {
				return errors.New("chat-id is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			if !hasChatUpdateFields(name, description, avatar, ownerID, addMemberPermission, shareCardPermission, atAllPermission, editPermission, joinMessageVisibility, leaveMessageVisibility, membershipApproval, chatType, groupMessageType, urgentSetting, videoConferenceSetting, hideMemberCountSetting, pinManageSetting, nameZh, nameEn, nameJa) {
				return errors.New("at least one field is required to update")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			var i18nNames *larksdk.I18nNames
			if nameZh != "" || nameEn != "" || nameJa != "" {
				i18nNames = &larksdk.I18nNames{ZhCn: nameZh, EnUs: nameEn, JaJp: nameJa}
			}
			req := larksdk.UpdateChatRequest{
				ChatID:                 chatID,
				UserIDType:             userIDType,
				Name:                   name,
				Description:            description,
				Avatar:                 avatar,
				OwnerID:                ownerID,
				AddMemberPermission:    addMemberPermission,
				ShareCardPermission:    shareCardPermission,
				AtAllPermission:        atAllPermission,
				EditPermission:         editPermission,
				JoinMessageVisibility:  joinMessageVisibility,
				LeaveMessageVisibility: leaveMessageVisibility,
				MembershipApproval:     membershipApproval,
				ChatType:               chatType,
				GroupMessageType:       groupMessageType,
				UrgentSetting:          urgentSetting,
				VideoConferenceSetting: videoConferenceSetting,
				HideMemberCountSetting: hideMemberCountSetting,
				PinManageSetting:       pinManageSetting,
				I18nNames:              i18nNames,
			}
			if err := state.SDK.UpdateChatInfo(context.Background(), token, req); err != nil {
				return err
			}
			payload := map[string]any{"chat_id": chatID, "updated": true}
			return state.Printer.Print(payload, chatID)
		},
	}

	cmd.Flags().StringVar(&userIDType, "user-id-type", "open_id", "user ID type (open_id, union_id, user_id)")
	cmd.Flags().StringVar(&name, "name", "", "chat name")
	cmd.Flags().StringVar(&description, "description", "", "chat description")
	cmd.Flags().StringVar(&avatar, "avatar", "", "avatar image key")
	cmd.Flags().StringVar(&ownerID, "owner-id", "", "new owner ID")
	cmd.Flags().StringVar(&addMemberPermission, "add-member-permission", "", "add member permission (only_owner or all_members)")
	cmd.Flags().StringVar(&shareCardPermission, "share-card-permission", "", "share card permission (allowed or not_allowed)")
	cmd.Flags().StringVar(&atAllPermission, "at-all-permission", "", "at-all permission (only_owner or all_members)")
	cmd.Flags().StringVar(&editPermission, "edit-permission", "", "edit permission (only_owner or all_members)")
	cmd.Flags().StringVar(&joinMessageVisibility, "join-message-visibility", "", "join message visibility (only_owner, all_members, not_anyone)")
	cmd.Flags().StringVar(&leaveMessageVisibility, "leave-message-visibility", "", "leave message visibility (only_owner, all_members, not_anyone)")
	cmd.Flags().StringVar(&membershipApproval, "membership-approval", "", "membership approval (no_approval_required or approval_required)")
	cmd.Flags().StringVar(&chatType, "chat-type", "", "chat type (private or public)")
	cmd.Flags().StringVar(&groupMessageType, "group-message-type", "", "message type (chat or thread)")
	cmd.Flags().StringVar(&urgentSetting, "urgent-setting", "", "urgent setting (only_owner or all_members)")
	cmd.Flags().StringVar(&videoConferenceSetting, "video-conference-setting", "", "video conference setting (only_owner or all_members)")
	cmd.Flags().StringVar(&hideMemberCountSetting, "hide-member-count-setting", "", "hide member count setting (only_owner or all_members)")
	cmd.Flags().StringVar(&pinManageSetting, "pin-manage-setting", "", "pin manage setting (only_owner or all_members)")
	cmd.Flags().StringVar(&nameZh, "name-zh", "", "Chinese name (i18n)")
	cmd.Flags().StringVar(&nameEn, "name-en", "", "English name (i18n)")
	cmd.Flags().StringVar(&nameJa, "name-ja", "", "Japanese name (i18n)")
	return cmd
}

func hasChatUpdateFields(values ...string) bool {
	for _, value := range values {
		if value != "" {
			return true
		}
	}
	return false
}
