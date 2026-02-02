package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newBaseAppGetCmd(state *appState) *cobra.Command {
	var appToken string

	cmd := &cobra.Command{
		Use:     "info",
		Aliases: []string{"get"},
		Short:   "Get a Bitable app",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if len(args) == 0 {
				if strings.TrimSpace(appToken) == "" {
					return errors.New("app-token is required")
				}
				return nil
			}
			if appToken != "" && appToken != args[0] {
				return errors.New("app-token provided twice")
			}
			appToken = strings.TrimSpace(args[0])
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithToken(cmd, state, nil, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				app, err := sdk.GetBitableApp(ctx, token, appToken)
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
	return cmd
}
