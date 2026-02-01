package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newDrivePermissionsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "permissions",
		Short: "Manage Drive file permissions",
		Long: `Manage Drive file collaborator permissions.

- file-token is the Drive file token (docx id, spreadsheet id, or file token).`,
	}
	cmd.AddCommand(newDrivePermissionAddCmd(state))
	cmd.AddCommand(newDrivePermissionListCmd(state))
	cmd.AddCommand(newDrivePermissionUpdateCmd(state))
	cmd.AddCommand(newDrivePermissionDeleteCmd(state))
	return cmd
}

func newDrivePermissionAddCmd(state *appState) *cobra.Command {
	var fileToken string
	var fileType string
	var memberType string
	var memberID string
	var perm string
	var permType string
	var memberKind string
	var needNotification bool

	cmd := &cobra.Command{
		Use:   "add <file-token> <member-type> <member-id>",
		Short: "Add a Drive collaborator permission",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(3)(cmd, args); err != nil {
				return err
			}
			if len(args) > 0 {
				if fileToken != "" && fileToken != args[0] {
					return errors.New("file-token provided twice")
				}
				if err := cmd.Flags().Set("file-token", args[0]); err != nil {
					return err
				}
			}
			if len(args) > 1 {
				if memberType != "" && memberType != args[1] {
					return errors.New("member-type provided twice")
				}
				if err := cmd.Flags().Set("member-type", args[1]); err != nil {
					return err
				}
			}
			if len(args) > 2 {
				if memberID != "" && memberID != args[2] {
					return errors.New("member-id provided twice")
				}
				if err := cmd.Flags().Set("member-id", args[2]); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			if strings.TrimSpace(fileToken) == "" {
				return errors.New("file-token is required")
			}
			ctx := context.Background()
			token, tokenType, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}

			needNotificationSet := cmd.Flags().Changed("need-notification")
			req := larksdk.AddDrivePermissionMemberRequest{
				MemberType:          strings.TrimSpace(memberType),
				MemberID:            strings.TrimSpace(memberID),
				Perm:                strings.TrimSpace(perm),
				PermType:            strings.TrimSpace(permType),
				Type:                strings.TrimSpace(memberKind),
				NeedNotification:    needNotification,
				NeedNotificationSet: needNotificationSet,
			}

			var member larksdk.DrivePermissionMember
			switch tokenType {
			case tokenTypeUser:
				member, err = state.SDK.AddDrivePermissionMemberWithUserToken(ctx, token, fileToken, fileType, req)
			case tokenTypeTenant:
				member, err = state.SDK.AddDrivePermissionMember(ctx, token, fileToken, fileType, req)
			default:
				return fmt.Errorf("unsupported token type %s", tokenType)
			}
			if err != nil {
				return err
			}

			payload := map[string]any{
				"member":     member,
				"file_token": fileToken,
				"type":       fileType,
			}
			text := tableTextRow(
				[]string{"member_type", "member_id", "perm", "perm_type", "type"},
				[]string{member.MemberType, member.MemberID, member.Perm, member.PermType, member.Type},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&fileToken, "file-token", "", "Drive file token (docx/sheet id; or provide as positional argument)")
	cmd.Flags().StringVar(&fileType, "type", "", "Drive file type (docx, sheet, file, wiki, bitable, folder, mindnote, minutes, slides)")
	cmd.Flags().StringVar(&memberType, "member-type", "", "member id type (email, openid, unionid, openchat, opendepartmentid, userid, groupid, wikispaceid) (or provide as positional argument)")
	cmd.Flags().StringVar(&memberID, "member-id", "", "member id (or provide as positional argument)")
	cmd.Flags().StringVar(&perm, "perm", "", "permission role (view, edit, full_access)")
	cmd.Flags().StringVar(&permType, "perm-type", "", "permission scope for wiki (container, single_page)")
	cmd.Flags().StringVar(&memberKind, "member-kind", "", "collaborator type (user, chat, department, group, wiki_space_member, wiki_space_viewer, wiki_space_editor)")
	cmd.Flags().BoolVar(&needNotification, "need-notification", false, "notify the member after adding permissions (user token only)")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("member-type")
	_ = cmd.MarkFlagRequired("member-id")
	_ = cmd.MarkFlagRequired("perm")
	return cmd
}

func newDrivePermissionListCmd(state *appState) *cobra.Command {
	var fileToken string
	var fileType string
	var fields string
	var permType string

	cmd := &cobra.Command{
		Use:   "list <file-token>",
		Short: "List Drive collaborator permissions",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) > 0 {
				if fileToken != "" && fileToken != args[0] {
					return errors.New("file-token provided twice")
				}
				if err := cmd.Flags().Set("file-token", args[0]); err != nil {
					return err
				}
			}
			if strings.TrimSpace(fileToken) == "" {
				return errors.New("file-token is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			ctx := context.Background()
			token, tokenType, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}

			req := larksdk.ListDrivePermissionMembersRequest{
				Fields:   strings.TrimSpace(fields),
				PermType: strings.TrimSpace(permType),
			}
			var result larksdk.ListDrivePermissionMembersResult
			switch tokenType {
			case tokenTypeUser:
				result, err = state.SDK.ListDrivePermissionMembersWithUserToken(ctx, token, fileToken, fileType, req)
			case tokenTypeTenant:
				result, err = state.SDK.ListDrivePermissionMembers(ctx, token, fileToken, fileType, req)
			default:
				return fmt.Errorf("unsupported token type %s", tokenType)
			}
			if err != nil {
				return err
			}

			payload := map[string]any{
				"members":    result.Items,
				"file_token": fileToken,
				"type":       fileType,
			}
			lines := make([]string, 0, len(result.Items))
			for _, member := range result.Items {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s", member.MemberType, member.MemberID, member.Perm, member.PermType, member.Type, member.Name))
			}
			text := tableText([]string{"member_type", "member_id", "perm", "perm_type", "type", "name"}, lines, "no members found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&fileToken, "file-token", "", "Drive file token (docx/sheet id; or provide as positional argument)")
	cmd.Flags().StringVar(&fileType, "type", "", "Drive file type (docx, sheet, file, wiki, bitable, folder, mindnote, minutes, slides)")
	cmd.Flags().StringVar(&fields, "fields", "", "fields to return (comma-separated, optional)")
	cmd.Flags().StringVar(&permType, "perm-type", "", "permission scope for wiki (container, single_page)")
	_ = cmd.MarkFlagRequired("type")
	return cmd
}

func newDrivePermissionUpdateCmd(state *appState) *cobra.Command {
	var fileToken string
	var fileType string
	var memberType string
	var memberID string
	var perm string
	var permType string
	var memberKind string
	var needNotification bool

	cmd := &cobra.Command{
		Use:   "update <file-token> <member-type> <member-id>",
		Short: "Update a Drive collaborator permission",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(3)(cmd, args); err != nil {
				return err
			}
			if len(args) > 0 {
				if fileToken != "" && fileToken != args[0] {
					return errors.New("file-token provided twice")
				}
				if err := cmd.Flags().Set("file-token", args[0]); err != nil {
					return err
				}
			}
			if len(args) > 1 {
				if memberType != "" && memberType != args[1] {
					return errors.New("member-type provided twice")
				}
				if err := cmd.Flags().Set("member-type", args[1]); err != nil {
					return err
				}
			}
			if len(args) > 2 {
				if memberID != "" && memberID != args[2] {
					return errors.New("member-id provided twice")
				}
				if err := cmd.Flags().Set("member-id", args[2]); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			if strings.TrimSpace(fileToken) == "" {
				return errors.New("file-token is required")
			}
			if strings.TrimSpace(perm) == "" && strings.TrimSpace(permType) == "" && strings.TrimSpace(memberKind) == "" {
				return errors.New("at least one of --perm, --perm-type, or --member-kind is required")
			}
			ctx := context.Background()
			token, tokenType, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}

			needNotificationSet := cmd.Flags().Changed("need-notification")
			req := larksdk.UpdateDrivePermissionMemberRequest{
				MemberType:          strings.TrimSpace(memberType),
				MemberID:            strings.TrimSpace(memberID),
				Perm:                strings.TrimSpace(perm),
				PermType:            strings.TrimSpace(permType),
				Type:                strings.TrimSpace(memberKind),
				NeedNotification:    needNotification,
				NeedNotificationSet: needNotificationSet,
			}

			var member larksdk.DrivePermissionMember
			switch tokenType {
			case tokenTypeUser:
				member, err = state.SDK.UpdateDrivePermissionMemberWithUserToken(ctx, token, fileToken, fileType, req)
			case tokenTypeTenant:
				member, err = state.SDK.UpdateDrivePermissionMember(ctx, token, fileToken, fileType, req)
			default:
				return fmt.Errorf("unsupported token type %s", tokenType)
			}
			if err != nil {
				return err
			}

			payload := map[string]any{
				"member":     member,
				"file_token": fileToken,
				"type":       fileType,
			}
			text := tableTextRow(
				[]string{"member_type", "member_id", "perm", "perm_type", "type"},
				[]string{member.MemberType, member.MemberID, member.Perm, member.PermType, member.Type},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&fileToken, "file-token", "", "Drive file token (docx/sheet id; or provide as positional argument)")
	cmd.Flags().StringVar(&fileType, "type", "", "Drive file type (docx, sheet, file, wiki, bitable, folder, mindnote, minutes, slides)")
	cmd.Flags().StringVar(&memberType, "member-type", "", "member id type (email, openid, unionid, openchat, opendepartmentid, userid, groupid, wikispaceid) (or provide as positional argument)")
	cmd.Flags().StringVar(&memberID, "member-id", "", "member id (or provide as positional argument)")
	cmd.Flags().StringVar(&perm, "perm", "", "permission role (view, edit, full_access)")
	cmd.Flags().StringVar(&permType, "perm-type", "", "permission scope for wiki (container, single_page)")
	cmd.Flags().StringVar(&memberKind, "member-kind", "", "collaborator type (user, chat, department, group, wiki_space_member, wiki_space_viewer, wiki_space_editor)")
	cmd.Flags().BoolVar(&needNotification, "need-notification", false, "notify the member after updating permissions (user token only)")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("member-type")
	_ = cmd.MarkFlagRequired("member-id")
	return cmd
}

func newDrivePermissionDeleteCmd(state *appState) *cobra.Command {
	var fileToken string
	var fileType string
	var memberType string
	var memberID string
	var permType string
	var memberKind string

	cmd := &cobra.Command{
		Use:   "delete <file-token> <member-type> <member-id>",
		Short: "Delete a Drive collaborator permission",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(3)(cmd, args); err != nil {
				return err
			}
			if len(args) > 0 {
				if fileToken != "" && fileToken != args[0] {
					return errors.New("file-token provided twice")
				}
				if err := cmd.Flags().Set("file-token", args[0]); err != nil {
					return err
				}
			}
			if len(args) > 1 {
				if memberType != "" && memberType != args[1] {
					return errors.New("member-type provided twice")
				}
				if err := cmd.Flags().Set("member-type", args[1]); err != nil {
					return err
				}
			}
			if len(args) > 2 {
				if memberID != "" && memberID != args[2] {
					return errors.New("member-id provided twice")
				}
				if err := cmd.Flags().Set("member-id", args[2]); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			if strings.TrimSpace(fileToken) == "" {
				return errors.New("file-token is required")
			}
			ctx := context.Background()
			token, tokenType, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}

			req := larksdk.DeleteDrivePermissionMemberRequest{
				MemberType: strings.TrimSpace(memberType),
				MemberID:   strings.TrimSpace(memberID),
				PermType:   strings.TrimSpace(permType),
				Type:       strings.TrimSpace(memberKind),
			}
			switch tokenType {
			case tokenTypeUser:
				err = state.SDK.DeleteDrivePermissionMemberWithUserToken(ctx, token, fileToken, fileType, req)
			case tokenTypeTenant:
				err = state.SDK.DeleteDrivePermissionMember(ctx, token, fileToken, fileType, req)
			default:
				return fmt.Errorf("unsupported token type %s", tokenType)
			}
			if err != nil {
				return err
			}

			payload := map[string]any{
				"deleted":    true,
				"member":     req,
				"file_token": fileToken,
				"type":       fileType,
			}
			text := "deleted"
			if req.MemberID != "" {
				text = req.MemberID
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&fileToken, "file-token", "", "Drive file token (docx/sheet id; or provide as positional argument)")
	cmd.Flags().StringVar(&fileType, "type", "", "Drive file type (docx, sheet, file, wiki, bitable, folder, mindnote, minutes, slides)")
	cmd.Flags().StringVar(&memberType, "member-type", "", "member id type (email, openid, unionid, openchat, opendepartmentid, userid, groupid, wikispaceid) (or provide as positional argument)")
	cmd.Flags().StringVar(&memberID, "member-id", "", "member id (or provide as positional argument)")
	cmd.Flags().StringVar(&permType, "perm-type", "", "permission scope for wiki (container, single_page)")
	cmd.Flags().StringVar(&memberKind, "member-kind", "", "collaborator type (user, chat, department, group, wiki_space_member, wiki_space_viewer, wiki_space_editor)")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("member-type")
	_ = cmd.MarkFlagRequired("member-id")
	return cmd
}
