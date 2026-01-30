package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newBaseCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "base",
		Short: "Manage Bitable bases",
	}
	cmd.AddCommand(newBaseTableCmd(state))
	return cmd
}

func newBaseTableCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "table",
		Short: "Manage Bitable tables",
	}
	cmd.AddCommand(newBaseTableListCmd(state))
	return cmd
}

func newBaseTableListCmd(state *appState) *cobra.Command {
	var appToken string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Bitable tables",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			result, err := state.SDK.ListBaseTables(context.Background(), token, appToken)
			if err != nil {
				return err
			}
			tables := result.Items
			payload := map[string]any{"tables": tables}
			lines := make([]string, 0, len(tables))
			for _, table := range tables {
				lines = append(lines, fmt.Sprintf("%s\t%s", table.TableID, table.Name))
			}
			text := "no tables found"
			if len(lines) > 0 {
				text = strings.Join(lines, "\n")
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
