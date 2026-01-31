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
	cmd.AddCommand(newWikiTaskGetCmd(state))
	return cmd
}

func newWikiTaskGetCmd(state *appState) *cobra.Command {
	var taskID string
	var taskType string

	cmd := &cobra.Command{
		Use:     "get <task-id>",
		Aliases: []string{"list"},
		Short:   "Get Wiki task results (v2)",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				if strings.TrimSpace(taskID) == "" {
					return errors.New("task-id is required")
				}
				return nil
			}
			if taskID != "" && taskID != args[0] {
				return errors.New("task-id provided twice")
			}
			return cmd.Flags().Set("task-id", args[0])
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			tenantToken, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}

			result, err := state.SDK.GetWikiTaskV2(context.Background(), tenantToken, larksdk.GetWikiTaskRequest{
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
			text := "no task results"
			if len(lines) > 0 {
				text = strings.Join(lines, "\n")
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&taskID, "task-id", "", "wiki task id (or provide as positional argument)")
	cmd.Flags().StringVar(&taskType, "task-type", "move", "task type (default: move)")
	return cmd
}
