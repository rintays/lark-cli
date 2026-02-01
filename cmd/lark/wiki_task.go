package main

import (
	"context"
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
			taskID := strings.TrimSpace(args[0])
			return runWithToken(cmd, state, tokenTypesTenantOrUser, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				var result larksdk.WikiTaskResult
				var err error
				req := larksdk.GetWikiTaskRequest{
					TaskID:   strings.TrimSpace(taskID),
					TaskType: strings.TrimSpace(taskType),
				}
				switch tokenType {
				case tokenTypeTenant:
					result, err = sdk.GetWikiTaskV2(ctx, token, req)
				case tokenTypeUser:
					result, err = sdk.GetWikiTaskV2WithUserToken(ctx, token, req)
				default:
					return nil, "", fmt.Errorf("unsupported token type %s", tokenType)
				}
				if err != nil {
					return nil, "", err
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
				return payload, text, nil
			})
		},
	}

	cmd.Flags().StringVar(&taskType, "task-type", "move", "task type (default: move)")
	return cmd
}
