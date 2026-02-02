package main

import (
	"context"
	"fmt"

	"lark/internal/larksdk"

	"github.com/spf13/cobra"
)

func newBaseAppCopyCmd(state *appState) *cobra.Command {
	var appToken string
	var name string
	var folderToken string
	var withoutContent bool
	var timeZone string

	cmd := &cobra.Command{
		Use:   "copy",
		Short: "Copy a Bitable app",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithToken(cmd, state, nil, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				var withoutContentPtr *bool
				if cmd.Flags().Changed("without-content") {
					withoutContentPtr = &withoutContent
				}
				opts := larksdk.BitableAppCopyOptions{
					Name:           name,
					FolderToken:    folderToken,
					WithoutContent: withoutContentPtr,
					TimeZone:       timeZone,
				}
				app, err := sdk.CopyBitableApp(ctx, token, appToken, opts)
				if err != nil {
					return nil, "", err
				}
				payload := map[string]any{"app": app}
				text := fmt.Sprintf("%s\t%s", app.AppToken, app.Name)
				return payload, text, nil
			})
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&name, "name", "", "Bitable app name")
	cmd.Flags().StringVar(&folderToken, "folder-token", "", "Drive folder token (optional)")
	cmd.Flags().BoolVar(&withoutContent, "without-content", false, "Copy structure only (no data)")
	cmd.Flags().StringVar(&timeZone, "time-zone", "", "Document time zone (optional)")
	_ = cmd.MarkFlagRequired("app-token")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
