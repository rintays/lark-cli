package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newMinutesCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "minutes",
		Short: "Manage Minutes",
		Long: `Minutes are meeting recordings/transcripts stored as Drive files.

- minute_token identifies a Minutes file; URL opens it.
- list reads Minutes entries from Drive folders.
- update manages sharing permissions; delete removes the file.`,
	}
	cmd.AddCommand(newMinutesInfoCmd(state))
	cmd.AddCommand(newMinutesListCmd(state))
	cmd.AddCommand(newMinutesDeleteCmd(state))
	cmd.AddCommand(newMinutesUpdateCmd(state))
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
				return nil
			}
			if strings.TrimSpace(minuteToken) == "" {
				return errors.New("minute-token is required")
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
			text := tableTextRow(
				[]string{"minute_token", "title", "url"},
				[]string{minute.Token, minute.Title, minute.URL},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&minuteToken, "minute-token", "", "minute token (or provide as positional argument)")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user ID type (user_id, union_id, open_id)")
	return cmd
}

func newMinutesListCmd(state *appState) *cobra.Command {
	var limit int
	var folderID string
	var fileType string
	var query string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Minutes",
		Args:  cobra.NoArgs,
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
			fileType = strings.TrimSpace(fileType)
			query = strings.TrimSpace(query)
			queryLower := strings.ToLower(query)
			minutes := make([]larksdk.Minute, 0, limit)
			pageToken := ""
			for {
				pageSize := maxDrivePageSize
				remaining := limit - len(minutes)
				if remaining <= 0 {
					break
				}
				if pageSize > remaining {
					pageSize = remaining
				}
				result, err := state.SDK.ListDriveFiles(context.Background(), token, larksdk.ListDriveFilesRequest{
					FolderToken: folderID,
					PageSize:    pageSize,
					PageToken:   pageToken,
				})
				if err != nil {
					return err
				}
				for _, file := range result.Files {
					if fileType != "" && !strings.EqualFold(file.FileType, fileType) {
						continue
					}
					if queryLower != "" && !strings.Contains(strings.ToLower(file.Name), queryLower) {
						continue
					}
					minutes = append(minutes, larksdk.Minute{
						Token: file.Token,
						Title: file.Name,
						URL:   file.URL,
					})
					if len(minutes) >= limit {
						break
					}
				}
				if len(minutes) >= limit || !result.HasMore {
					break
				}
				pageToken = result.PageToken
				if pageToken == "" {
					break
				}
			}
			payload := map[string]any{"minutes": minutes}
			lines := make([]string, 0, len(minutes))
			for _, minute := range minutes {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s", minute.Token, minute.Title, minute.URL))
			}
			text := tableText([]string{"minute_token", "title", "url"}, lines, "no minutes found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "max number of minutes to return")
	cmd.Flags().StringVar(&folderID, "folder-id", "", "Drive folder token to list minutes from (default: root)")
	cmd.Flags().StringVar(&fileType, "type", "minutes", "Drive file type to match (default: minutes)")
	cmd.Flags().StringVar(&query, "query", "", "filter minutes by title substring")
	return cmd
}

func newMinutesDeleteCmd(state *appState) *cobra.Command {
	var minuteToken string
	var fileType string

	cmd := &cobra.Command{
		Use:   "delete <minute-token>",
		Short: "Delete Minutes",
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
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			resolvedType := strings.TrimSpace(fileType)
			if resolvedType == "" {
				file, err := state.SDK.GetDriveFileMetadata(context.Background(), token, larksdk.GetDriveFileRequest{
					FileToken: minuteToken,
				})
				if err != nil {
					return fmt.Errorf("file type is required (failed to resolve minutes type: %w)", err)
				}
				resolvedType = strings.TrimSpace(file.FileType)
			}
			if resolvedType == "" {
				return errors.New("file type is required")
			}
			result, err := state.SDK.DeleteDriveFile(context.Background(), token, minuteToken, resolvedType)
			if err != nil {
				return err
			}
			payload := map[string]any{
				"delete":       result,
				"minute_token": minuteToken,
				"type":         resolvedType,
			}
			text := tableTextRow(
				[]string{"minute_token", "type", "task_id"},
				[]string{minuteToken, resolvedType, result.TaskID},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&minuteToken, "minute-token", "", "minute token (or provide as positional argument)")
	cmd.Flags().StringVar(&fileType, "type", "", "Drive file type to delete (default: auto-detect via drive metadata)")
	_ = cmd.MarkFlagRequired("minute-token")
	return cmd
}

func newMinutesUpdateCmd(state *appState) *cobra.Command {
	var minuteToken string
	var linkShare string
	var externalAccess bool
	var inviteExternal bool
	var shareEntity string
	var securityEntity string
	var commentEntity string

	cmd := &cobra.Command{
		Use:   "update <minute-token>",
		Short: "Update Minutes sharing permissions",
		Args:  cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return nil
			}
			if minuteToken != "" && minuteToken != args[0] {
				return errors.New("minute-token provided twice")
			}
			return cmd.Flags().Set("minute-token", args[0])
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			req := larksdk.UpdateDrivePermissionPublicRequest{
				LinkShareEntity: linkShare,
				ShareEntity:     shareEntity,
				SecurityEntity:  securityEntity,
				CommentEntity:   commentEntity,
			}
			if cmd.Flags().Changed("external-access") {
				req.ExternalAccess = &externalAccess
			}
			if cmd.Flags().Changed("invite-external") {
				req.InviteExternal = &inviteExternal
			}
			permission, err := state.SDK.UpdateDrivePermissionPublic(context.Background(), token, minuteToken, "minutes", req)
			if err != nil {
				return err
			}
			payload := map[string]any{
				"permission":   permission,
				"minute_token": minuteToken,
				"type":         "minutes",
			}
			text := tableTextRow(
				[]string{"minute_token", "type", "link_share", "external_access", "invite_external"},
				[]string{
					minuteToken,
					"minutes",
					permission.LinkShareEntity,
					fmt.Sprintf("%t", permission.ExternalAccess),
					fmt.Sprintf("%t", permission.InviteExternal),
				},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&minuteToken, "minute-token", "", "minute token (or provide as positional argument)")
	cmd.Flags().StringVar(&linkShare, "link-share", "", "link share permission (for example: tenant_readable, anyone_readable)")
	cmd.Flags().BoolVar(&externalAccess, "external-access", false, "allow external access")
	cmd.Flags().BoolVar(&inviteExternal, "invite-external", false, "allow external invite")
	cmd.Flags().StringVar(&shareEntity, "share-entity", "", "share permission scope (for example: tenant_editable)")
	cmd.Flags().StringVar(&securityEntity, "security-entity", "", "security permission scope (for example: tenant_editable)")
	cmd.Flags().StringVar(&commentEntity, "comment-entity", "", "comment permission scope (for example: tenant_editable)")
	_ = cmd.MarkFlagRequired("minute-token")
	cmd.MarkFlagsOneRequired("link-share", "external-access", "invite-external", "share-entity", "security-entity", "comment-entity")
	return cmd
}
