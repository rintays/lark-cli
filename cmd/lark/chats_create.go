package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newChatsCreateCmd(state *appState) *cobra.Command {
	var name string
	var description string
	var avatar string
	var ownerID string
	var userIDs []string
	var botIDs []string
	var groupMessageType string
	var chatMode string
	var chatType string
	var joinMessageVisibility string
	var leaveMessageVisibility string
	var membershipApproval string
	var urgentSetting string
	var videoConferenceSetting string
	var editPermission string
	var hideMemberCountSetting string
	var pinManageSetting string
	var userIDType string
	var setBotManager bool
	var uuid string
	var external bool
	var nameZh string
	var nameEn string
	var nameJa string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a chat",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}

			var i18nNames *larksdk.I18nNames
			if nameZh != "" || nameEn != "" || nameJa != "" {
				i18nNames = &larksdk.I18nNames{ZhCn: nameZh, EnUs: nameEn, JaJp: nameJa}
			}
			var externalPtr *bool
			if external {
				externalPtr = &external
			}

			req := larksdk.CreateChatRequest{
				UserIDType:             userIDType,
				SetBotManager:          boolPtrIfSet(setBotManager),
				UUID:                   uuid,
				Avatar:                 avatar,
				Name:                   name,
				Description:            description,
				I18nNames:              i18nNames,
				OwnerID:                ownerID,
				UserIDList:             userIDs,
				BotIDList:              botIDs,
				GroupMessageType:       groupMessageType,
				ChatMode:               chatMode,
				ChatType:               chatType,
				JoinMessageVisibility:  joinMessageVisibility,
				LeaveMessageVisibility: leaveMessageVisibility,
				MembershipApproval:     membershipApproval,
				UrgentSetting:          urgentSetting,
				VideoConferenceSetting: videoConferenceSetting,
				EditPermission:         editPermission,
				HideMemberCountSetting: hideMemberCountSetting,
				PinManageSetting:       pinManageSetting,
				External:               externalPtr,
			}

			chat, err := state.SDK.CreateChatDetail(context.Background(), token, req)
			if err != nil {
				return err
			}
			payload := map[string]any{"chat": chat}
			text := chat.ChatID
			if chat.Name != "" {
				text = fmt.Sprintf("%s\t%s", chat.ChatID, chat.Name)
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "chat name")
	cmd.Flags().StringVar(&description, "description", "", "chat description")
	cmd.Flags().StringVar(&avatar, "avatar", "", "avatar image key")
	cmd.Flags().StringVar(&ownerID, "owner-id", "", "owner ID")
	cmd.Flags().StringArrayVar(&userIDs, "user-id", nil, "user IDs to invite (repeatable)")
	cmd.Flags().StringArrayVar(&botIDs, "bot-id", nil, "bot app IDs to invite (repeatable)")
	cmd.Flags().StringVar(&groupMessageType, "group-message-type", "", "message type (chat or thread)")
	cmd.Flags().StringVar(&chatMode, "chat-mode", "", "chat mode (group)")
	cmd.Flags().StringVar(&chatType, "chat-type", "", "chat type (private or public)")
	cmd.Flags().StringVar(&joinMessageVisibility, "join-message-visibility", "", "join message visibility (only_owner, all_members, not_anyone)")
	cmd.Flags().StringVar(&leaveMessageVisibility, "leave-message-visibility", "", "leave message visibility (only_owner, all_members, not_anyone)")
	cmd.Flags().StringVar(&membershipApproval, "membership-approval", "", "membership approval (no_approval_required or approval_required)")
	cmd.Flags().StringVar(&urgentSetting, "urgent-setting", "", "urgent setting (only_owner or all_members)")
	cmd.Flags().StringVar(&videoConferenceSetting, "video-conference-setting", "", "video conference setting (only_owner or all_members)")
	cmd.Flags().StringVar(&editPermission, "edit-permission", "", "edit permission (only_owner or all_members)")
	cmd.Flags().StringVar(&hideMemberCountSetting, "hide-member-count-setting", "", "hide member count setting (only_owner or all_members)")
	cmd.Flags().StringVar(&pinManageSetting, "pin-manage-setting", "", "pin manage setting (only_owner or all_members)")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "open_id", "user ID type (open_id, union_id, user_id)")
	cmd.Flags().BoolVar(&setBotManager, "set-bot-manager", false, "set the creating bot as admin (when owner-id is set)")
	cmd.Flags().StringVar(&uuid, "uuid", "", "idempotency UUID")
	cmd.Flags().BoolVar(&external, "external", false, "create as external chat")
	cmd.Flags().StringVar(&nameZh, "name-zh", "", "Chinese name (i18n)")
	cmd.Flags().StringVar(&nameEn, "name-en", "", "English name (i18n)")
	cmd.Flags().StringVar(&nameJa, "name-ja", "", "Japanese name (i18n)")
	return cmd
}

func boolPtrIfSet(value bool) *bool {
	if !value {
		return nil
	}
	return &value
}
