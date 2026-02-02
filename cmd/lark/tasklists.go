package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newTasklistsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tasklists",
		Short: "Manage task lists",
		Long: `Task lists group related tasks.

- tasklist-guid identifies a list (UUID-like string).
- create/update support members and ownership updates.`,
	}
	cmd.AddCommand(newTasklistCreateCmd(state))
	cmd.AddCommand(newTasklistInfoCmd(state))
	cmd.AddCommand(newTasklistUpdateCmd(state))
	cmd.AddCommand(newTasklistDeleteCmd(state))
	return cmd
}

func newTasklistCreateCmd(state *appState) *cobra.Command {
	var name string
	var editors []string
	var viewers []string
	var memberType string
	var membersJSON string
	var userIDType string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a task list",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			if strings.TrimSpace(name) == "" {
				return errors.New("name is required")
			}
			if membersJSON != "" && (len(editors) > 0 || len(viewers) > 0) {
				return errors.New("--members-json cannot be combined with --editor/--viewer")
			}
			members, err := buildTasklistMembers(memberType, editors, viewers, membersJSON)
			if err != nil {
				return err
			}
			req := larksdk.CreateTasklistRequest{
				Name:       strings.TrimSpace(name),
				Members:    members,
				UserIDType: strings.TrimSpace(userIDType),
			}

			ctx := context.Background()
			token, tokenType, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			var tasklist larksdk.TaskList
			switch tokenType {
			case tokenTypeTenant:
				tasklist, err = state.SDK.CreateTasklist(ctx, token, req)
			case tokenTypeUser:
				tasklist, err = state.SDK.CreateTasklistWithUserToken(ctx, token, req)
			default:
				return fmt.Errorf("unsupported token type %s", tokenType)
			}
			if err != nil {
				return err
			}
			payload := map[string]any{"tasklist": tasklist}
			text := tableTextRow(
				[]string{"tasklist_guid", "name", "url"},
				[]string{tasklist.GUID, tasklist.Name, tasklist.URL},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "task list name")
	cmd.Flags().StringArrayVar(&editors, "editor", nil, "editor member IDs (repeatable)")
	cmd.Flags().StringArrayVar(&viewers, "viewer", nil, "viewer member IDs (repeatable)")
	cmd.Flags().StringVar(&memberType, "member-type", "user", "member type for --editor/--viewer (user, chat, app)")
	cmd.Flags().StringVar(&membersJSON, "members-json", "", "raw members JSON array (advanced)")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "open_id", "user ID type (open_id, union_id, user_id)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newTasklistInfoCmd(state *appState) *cobra.Command {
	var userIDType string

	cmd := &cobra.Command{
		Use:   "info <tasklist-guid>",
		Short: "Show task list details",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if strings.TrimSpace(args[0]) == "" {
				return argsUsageError(cmd, errors.New("tasklist-guid is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			tasklistGUID := strings.TrimSpace(args[0])
			if err := confirmDestructive(cmd, state, fmt.Sprintf("delete task list %s", tasklistGUID)); err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			ctx := cmd.Context()
			token, tokenType, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			req := larksdk.GetTasklistRequest{
				TasklistGUID: tasklistGUID,
				UserIDType:   strings.TrimSpace(userIDType),
			}
			var tasklist larksdk.TaskList
			switch tokenType {
			case tokenTypeTenant:
				tasklist, err = state.SDK.GetTasklist(ctx, token, req)
			case tokenTypeUser:
				tasklist, err = state.SDK.GetTasklistWithUserToken(ctx, token, req)
			default:
				return fmt.Errorf("unsupported token type %s", tokenType)
			}
			if err != nil {
				return err
			}
			payload := map[string]any{"tasklist": tasklist}
			rows := tasklistDetailRows(tasklist)
			text := tableTextFromRows([]string{"name", "value"}, rows, "no tasklist found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&userIDType, "user-id-type", "open_id", "user ID type (open_id, union_id, user_id)")
	return cmd
}

func newTasklistUpdateCmd(state *appState) *cobra.Command {
	var name string
	var ownerID string
	var ownerType string
	var originOwnerRole string
	var userIDType string

	cmd := &cobra.Command{
		Use:   "update <tasklist-guid>",
		Short: "Update a task list",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if strings.TrimSpace(args[0]) == "" {
				return errors.New("tasklist-guid is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			tasklistGUID := strings.TrimSpace(args[0])
			updateFields := make([]string, 0, 2)
			tasklistPayload := map[string]any{}

			if cmd.Flags().Changed("name") {
				if strings.TrimSpace(name) == "" {
					return errors.New("name cannot be empty")
				}
				tasklistPayload["name"] = strings.TrimSpace(name)
				updateFields = append(updateFields, "name")
			}
			if cmd.Flags().Changed("owner-id") || cmd.Flags().Changed("owner-type") {
				if strings.TrimSpace(ownerID) == "" {
					return errors.New("owner-id is required")
				}
				if strings.TrimSpace(ownerType) == "" {
					ownerType = "user"
				}
				tasklistPayload["owner"] = larksdk.TaskMember{
					ID:   strings.TrimSpace(ownerID),
					Type: strings.TrimSpace(ownerType),
					Role: "owner",
				}
				updateFields = append(updateFields, "owner")
			}
			if strings.TrimSpace(originOwnerRole) != "" && !containsString(updateFields, "owner") {
				return errors.New("--origin-owner-role requires --owner-id")
			}
			if len(updateFields) == 0 {
				return errors.New("at least one field must be provided")
			}

			req := larksdk.UpdateTasklistRequest{
				TasklistGUID:      tasklistGUID,
				Tasklist:          tasklistPayload,
				UpdateFields:      updateFields,
				OriginOwnerToRole: strings.TrimSpace(originOwnerRole),
				UserIDType:        strings.TrimSpace(userIDType),
			}

			ctx := context.Background()
			token, tokenType, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			var tasklist larksdk.TaskList
			switch tokenType {
			case tokenTypeTenant:
				tasklist, err = state.SDK.UpdateTasklist(ctx, token, req)
			case tokenTypeUser:
				tasklist, err = state.SDK.UpdateTasklistWithUserToken(ctx, token, req)
			default:
				return fmt.Errorf("unsupported token type %s", tokenType)
			}
			if err != nil {
				return err
			}
			payload := map[string]any{"tasklist": tasklist}
			text := tableTextRow(
				[]string{"tasklist_guid", "name", "url"},
				[]string{tasklist.GUID, tasklist.Name, tasklist.URL},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "task list name")
	cmd.Flags().StringVar(&ownerID, "owner-id", "", "new owner user/app ID")
	cmd.Flags().StringVar(&ownerType, "owner-type", "user", "owner member type (user or app)")
	cmd.Flags().StringVar(&originOwnerRole, "origin-owner-role", "", "role for original owner (editor, viewer, none)")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "open_id", "user ID type (open_id, union_id, user_id)")
	return cmd
}

func newTasklistDeleteCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <tasklist-guid>",
		Short: "Delete a task list",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if strings.TrimSpace(args[0]) == "" {
				return errors.New("tasklist-guid is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			tasklistGUID := strings.TrimSpace(args[0])
			ctx := context.Background()
			token, tokenType, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			switch tokenType {
			case tokenTypeTenant:
				err = state.SDK.DeleteTasklist(ctx, token, tasklistGUID)
			case tokenTypeUser:
				err = state.SDK.DeleteTasklistWithUserToken(ctx, token, tasklistGUID)
			default:
				return fmt.Errorf("unsupported token type %s", tokenType)
			}
			if err != nil {
				return err
			}
			payload := map[string]any{"tasklist_guid": tasklistGUID, "deleted": true}
			text := tableTextRow([]string{"tasklist_guid", "deleted"}, []string{tasklistGUID, "true"})
			return state.Printer.Print(payload, text)
		},
	}
	return cmd
}

func buildTasklistMembers(memberType string, editors, viewers []string, membersJSON string) ([]larksdk.TaskMember, error) {
	if strings.TrimSpace(membersJSON) != "" {
		var members []larksdk.TaskMember
		if err := json.Unmarshal([]byte(membersJSON), &members); err != nil {
			return nil, fmt.Errorf("members-json must be a JSON array: %w", err)
		}
		return members, nil
	}
	memberType = strings.TrimSpace(memberType)
	if memberType == "" {
		memberType = "user"
	}
	members := make([]larksdk.TaskMember, 0, len(editors)+len(viewers))
	for _, id := range editors {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		members = append(members, larksdk.TaskMember{ID: id, Type: memberType, Role: "editor"})
	}
	for _, id := range viewers {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		members = append(members, larksdk.TaskMember{ID: id, Type: memberType, Role: "viewer"})
	}
	return members, nil
}

func tasklistDetailRows(tasklist larksdk.TaskList) [][]string {
	rows := make([][]string, 0, 10)
	add := func(name, value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		rows = append(rows, []string{name, value})
	}
	add("guid", tasklist.GUID)
	add("name", tasklist.Name)
	if tasklist.Owner != nil {
		add("owner", tasklist.Owner.ID)
	}
	add("url", tasklist.URL)
	add("created_at", tasklist.CreatedAt)
	add("updated_at", tasklist.UpdatedAt)
	if len(tasklist.Members) > 0 {
		add("members", fmt.Sprintf("%d", len(tasklist.Members)))
	}
	return rows
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
