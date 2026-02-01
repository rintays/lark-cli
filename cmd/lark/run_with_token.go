package main

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

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
