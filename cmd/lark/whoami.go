package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newWhoamiCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Show tenant or user information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithToken(cmd, state, tokenTypesTenantOrUser, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				switch tokenType {
				case tokenTypeTenant:
					info, err := sdk.WhoAmI(ctx, token)
					if err != nil {
						return nil, "", err
					}
					payload := map[string]any{
						"tenant_key": info.TenantKey,
						"name":       info.Name,
					}
					return payload, fmt.Sprintf("%s (%s)", info.Name, info.TenantKey), nil
				case tokenTypeUser:
					info, err := sdk.UserInfo(ctx, token)
					if err != nil {
						return nil, "", err
					}
					return info, formatWhoamiUserText(info), nil
				default:
					return nil, "", fmt.Errorf("token type %s not supported", tokenType)
				}
			})
		},
	}
	return cmd
}

func formatWhoamiUserText(info larksdk.UserInfo) string {
	name := strings.TrimSpace(info.Name)
	if name == "" {
		name = strings.TrimSpace(info.EnName)
	}
	if name == "" {
		name = firstNonEmpty(info.UserID, info.OpenID, info.UnionID, info.TenantKey, "user")
	}
	labels := make([]string, 0, 4)
	if info.UserID != "" {
		labels = append(labels, fmt.Sprintf("user_id: %s", info.UserID))
	}
	if info.OpenID != "" {
		labels = append(labels, fmt.Sprintf("open_id: %s", info.OpenID))
	}
	if info.UnionID != "" {
		labels = append(labels, fmt.Sprintf("union_id: %s", info.UnionID))
	}
	if info.TenantKey != "" {
		labels = append(labels, fmt.Sprintf("tenant_key: %s", info.TenantKey))
	}
	if len(labels) == 0 {
		return name
	}
	return fmt.Sprintf("%s (%s)", name, strings.Join(labels, ", "))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
