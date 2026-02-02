package main

import (
	"context"
	"fmt"

	"lark/internal/larksdk"

	"github.com/spf13/cobra"
)

func newBaseAppCreateCmd(state *appState) *cobra.Command {
	var name string
	var folderToken string
	var timeZone string
	var customizedConfig bool
	var sourceAppToken string
	var copyTypes []string
	var apiType string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Bitable app",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithToken(cmd, state, nil, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				var customizedConfigPtr *bool
				if cmd.Flags().Changed("customized-config") {
					customizedConfigPtr = &customizedConfig
				}
				opts := larksdk.BitableAppCreateOptions{
					FolderToken:      folderToken,
					TimeZone:         timeZone,
					CustomizedConfig: customizedConfigPtr,
					SourceAppToken:   sourceAppToken,
					CopyTypes:        copyTypes,
					ApiType:          apiType,
				}
				app, err := sdk.CreateBitableApp(ctx, token, name, opts)
				if err != nil {
					return nil, "", err
				}
				payload := map[string]any{"app": app}
				text := fmt.Sprintf("%s\t%s", app.AppToken, app.Name)
				return payload, text, nil
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Bitable app name")
	cmd.Flags().StringVar(&folderToken, "folder-token", "", "Drive folder token (optional)")
	cmd.Flags().StringVar(&timeZone, "time-zone", "", "Document time zone (optional)")
	cmd.Flags().BoolVar(&customizedConfig, "customized-config", false, "Use customized config create flow")
	cmd.Flags().StringVar(&sourceAppToken, "source-app-token", "", "Source app token for customized config (optional)")
	cmd.Flags().StringSliceVar(&copyTypes, "copy-types", nil, "Customized config copy types (comma-separated)")
	cmd.Flags().StringVar(&apiType, "api-type", "", "API type override for legacy flows (optional)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
