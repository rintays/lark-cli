package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/authregistry"
	"lark/internal/larksdk"
)

type runWithTokenFunc func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error)

func runWithToken(cmd *cobra.Command, state *appState, allowed []tokenType, override *tokenOverride, fn runWithTokenFunc) error {
	if cmd == nil {
		return errors.New("command is required")
	}
	if fn == nil {
		return errors.New("token runner is required")
	}
	if _, err := requireSDK(state); err != nil {
		return err
	}
	if len(allowed) == 0 {
		if state.Command == "" && cmd != nil {
			command := canonicalCommandPath(cmd)
			if root := cmd.Root(); root != nil {
				rootName := strings.TrimSpace(root.Name())
				if rootName != "" && !strings.EqualFold(rootName, "lark") {
					command = strings.TrimSpace(cmd.CommandPath())
				}
			}
			if command == "" {
				command = strings.TrimSpace(cmd.CommandPath())
			}
			state.Command = command
		}
		resolved, err := allowedTokenTypesForCommand(state)
		if err != nil {
			return err
		}
		allowed = resolved
	}
	ctx := cmd.Context()
	token, tokenTypeValue, err := resolveAccessToken(ctx, state, allowed, override)
	if err != nil {
		return err
	}
	payload, text, err := fn(ctx, state.SDK, token, tokenTypeValue)
	if err != nil {
		if tokenTypeValue == tokenTypeUser {
			return withUserScopeHintForCommand(state, err)
		}
		return err
	}
	if payload == nil && text == "" {
		return nil
	}
	return state.Printer.Print(payload, text)
}

func allowedTokenTypesForCommand(state *appState) ([]tokenType, error) {
	if state == nil {
		return nil, errors.New("state is required")
	}
	command := strings.TrimSpace(state.Command)
	if command == "" {
		return nil, errors.New("command is required")
	}
	services, ok := authregistry.ServicesForCommand(command)
	if !ok {
		return nil, fmt.Errorf("no auth registry mapping found for command %q", command)
	}
	types, err := authregistry.TokenTypesFromServices(services)
	if err != nil {
		return nil, err
	}
	out := make([]tokenType, 0, len(types))
	for _, tt := range types {
		switch tt {
		case authregistry.TokenTenant:
			out = append(out, tokenTypeTenant)
		case authregistry.TokenUser:
			out = append(out, tokenTypeUser)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no token types declared for command %q", command)
	}
	return out, nil
}
