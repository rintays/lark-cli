package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

const maxTasksPageSize = 100

func newTasksCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "Manage tasks",
		Long: `Tasks are personal or shared work items.

- task-guid identifies a task (UUID-like string).
- list returns "my_tasks" and requires a user access token.
- create/update support due/start timestamps (ms or RFC3339); use --*-all-day for date-only.`,
	}
	cmd.AddCommand(newTaskCreateCmd(state))
	cmd.AddCommand(newTaskInfoCmd(state))
	cmd.AddCommand(newTaskUpdateCmd(state))
	cmd.AddCommand(newTaskDeleteCmd(state))
	cmd.AddCommand(newTaskListCmd(state))
	return cmd
}

func newTaskCreateCmd(state *appState) *cobra.Command {
	var summary string
	var description string
	var due string
	var dueAllDay bool
	var start string
	var startAllDay bool
	var completedAt string
	var extra string
	var repeatRule string
	var mode int
	var milestone bool
	var clientToken string
	var userIDType string
	var memberType string
	var assignees []string
	var followers []string
	var membersJSON string
	var tasklists []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a task",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			if strings.TrimSpace(summary) == "" {
				return errors.New("summary is required")
			}
			if membersJSON != "" && (len(assignees) > 0 || len(followers) > 0) {
				return errors.New("--members-json cannot be combined with --assignee/--follower")
			}
			if memberType == "" {
				memberType = "user"
			}
			members, err := buildTaskMembers(memberType, assignees, followers, membersJSON)
			if err != nil {
				return err
			}
			tasklistInfos, err := parseTasklistRefs(tasklists)
			if err != nil {
				return err
			}
			dueTime, err := buildTaskTime(cmd, due, "due-all-day", dueAllDay)
			if err != nil {
				return err
			}
			startTime, err := buildTaskTime(cmd, start, "start-all-day", startAllDay)
			if err != nil {
				return err
			}
			var completedAtValue *string
			if cmd.Flags().Changed("completed-at") {
				value, err := parseTaskCompletedAt(completedAt)
				if err != nil {
					return err
				}
				completedAtValue = &value
			}
			var modeValue *int
			if cmd.Flags().Changed("mode") {
				if mode < 1 || mode > 2 {
					return errors.New("mode must be 1 or 2")
				}
				modeValue = &mode
			}
			var milestoneValue *bool
			if cmd.Flags().Changed("milestone") {
				milestoneValue = &milestone
			}

			req := larksdk.CreateTaskRequest{
				Summary:     strings.TrimSpace(summary),
				Description: strings.TrimSpace(description),
				Due:         dueTime,
				Start:       startTime,
				Members:     members,
				Tasklists:   tasklistInfos,
				ClientToken: strings.TrimSpace(clientToken),
				CompletedAt: completedAtValue,
				Extra:       extra,
				RepeatRule:  repeatRule,
				Mode:        modeValue,
				IsMilestone: milestoneValue,
				UserIDType:  strings.TrimSpace(userIDType),
			}

			ctx := context.Background()
			token, tokenType, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			var task larksdk.Task
			switch tokenType {
			case tokenTypeTenant:
				task, err = state.SDK.CreateTask(ctx, token, req)
			case tokenTypeUser:
				task, err = state.SDK.CreateTaskWithUserToken(ctx, token, req)
			default:
				return fmt.Errorf("unsupported token type %s", tokenType)
			}
			if err != nil {
				return err
			}
			payload := map[string]any{"task": task}
			text := tableTextRow(
				[]string{"task_guid", "summary", "status", "due"},
				[]string{task.GUID, task.Summary, task.Status, formatTaskTime(task.Due)},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&summary, "summary", "", "task summary/title")
	cmd.Flags().StringVar(&description, "description", "", "task description")
	cmd.Flags().StringVar(&due, "due", "", "due timestamp in ms/seconds or RFC3339")
	cmd.Flags().BoolVar(&dueAllDay, "due-all-day", false, "treat due timestamp as date-only")
	cmd.Flags().StringVar(&start, "start", "", "start timestamp in ms/seconds or RFC3339")
	cmd.Flags().BoolVar(&startAllDay, "start-all-day", false, "treat start timestamp as date-only")
	cmd.Flags().StringVar(&completedAt, "completed-at", "", "completed timestamp in ms/seconds or RFC3339 (use 0 to mark incomplete)")
	cmd.Flags().StringVar(&extra, "extra", "", "extra payload (string, returned as-is)")
	cmd.Flags().StringVar(&repeatRule, "repeat-rule", "", "RRULE string for recurring tasks")
	cmd.Flags().IntVar(&mode, "mode", 0, "completion mode: 1 (AND) or 2 (OR)")
	cmd.Flags().BoolVar(&milestone, "milestone", false, "mark task as milestone")
	cmd.Flags().StringVar(&clientToken, "client-token", "", "idempotency token")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "open_id", "user ID type (open_id, union_id, user_id)")
	cmd.Flags().StringVar(&memberType, "member-type", "user", "member type for --assignee/--follower (user or app)")
	cmd.Flags().StringArrayVar(&assignees, "assignee", nil, "assignee member IDs (repeatable)")
	cmd.Flags().StringArrayVar(&followers, "follower", nil, "follower member IDs (repeatable)")
	cmd.Flags().StringVar(&membersJSON, "members-json", "", "raw members JSON array (advanced)")
	cmd.Flags().StringArrayVar(&tasklists, "tasklist", nil, "tasklist reference: <tasklist-guid>[:section-guid] (repeatable)")
	_ = cmd.MarkFlagRequired("summary")
	return cmd
}

func newTaskInfoCmd(state *appState) *cobra.Command {
	var userIDType string

	cmd := &cobra.Command{
		Use:   "info <task-guid>",
		Short: "Show task details",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if strings.TrimSpace(args[0]) == "" {
				return argsUsageError(cmd, errors.New("task-guid is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			taskGUID := strings.TrimSpace(args[0])
			if err := confirmDestructive(cmd, state, fmt.Sprintf("delete task %s", taskGUID)); err != nil {
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
			req := larksdk.GetTaskRequest{
				TaskGUID:   taskGUID,
				UserIDType: strings.TrimSpace(userIDType),
			}
			var task larksdk.Task
			switch tokenType {
			case tokenTypeTenant:
				task, err = state.SDK.GetTask(ctx, token, req)
			case tokenTypeUser:
				task, err = state.SDK.GetTaskWithUserToken(ctx, token, req)
			default:
				return fmt.Errorf("unsupported token type %s", tokenType)
			}
			if err != nil {
				return err
			}
			payload := map[string]any{"task": task}
			rows := taskDetailRows(task)
			text := tableTextFromRows([]string{"name", "value"}, rows, "no task found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&userIDType, "user-id-type", "open_id", "user ID type (open_id, union_id, user_id)")
	return cmd
}

func newTaskUpdateCmd(state *appState) *cobra.Command {
	var summary string
	var description string
	var due string
	var dueAllDay bool
	var start string
	var startAllDay bool
	var clearDue bool
	var clearStart bool
	var completedAt string
	var extra string
	var repeatRule string
	var mode int
	var milestone bool
	var userIDType string

	cmd := &cobra.Command{
		Use:   "update <task-guid>",
		Short: "Update a task",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if strings.TrimSpace(args[0]) == "" {
				return errors.New("task-guid is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			if clearDue && due != "" {
				return errors.New("--clear-due cannot be combined with --due")
			}
			if clearStart && start != "" {
				return errors.New("--clear-start cannot be combined with --start")
			}
			taskPayload := map[string]any{}
			updateFields := make([]string, 0, 6)

			if cmd.Flags().Changed("summary") {
				if strings.TrimSpace(summary) == "" {
					return errors.New("summary cannot be empty")
				}
				taskPayload["summary"] = strings.TrimSpace(summary)
				updateFields = append(updateFields, "summary")
			}
			if cmd.Flags().Changed("description") {
				taskPayload["description"] = strings.TrimSpace(description)
				updateFields = append(updateFields, "description")
			}

			if clearDue {
				taskPayload["due"] = nil
				updateFields = append(updateFields, "due")
			} else if due != "" || cmd.Flags().Changed("due-all-day") {
				if due == "" {
					return errors.New("--due-all-day requires --due")
				}
				dueTime, err := buildTaskTime(cmd, due, "due-all-day", dueAllDay)
				if err != nil {
					return err
				}
				taskPayload["due"] = dueTime
				updateFields = append(updateFields, "due")
			}

			if clearStart {
				taskPayload["start"] = nil
				updateFields = append(updateFields, "start")
			} else if start != "" || cmd.Flags().Changed("start-all-day") {
				if start == "" {
					return errors.New("--start-all-day requires --start")
				}
				startTime, err := buildTaskTime(cmd, start, "start-all-day", startAllDay)
				if err != nil {
					return err
				}
				taskPayload["start"] = startTime
				updateFields = append(updateFields, "start")
			}

			if cmd.Flags().Changed("completed-at") {
				value, err := parseTaskCompletedAt(completedAt)
				if err != nil {
					return err
				}
				taskPayload["completed_at"] = value
				updateFields = append(updateFields, "completed_at")
			}
			if cmd.Flags().Changed("extra") {
				taskPayload["extra"] = extra
				updateFields = append(updateFields, "extra")
			}
			if cmd.Flags().Changed("repeat-rule") {
				taskPayload["repeat_rule"] = repeatRule
				updateFields = append(updateFields, "repeat_rule")
			}
			if cmd.Flags().Changed("mode") {
				if mode < 1 || mode > 2 {
					return errors.New("mode must be 1 or 2")
				}
				taskPayload["mode"] = mode
				updateFields = append(updateFields, "mode")
			}
			if cmd.Flags().Changed("milestone") {
				taskPayload["is_milestone"] = milestone
				updateFields = append(updateFields, "is_milestone")
			}

			if len(updateFields) == 0 {
				return errors.New("at least one field must be provided")
			}

			ctx := context.Background()
			token, tokenType, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			req := larksdk.UpdateTaskRequest{
				TaskGUID:     strings.TrimSpace(args[0]),
				Task:         taskPayload,
				UpdateFields: updateFields,
				UserIDType:   strings.TrimSpace(userIDType),
			}
			var task larksdk.Task
			switch tokenType {
			case tokenTypeTenant:
				task, err = state.SDK.UpdateTask(ctx, token, req)
			case tokenTypeUser:
				task, err = state.SDK.UpdateTaskWithUserToken(ctx, token, req)
			default:
				return fmt.Errorf("unsupported token type %s", tokenType)
			}
			if err != nil {
				return err
			}
			payload := map[string]any{"task": task}
			text := tableTextRow(
				[]string{"task_guid", "summary", "status", "due"},
				[]string{task.GUID, task.Summary, task.Status, formatTaskTime(task.Due)},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&summary, "summary", "", "task summary/title")
	cmd.Flags().StringVar(&description, "description", "", "task description")
	cmd.Flags().StringVar(&due, "due", "", "due timestamp in ms/seconds or RFC3339")
	cmd.Flags().BoolVar(&dueAllDay, "due-all-day", false, "treat due timestamp as date-only")
	cmd.Flags().StringVar(&start, "start", "", "start timestamp in ms/seconds or RFC3339")
	cmd.Flags().BoolVar(&startAllDay, "start-all-day", false, "treat start timestamp as date-only")
	cmd.Flags().BoolVar(&clearDue, "clear-due", false, "clear due time")
	cmd.Flags().BoolVar(&clearStart, "clear-start", false, "clear start time")
	cmd.Flags().StringVar(&completedAt, "completed-at", "", "completed timestamp in ms/seconds or RFC3339 (use 0 to mark incomplete)")
	cmd.Flags().StringVar(&extra, "extra", "", "extra payload (string, returned as-is)")
	cmd.Flags().StringVar(&repeatRule, "repeat-rule", "", "RRULE string for recurring tasks")
	cmd.Flags().IntVar(&mode, "mode", 0, "completion mode: 1 (AND) or 2 (OR)")
	cmd.Flags().BoolVar(&milestone, "milestone", false, "mark task as milestone")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "open_id", "user ID type (open_id, union_id, user_id)")
	return cmd
}

func newTaskDeleteCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <task-guid>",
		Short: "Delete a task",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if strings.TrimSpace(args[0]) == "" {
				return errors.New("task-guid is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			taskGUID := strings.TrimSpace(args[0])
			ctx := context.Background()
			token, tokenType, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			switch tokenType {
			case tokenTypeTenant:
				err = state.SDK.DeleteTask(ctx, token, taskGUID)
			case tokenTypeUser:
				err = state.SDK.DeleteTaskWithUserToken(ctx, token, taskGUID)
			default:
				return fmt.Errorf("unsupported token type %s", tokenType)
			}
			if err != nil {
				return err
			}
			payload := map[string]any{"task_guid": taskGUID, "deleted": true}
			text := tableTextRow([]string{"task_guid", "deleted"}, []string{taskGUID, "true"})
			return state.Printer.Print(payload, text)
		},
	}
	return cmd
}

func newTaskListCmd(state *appState) *cobra.Command {
	var limit int
	var pageSize int
	var completed bool
	var taskType string
	var userIDType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks (my tasks)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			if limit <= 0 {
				return errors.New("limit must be greater than 0")
			}
			if pageSize <= 0 {
				return errors.New("page-size must be greater than 0")
			}

			token, err := tokenFor(context.Background(), state, tokenTypesUser)
			if err != nil {
				return err
			}
			var completedPtr *bool
			if cmd.Flags().Changed("completed") {
				v := completed
				completedPtr = &v
			}

			items := make([]larksdk.Task, 0, limit)
			pageToken := ""
			remaining := limit
			for {
				size := pageSize
				if size > maxTasksPageSize {
					size = maxTasksPageSize
				}
				if size > remaining {
					size = remaining
				}
				result, err := state.SDK.ListTasks(context.Background(), token, larksdk.ListTasksRequest{
					PageSize:   size,
					PageToken:  pageToken,
					Completed:  completedPtr,
					Type:       strings.TrimSpace(taskType),
					UserIDType: strings.TrimSpace(userIDType),
				})
				if err != nil {
					return withUserScopeHintForCommand(state, err)
				}
				items = append(items, result.Items...)
				if len(items) >= limit || !result.HasMore {
					break
				}
				remaining = limit - len(items)
				pageToken = result.PageToken
				if strings.TrimSpace(pageToken) == "" {
					break
				}
			}
			if len(items) > limit {
				items = items[:limit]
			}

			payload := map[string]any{"tasks": items}
			lines := make([]string, 0, len(items))
			for _, task := range items {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", task.GUID, task.Summary, task.Status, formatTaskTime(task.Due), taskAssignees(task.Members)))
			}
			text := tableText([]string{"task_guid", "summary", "status", "due", "assignees"}, lines, "no tasks found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 50, "max number of tasks to return")
	cmd.Flags().IntVar(&pageSize, "page-size", 50, "page size per request")
	cmd.Flags().BoolVar(&completed, "completed", false, "filter by completion status (true/false)")
	cmd.Flags().StringVar(&taskType, "type", "my_tasks", "task list type (default: my_tasks)")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "open_id", "user ID type (open_id, union_id, user_id)")
	return cmd
}

func buildTaskMembers(memberType string, assignees, followers []string, membersJSON string) ([]larksdk.TaskMember, error) {
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
	members := make([]larksdk.TaskMember, 0, len(assignees)+len(followers))
	for _, id := range assignees {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		members = append(members, larksdk.TaskMember{ID: id, Type: memberType, Role: "assignee"})
	}
	for _, id := range followers {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		members = append(members, larksdk.TaskMember{ID: id, Type: memberType, Role: "follower"})
	}
	return members, nil
}

func parseTasklistRefs(entries []string) ([]larksdk.TaskInTasklistInfo, error) {
	if len(entries) == 0 {
		return nil, nil
	}
	infos := make([]larksdk.TaskInTasklistInfo, 0, len(entries))
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.Split(entry, ":")
		if len(parts) > 2 {
			return nil, fmt.Errorf("invalid tasklist reference %q (expected <tasklist-guid>[:section-guid])", entry)
		}
		tasklistGUID := strings.TrimSpace(parts[0])
		if tasklistGUID == "" {
			return nil, fmt.Errorf("invalid tasklist reference %q (missing tasklist guid)", entry)
		}
		info := larksdk.TaskInTasklistInfo{TasklistGUID: tasklistGUID}
		if len(parts) == 2 {
			info.SectionGUID = strings.TrimSpace(parts[1])
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func buildTaskTime(cmd *cobra.Command, raw, flagName string, allDay bool) (*larksdk.TaskTime, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if cmd != nil && cmd.Flags().Changed(flagName) {
			return nil, fmt.Errorf("--%s requires a timestamp", flagName)
		}
		return nil, nil
	}
	ms, err := parseTaskTimestamp(raw)
	if err != nil {
		return nil, err
	}
	taskTime := &larksdk.TaskTime{Timestamp: fmt.Sprintf("%d", ms)}
	if cmd != nil && cmd.Flags().Changed(flagName) {
		v := allDay
		taskTime.IsAllDay = &v
	}
	return taskTime, nil
}

func parseTaskTimestamp(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, errors.New("timestamp is required")
	}
	if rawInt, err := strconv.ParseInt(raw, 10, 64); err == nil {
		if rawInt < 0 {
			return 0, errors.New("timestamp must be non-negative")
		}
		if len(raw) <= 10 {
			return rawInt * 1000, nil
		}
		return rawInt, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t.UnixMilli(), nil
	}
	if t, err := time.Parse("2006-01-02", raw); err == nil {
		return t.UnixMilli(), nil
	}
	return 0, fmt.Errorf("invalid timestamp: %s", raw)
}

func parseTaskCompletedAt(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("completed-at is required")
	}
	if raw == "0" {
		return "0", nil
	}
	ms, err := parseTaskTimestamp(raw)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", ms), nil
}

func formatTaskTime(taskTime *larksdk.TaskTime) string {
	if taskTime == nil || taskTime.Timestamp == "" {
		return ""
	}
	ms, err := strconv.ParseInt(taskTime.Timestamp, 10, 64)
	if err != nil {
		return taskTime.Timestamp
	}
	parsed := time.Unix(0, ms*int64(time.Millisecond)).UTC()
	if taskTime.IsAllDay != nil && *taskTime.IsAllDay {
		return parsed.Format("2006-01-02")
	}
	return parsed.Format(time.RFC3339)
}

func taskAssignees(members []larksdk.TaskMember) string {
	return taskMembersByRole(members, "assignee")
}

func taskFollowers(members []larksdk.TaskMember) string {
	return taskMembersByRole(members, "follower")
}

func taskMembersByRole(members []larksdk.TaskMember, role string) string {
	if len(members) == 0 {
		return ""
	}
	ids := make([]string, 0, len(members))
	for _, member := range members {
		if member.Role != role {
			continue
		}
		id := strings.TrimSpace(member.ID)
		if id == "" {
			continue
		}
		ids = append(ids, id)
	}
	return strings.Join(ids, ",")
}

func taskDetailRows(task larksdk.Task) [][]string {
	rows := make([][]string, 0, 12)
	add := func(name, value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		rows = append(rows, []string{name, value})
	}
	add("guid", task.GUID)
	add("task_id", task.TaskID)
	add("summary", task.Summary)
	add("status", task.Status)
	add("description", task.Description)
	add("due", formatTaskTime(task.Due))
	add("start", formatTaskTime(task.Start))
	add("completed_at", task.CompletedAt)
	add("repeat_rule", task.RepeatRule)
	add("assignees", taskAssignees(task.Members))
	add("followers", taskFollowers(task.Members))
	add("url", task.URL)
	add("created_at", task.CreatedAt)
	add("updated_at", task.UpdatedAt)
	return rows
}
