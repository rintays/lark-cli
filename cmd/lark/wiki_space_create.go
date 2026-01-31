package main

import (
	"context"
	"errors"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newWikiSpaceCreateCmd(state *appState) *cobra.Command {
	var name string
	var description string
	var spaceType string
	var visibility string
	var openSharing string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Wiki space (v2)",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				if strings.TrimSpace(name) == "" {
					return errors.New("name is required")
				}
				return nil
			}
			if name != "" && name != args[0] {
				return errors.New("name provided twice")
			}
			return cmd.Flags().Set("name", args[0])
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			space, err := state.SDK.CreateWikiSpaceV2(context.Background(), token, larksdk.CreateWikiSpaceRequest{
				Name:        strings.TrimSpace(name),
				Description: strings.TrimSpace(description),
				SpaceType:   strings.TrimSpace(spaceType),
				Visibility:  strings.TrimSpace(visibility),
				OpenSharing: strings.TrimSpace(openSharing),
			})
			if err != nil {
				return err
			}
			payload := map[string]any{"space": space}
			text := tableTextRow(
				[]string{"space_id", "name", "space_type", "visibility"},
				[]string{space.SpaceID, space.Name, space.SpaceType, space.Visibility},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "space name (or provide as positional argument)")
	cmd.Flags().StringVar(&description, "description", "", "space description")
	cmd.Flags().StringVar(&spaceType, "space-type", "", "space type (team or personal)")
	cmd.Flags().StringVar(&visibility, "visibility", "", "space visibility (public or private)")
	cmd.Flags().StringVar(&openSharing, "open-sharing", "", "open sharing status (open or closed)")
	return cmd
}
