package main

import (
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

- file-token is the Drive file token (docx document_id, spreadsheet_token, or file_token).`,
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
		Long: `Add a Drive collaborator permission.

Arguments:
  file-token: Drive file token (docx document_id, spreadsheet_token, or file_token).
  member-type: collaborator identifier type (openid, userid, email, openchat, opendepartmentid).
  member-id: identifier value for the chosen member-type.
`,
		Example: `  lark drive permissions add <file-token> openid ou_xxx --type docx --perm view --member-kind user`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(3)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			refToken, refKind, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			fileToken = strings.TrimSpace(refToken)
			memberType = strings.TrimSpace(args[1])
			memberID = strings.TrimSpace(args[2])
			if fileToken == "" {
				return argsUsageError(cmd, errors.New("file-token is required"))
			}
			if memberType == "" {
				return argsUsageError(cmd, errors.New("member-type is required"))
			}
			normalizedType, ok := normalizeDriveMemberType(memberType)
			if !ok {
				return argsUsageError(cmd, fmt.Errorf("member-type must be one of %s", strings.Join(driveMemberTypeValues, ", ")))
			}
			memberType = normalizedType
			if memberID == "" {
				return argsUsageError(cmd, errors.New("member-id is required"))
			}
			if refKind != "" && !cmd.Flags().Changed("type") {
				_ = cmd.Flags().Set("type", refKind)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOneOf(cmd, "perm", perm, drivePermValues); err != nil {
				return err
			}
			if err := validateOneOf(cmd, "perm-type", permType, drivePermTypeValues); err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			ctx := cmd.Context()
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

	cmd.Flags().StringVar(&fileType, "type", "", "Drive file type (docx, sheet, file, wiki, bitable, folder, mindnote, minutes, slides)")
	cmd.Flags().StringVar(&perm, "perm", "", "permission role (view, edit, full_access)")
	cmd.Flags().StringVar(&permType, "perm-type", "", "permission scope for wiki (container, single_page)")
	cmd.Flags().StringVar(&memberKind, "member-kind", "", "collaborator type (user, chat, department, group, wiki_space_member, wiki_space_viewer, wiki_space_editor)")
	cmd.Flags().BoolVar(&needNotification, "need-notification", false, "notify the member after adding permissions (user token only)")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("perm")
	registerEnumCompletion(cmd, "perm", drivePermValues)
	registerEnumCompletion(cmd, "perm-type", drivePermTypeValues)
	return cmd
}

func newDrivePermissionListCmd(state *appState) *cobra.Command {
	var fileToken string
	var fileType string
	var fields string
	var permType string

	cmd := &cobra.Command{
		Use:     "list <file-token>",
		Short:   "List Drive collaborator permissions",
		Example: `  lark drive permissions list <file-token> --type docx`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			refToken, refKind, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			fileToken = strings.TrimSpace(refToken)
			if fileToken == "" {
				return usageError(cmd, "file-token is required", `Example:
  lark drive permissions list <file-token> --type docx`)
			}
			if refKind != "" && !cmd.Flags().Changed("type") {
				_ = cmd.Flags().Set("type", refKind)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := requireSDK(state); err != nil {
				return err
			}
			ctx := cmd.Context()
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
		Use:     "update <file-token> <member-type> <member-id>",
		Short:   "Update a Drive collaborator permission",
		Example: `  lark drive permissions update <file-token> email fred@srv.work --type docx --perm edit`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(3)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			refToken, refKind, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			fileToken = strings.TrimSpace(refToken)
			memberType = strings.TrimSpace(args[1])
			memberID = strings.TrimSpace(args[2])
			if fileToken == "" {
				return argsUsageError(cmd, errors.New("file-token is required"))
			}
			if memberType == "" {
				return argsUsageError(cmd, errors.New("member-type is required"))
			}
			normalizedType, ok := normalizeDriveMemberType(memberType)
			if !ok {
				return argsUsageError(cmd, fmt.Errorf("member-type must be one of %s", strings.Join(driveMemberTypeValues, ", ")))
			}
			memberType = normalizedType
			if memberID == "" {
				return argsUsageError(cmd, errors.New("member-id is required"))
			}
			if refKind != "" && !cmd.Flags().Changed("type") {
				_ = cmd.Flags().Set("type", refKind)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOneOf(cmd, "perm", perm, drivePermValues); err != nil {
				return err
			}
			if err := validateOneOf(cmd, "perm-type", permType, drivePermTypeValues); err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			if strings.TrimSpace(perm) == "" && strings.TrimSpace(permType) == "" && strings.TrimSpace(memberKind) == "" {
				return usageError(cmd, "at least one of --perm, --perm-type, or --member-kind is required", `Example:
  lark drive permissions update <file-token> email fred@srv.work --type docx --perm edit`)
			}
			ctx := cmd.Context()
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

	cmd.Flags().StringVar(&fileType, "type", "", "Drive file type (docx, sheet, file, wiki, bitable, folder, mindnote, minutes, slides)")
	cmd.Flags().StringVar(&perm, "perm", "", "permission role (view, edit, full_access)")
	cmd.Flags().StringVar(&permType, "perm-type", "", "permission scope for wiki (container, single_page)")
	cmd.Flags().StringVar(&memberKind, "member-kind", "", "collaborator type (user, chat, department, group, wiki_space_member, wiki_space_viewer, wiki_space_editor)")
	cmd.Flags().BoolVar(&needNotification, "need-notification", false, "notify the member after updating permissions (user token only)")
	_ = cmd.MarkFlagRequired("type")
	registerEnumCompletion(cmd, "perm", drivePermValues)
	registerEnumCompletion(cmd, "perm-type", drivePermTypeValues)
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
		Use:     "delete <file-token> <member-type> <member-id>",
		Short:   "Delete a Drive collaborator permission",
		Example: `  lark drive permissions delete <file-token> email fred@srv.work --type docx`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(3)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			refToken, refKind, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			fileToken = strings.TrimSpace(refToken)
			memberType = strings.TrimSpace(args[1])
			memberID = strings.TrimSpace(args[2])
			if fileToken == "" {
				return argsUsageError(cmd, errors.New("file-token is required"))
			}
			if memberType == "" {
				return argsUsageError(cmd, errors.New("member-type is required"))
			}
			normalizedType, ok := normalizeDriveMemberType(memberType)
			if !ok {
				return argsUsageError(cmd, fmt.Errorf("member-type must be one of %s", strings.Join(driveMemberTypeValues, ", ")))
			}
			memberType = normalizedType
			if memberID == "" {
				return argsUsageError(cmd, errors.New("member-id is required"))
			}
			if refKind != "" && !cmd.Flags().Changed("type") {
				_ = cmd.Flags().Set("type", refKind)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := confirmDestructive(cmd, state, fmt.Sprintf("delete permission for %s on %s", memberID, fileToken)); err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			ctx := cmd.Context()
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

	cmd.Flags().StringVar(&fileType, "type", "", "Drive file type (docx, sheet, file, wiki, bitable, folder, mindnote, minutes, slides)")
	cmd.Flags().StringVar(&permType, "perm-type", "", "permission scope for wiki (container, single_page)")
	cmd.Flags().StringVar(&memberKind, "member-kind", "", "collaborator type (user, chat, department, group, wiki_space_member, wiki_space_viewer, wiki_space_editor)")
	_ = cmd.MarkFlagRequired("type")
	return cmd
}
