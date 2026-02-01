package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newWikiNodeAttachCmd(state *appState) *cobra.Command {
	var spaceID string
	var objType string
	var objToken string
	var parentNodeToken string
	var apply bool

	cmd := &cobra.Command{
		Use:     "attach [obj-type] [obj-token]",
		Short:   "Attach a Drive document to a Wiki space (v2)",
		Aliases: []string{"move-docs"},
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if len(args) > 0 {
				if strings.TrimSpace(objType) != "" && objType != args[0] {
					return errors.New("obj-type provided twice")
				}
				objType = args[0]
			}
			if len(args) > 1 {
				if strings.TrimSpace(objToken) != "" && objToken != args[1] {
					return errors.New("obj-token provided twice")
				}
				objToken = args[1]
			}
			if strings.TrimSpace(objType) == "" {
				return errors.New("obj-type is required")
			}
			if strings.TrimSpace(objToken) == "" {
				return errors.New("obj-token is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			req := larksdk.MoveDocsToWikiRequest{
				SpaceID:         strings.TrimSpace(spaceID),
				ParentNodeToken: strings.TrimSpace(parentNodeToken),
				ObjType:         strings.TrimSpace(objType),
				ObjToken:        strings.TrimSpace(objToken),
				Apply:           apply,
				ApplySet:        cmd.Flags().Changed("apply"),
			}
			return runWithToken(cmd, state, tokenTypesTenantOrUser, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				var result larksdk.MoveDocsToWikiResult
				var err error
				switch tokenType {
				case tokenTypeTenant:
					result, err = sdk.MoveDocsToWikiV2(ctx, token, req)
				case tokenTypeUser:
					result, err = sdk.MoveDocsToWikiV2WithUserToken(ctx, token, req)
				default:
					return nil, "", fmt.Errorf("unsupported token type %s", tokenType)
				}
				if err != nil {
					return nil, "", err
				}
				applied := ""
				if result.Applied != nil {
					applied = strconv.FormatBool(*result.Applied)
				}
				payload := map[string]any{"result": result}
				text := tableTextRow(
					[]string{"wiki_token", "task_id", "applied"},
					[]string{result.WikiToken, result.TaskID, applied},
				)
				return payload, text, nil
			})
		},
	}

	cmd.Flags().StringVar(&spaceID, "space-id", "", "Wiki space ID")
	cmd.Flags().StringVar(&objType, "obj-type", "", "object type (doc, docx, sheet, slides, bitable, mindnote, file)")
	cmd.Flags().StringVar(&objToken, "obj-token", "", "object token")
	cmd.Flags().StringVar(&parentNodeToken, "parent-node-token", "", "parent node token (optional)")
	cmd.Flags().BoolVar(&apply, "apply", false, "apply for permission if lacking access")
	_ = cmd.MarkFlagRequired("space-id")
	return cmd
}
