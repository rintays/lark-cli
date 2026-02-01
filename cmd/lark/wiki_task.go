package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newWikiTaskCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Query Wiki task results",
	}
	cmd.AddCommand(newWikiTaskInfoCmd(state))
	return cmd
}

func newWikiTaskInfoCmd(state *appState) *cobra.Command {
	var taskType string

	cmd := &cobra.Command{
		Use:   "info <task-id>",
		Short: "Show Wiki task results (v2)",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if strings.TrimSpace(args[0]) == "" {
				return errors.New("task-id is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := requireSDK(state); err != nil {
				return err
			}
			taskID := strings.TrimSpace(args[0])
			tenantToken, err := tokenFor(cmd.Context(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}

			result, err := state.SDK.GetWikiTaskV2(cmd.Context(), tenantToken, larksdk.GetWikiTaskRequest{
				TaskID:   strings.TrimSpace(taskID),
				TaskType: strings.TrimSpace(taskType),
			})
			if err != nil {
				return err
			}

			payload := map[string]any{"task": result}
			lines := make([]string, 0, len(result.MoveResult))
			for _, mr := range result.MoveResult {
				nodeToken := ""
				objType := ""
				title := ""
				objToken := ""
				if mr.Node != nil {
					nodeToken = mr.Node.NodeToken
					objType = mr.Node.ObjType
					title = mr.Node.Title
					objToken = mr.Node.ObjToken
				}
				lines = append(lines, fmt.Sprintf("%d\t%s\t%s\t%s\t%s\t%s", mr.Status, mr.StatusMsg, nodeToken, objType, title, objToken))
			}
			text := tableText(
				[]string{"status", "status_msg", "node_token", "obj_type", "title", "obj_token"},
				lines,
				"no task results",
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&taskType, "task-type", "move", "task type (default: move)")
	return cmd
}
