package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type tokenType string

type tokenOverride struct {
	Token string
	Type  tokenType
}

const (
	tokenTypeAuto   tokenType = "auto"
	tokenTypeTenant tokenType = "tenant"
	tokenTypeUser   tokenType = "user"
)

var (
	tokenTypesTenant       = []tokenType{tokenTypeTenant}
	tokenTypesUser         = []tokenType{tokenTypeUser}
	tokenTypesTenantOrUser = []tokenType{tokenTypeTenant, tokenTypeUser}
)

func parseTokenType(raw string) (tokenType, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return tokenTypeAuto, nil
	}
	switch tokenType(value) {
	case tokenTypeAuto, tokenTypeTenant, tokenTypeUser:
		return tokenType(value), nil
	default:
		return "", fmt.Errorf("unknown token type %q (expected auto, tenant, or user)", raw)
	}
}

func normalizeDefaultTokenType(raw string) tokenType {
	parsed, err := parseTokenType(raw)
	if err != nil {
		return tokenTypeTenant
	}
	if parsed == tokenTypeUser {
		return tokenTypeUser
	}
	return tokenTypeTenant
}

func resolveAccessToken(ctx context.Context, state *appState, allowed []tokenType, override *tokenOverride) (string, tokenType, error) {
	if state == nil {
		return "", "", errors.New("state is required")
	}
	if state.Config == nil {
		return "", "", errors.New("config is required")
	}
	allowedSet := map[tokenType]struct{}{}
	for _, t := range allowed {
		allowedSet[t] = struct{}{}
	}
	if len(allowedSet) == 0 {
		return "", "", errors.New("token policy is required")
	}
	allowedLabel := tokenAllowedLabel(allowedSet)

	requested, err := parseTokenType(state.TokenType)
	if err != nil {
		return "", "", err
	}

	if override != nil && override.Token != "" {
		if requested != tokenTypeAuto && requested != override.Type {
			return "", requested, fmt.Errorf("token type %s conflicts with provided %s token", requested, override.Type)
		}
		if _, ok := allowedSet[override.Type]; !ok {
			return "", override.Type, fmt.Errorf("token type %s not supported; supported: %s", override.Type, allowedLabel)
		}
		return override.Token, override.Type, nil
	}

	chosen := requested
	if len(allowedSet) == 1 {
		for t := range allowedSet {
			chosen = t
			break
		}
		if requested != tokenTypeAuto && requested != chosen {
			return "", requested, fmt.Errorf("token type %s not supported; supported: %s", requested, allowedLabel)
		}
	} else {
		if chosen == tokenTypeAuto {
			chosen = normalizeDefaultTokenType(state.Config.DefaultTokenType)
		}
		if _, ok := allowedSet[chosen]; !ok {
			return "", chosen, fmt.Errorf("token type %s not supported; supported: %s", chosen, allowedLabel)
		}
	}

	switch chosen {
	case tokenTypeTenant:
		token, err := ensureTenantToken(ctx, state)
		if err != nil {
			return "", chosen, err
		}
		return token, chosen, nil
	case tokenTypeUser:
		token, err := ensureUserToken(ctx, state)
		if err != nil {
			return "", chosen, err
		}
		if err := preflightUserScopes(state); err != nil {
			return "", chosen, err
		}
		if token == "" {
			return "", chosen, errors.New("user access token missing; set user_access_token in config or run `lark auth user login`")
		}
		return token, chosen, nil
	default:
		return "", chosen, fmt.Errorf("invalid token type: %s", chosen)
	}
}

func tokenAllowedLabel(allowed map[tokenType]struct{}) string {
	parts := make([]string, 0, len(allowed))
	if _, ok := allowed[tokenTypeTenant]; ok {
		parts = append(parts, string(tokenTypeTenant))
	}
	if _, ok := allowed[tokenTypeUser]; ok {
		parts = append(parts, string(tokenTypeUser))
	}
	return strings.Join(parts, ", ")
}

func tokenFor(ctx context.Context, state *appState, allowed []tokenType) (string, error) {
	token, _, err := resolveAccessToken(ctx, state, allowed, nil)
	return token, err
}

func tokenForOverride(ctx context.Context, state *appState, allowed []tokenType, override tokenOverride) (string, error) {
	token, _, err := resolveAccessToken(ctx, state, allowed, &override)
	return token, err
}
