package main

import (
	"context"
	"errors"
	"fmt"

	"lark/internal/larksdk"

	"github.com/spf13/cobra"
)

func newBaseAppUpdateCmd(state *appState) *cobra.Command {
	var appToken string
	var name string
	var isAdvanced bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a Bitable app",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenant)
			if err != nil {
				return err
			}
			advancedSet := cmd.Flags().Changed("is-advanced")
			if name == "" && !advancedSet {
				return errors.New("one of --name or --is-advanced is required")
			}
			var isAdvancedPtr *bool
			if advancedSet {
				isAdvancedPtr = &isAdvanced
			}
			opts := larksdk.BitableAppUpdateOptions{
				Name:       name,
				IsAdvanced: isAdvancedPtr,
			}
			app, err := state.SDK.UpdateBitableApp(context.Background(), token, appToken, opts)
			if err != nil {
				return err
			}
			payload := map[string]any{"app": app}
			text := fmt.Sprintf("%s\t%s", app.AppToken, app.Name)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&name, "name", "", "Bitable app name")
	cmd.Flags().BoolVar(&isAdvanced, "is-advanced", false, "Enable or disable advanced permissions")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
