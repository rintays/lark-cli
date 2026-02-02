package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newWikiSpaceUpdateSettingCmd(state *appState) *cobra.Command {
	var spaceID string
	var createSetting string
	var securitySetting string
	var commentSetting string

	cmd := &cobra.Command{
		Use:   "update-setting",
		Short: "Update Wiki space settings (v2)",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.NoArgs(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			changedCreate := cmd.Flags().Changed("create-setting")
			changedSecurity := cmd.Flags().Changed("security-setting")
			changedComment := cmd.Flags().Changed("comment-setting")
			if !changedCreate && !changedSecurity && !changedComment {
				return errors.New("at least one of --create-setting, --security-setting, or --comment-setting is required")
			}
			if changedCreate && strings.TrimSpace(createSetting) == "" {
				return errors.New("create-setting must not be empty")
			}
			if changedSecurity && strings.TrimSpace(securitySetting) == "" {
				return errors.New("security-setting must not be empty")
			}
			if changedComment && strings.TrimSpace(commentSetting) == "" {
				return errors.New("comment-setting must not be empty")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceID = strings.TrimSpace(spaceID)
			req := larksdk.UpdateWikiSpaceSettingRequest{
				SpaceID:            spaceID,
				CreateSetting:      strings.TrimSpace(createSetting),
				SecuritySetting:    strings.TrimSpace(securitySetting),
				CommentSetting:     strings.TrimSpace(commentSetting),
				CreateSettingSet:   cmd.Flags().Changed("create-setting"),
				SecuritySettingSet: cmd.Flags().Changed("security-setting"),
				CommentSettingSet:  cmd.Flags().Changed("comment-setting"),
			}
			return runWithToken(cmd, state, nil, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				var setting larksdk.WikiSpaceSetting
				var err error
				switch tokenType {
				case tokenTypeTenant:
					setting, err = sdk.UpdateWikiSpaceSettingV2(ctx, token, req)
				case tokenTypeUser:
					setting, err = sdk.UpdateWikiSpaceSettingV2WithUserToken(ctx, token, req)
				default:
					return nil, "", fmt.Errorf("unsupported token type %s", tokenType)
				}
				if err != nil {
					return nil, "", err
				}
				payload := map[string]any{"setting": setting}
				text := tableTextRow(
					[]string{"create_setting", "security_setting", "comment_setting"},
					[]string{setting.CreateSetting, setting.SecuritySetting, setting.CommentSetting},
				)
				return payload, text, nil
			})
		},
	}

	cmd.Flags().StringVar(&spaceID, "space-id", "", "Wiki space ID")
	cmd.Flags().StringVar(&createSetting, "create-setting", "", "who can create top-level pages (admin or admin_and_member)")
	cmd.Flags().StringVar(&securitySetting, "security-setting", "", "readers can copy/print/export (allow or not_allow)")
	cmd.Flags().StringVar(&commentSetting, "comment-setting", "", "readers can comment (allow or not_allow)")
	_ = cmd.MarkFlagRequired("space-id")
	return cmd
}
