package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newWikiNodeUpdateTitleCmd(state *appState) *cobra.Command {
	var spaceID string
	var title string

	cmd := &cobra.Command{
		Use:   "update-title <node-token> <title>",
		Short: "Update a Wiki node title (v2)",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.RangeArgs(1, 2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if strings.TrimSpace(args[0]) == "" {
				return errors.New("node-token is required")
			}
			if len(args) == 2 {
				if strings.TrimSpace(title) != "" && title != args[1] {
					return errors.New("title provided twice")
				}
				title = args[1]
			}
			if strings.TrimSpace(title) == "" {
				return errors.New("title is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeToken := strings.TrimSpace(args[0])
			title = strings.TrimSpace(title)
			req := larksdk.UpdateWikiNodeTitleRequest{
				SpaceID:   strings.TrimSpace(spaceID),
				NodeToken: nodeToken,
				Title:     title,
			}
			return runWithToken(cmd, state, nil, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				var err error
				switch tokenType {
				case tokenTypeTenant:
					err = sdk.UpdateWikiNodeTitleV2(ctx, token, req)
				case tokenTypeUser:
					err = sdk.UpdateWikiNodeTitleV2WithUserToken(ctx, token, req)
				default:
					return nil, "", fmt.Errorf("unsupported token type %s", tokenType)
				}
				if err != nil {
					return nil, "", err
				}
				payload := map[string]any{"updated": true, "node_token": nodeToken, "title": title}
				text := tableTextRow([]string{"node_token", "title"}, []string{nodeToken, title})
				return payload, text, nil
			})
		},
	}

	cmd.Flags().StringVar(&spaceID, "space-id", "", "Wiki space ID")
	cmd.Flags().StringVar(&title, "title", "", "new node title")
	_ = cmd.MarkFlagRequired("space-id")
	return cmd
}
