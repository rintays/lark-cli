package main

import (
    "context"
    "fmt"

    "github.com/spf13/cobra"
)

func newWhoamiCmd(state *appState) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "whoami",
        Short: "Show tenant information",
        RunE: func(cmd *cobra.Command, args []string) error {
            token, err := ensureTenantToken(context.Background(), state)
            if err != nil {
                return err
            }
            info, err := state.Client.WhoAmI(context.Background(), token)
            if err != nil {
                return err
            }
            payload := map[string]any{
                "tenant_key": info.TenantKey,
                "name":       info.Name,
            }
            return state.Printer.Print(payload, fmt.Sprintf("%s (%s)", info.Name, info.TenantKey))
        },
    }
    return cmd
}
