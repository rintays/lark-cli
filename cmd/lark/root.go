package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"lark/internal/config"
	"lark/internal/larksdk"
	"lark/internal/output"
)

type appState struct {
	ConfigPath     string
	Config         *config.Config
	JSON           bool
	Verbose        bool
	TokenType      string
	Printer        output.Printer
	SDK            *larksdk.Client
	Platform       string
	BaseURL        string
	baseURLPersist string
}

func newRootCmd() *cobra.Command {
	state := &appState{}
	cmd := &cobra.Command{
		Use:          "lark",
		Short:        "A Go CLI for Feishu/Lark inspired by gog",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if state.ConfigPath == "" {
				path, err := config.DefaultPath()
				if err != nil {
					return err
				}
				state.ConfigPath = path
			}
			cfg, err := config.Load(state.ConfigPath)
			if err != nil {
				return err
			}
			state.Config = cfg
			if err := applyBaseURLOverrides(state, cfg); err != nil {
				return err
			}
			state.Printer = output.Printer{Writer: cmd.OutOrStdout(), JSON: state.JSON}
			sdkClient, err := larksdk.New(cfg)
			if err == nil {
				state.SDK = sdkClient
			} else if state.Verbose {
				fmt.Fprintf(state.Printer.Writer, "SDK disabled: %v\n", err)
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&state.ConfigPath, "config", "", "config path (default: ~/.config/lark/config.json)")
	cmd.PersistentFlags().BoolVar(&state.JSON, "json", false, "output JSON")
	cmd.PersistentFlags().BoolVar(&state.Verbose, "verbose", false, "verbose output")
	cmd.PersistentFlags().StringVar(&state.TokenType, "token-type", "auto", "access token type (auto|tenant|user)")
	cmd.PersistentFlags().StringVar(&state.Platform, "platform", "", "platform (feishu|lark)")
	cmd.PersistentFlags().StringVar(&state.BaseURL, "base-url", "", "base URL override")

	cmd.AddCommand(newVersionCmd(state))
	cmd.AddCommand(newAuthCmd(state))
	cmd.AddCommand(newWhoamiCmd(state))
	cmd.AddCommand(newMsgCmd(state))
	cmd.AddCommand(newChatsCmd(state))
	cmd.AddCommand(newUsersCmd(state))
	cmd.AddCommand(newDriveCmd(state))
	cmd.AddCommand(newDocsCmd(state))
	cmd.AddCommand(newSheetsCmd(state))
	cmd.AddCommand(newCalendarCmd(state))
	cmd.AddCommand(newMeetingsCmd(state))
	cmd.AddCommand(newWikiCmd(state))
	cmd.AddCommand(newMinutesCmd(state))
	cmd.AddCommand(newContactsCmd(state))
	cmd.AddCommand(newMailCmd(state))
	cmd.AddCommand(newBaseCmd(state))
	cmd.AddCommand(newConfigCmd(state))

	return cmd
}

func applyBaseURLOverrides(state *appState, cfg *config.Config) error {
	state.baseURLPersist = cfg.BaseURL
	if state.BaseURL != "" {
		cfg.BaseURL = normalizeBaseURL(state.BaseURL)
		return nil
	}
	if state.Platform == "" {
		cfg.BaseURL = normalizeBaseURL(cfg.BaseURL)
		return nil
	}
	baseURL, err := platformBaseURL(state.Platform)
	if err != nil {
		return err
	}
	cfg.BaseURL = normalizeBaseURL(baseURL)
	return nil
}

func normalizeBaseURL(raw string) string {
	base := strings.TrimSpace(raw)
	base = strings.TrimRight(base, "/")
	base = strings.TrimSuffix(base, "/open-apis")
	base = strings.TrimSuffix(base, "/open-apis/")
	base = strings.TrimRight(base, "/")
	return base
}

func platformBaseURL(platform string) (string, error) {
	switch strings.ToLower(platform) {
	case "feishu":
		return "https://open.feishu.cn", nil
	case "lark":
		return "https://open.larksuite.com", nil
	default:
		return "", fmt.Errorf("unsupported platform %q (expected feishu or lark)", platform)
	}
}

func platformFromBaseURL(baseURL string) string {
	normalized := normalizeBaseURL(baseURL)
	switch {
	case strings.EqualFold(normalized, "https://open.feishu.cn"):
		return "feishu"
	case strings.EqualFold(normalized, "https://open.larkoffice.com"):
		return "lark"
	case strings.EqualFold(normalized, "https://open.larksuite.com"):
		return "lark"
	default:
		return "custom"
	}
}

func (state *appState) saveConfig() error {
	if state.Config == nil {
		return errors.New("config is required")
	}
	cfg := *state.Config
	// Runtime overrides (--base-url/--platform) must not mutate persisted config.
	// Always restore the originally loaded base URL (even if empty).
	if state.BaseURL != "" || state.Platform != "" {
		cfg.BaseURL = state.baseURLPersist
	}
	return config.Save(state.ConfigPath, &cfg)
}

func requireCredentials(cfg *config.Config) error {
	if cfg.AppID == "" || cfg.AppSecret == "" {
		return errors.New("app_id and app_secret must be set in config")
	}
	return nil
}

func cachedTokenValid(cfg *config.Config, now time.Time) bool {
	if cfg.TenantAccessToken == "" || cfg.TenantAccessTokenExpiresAt == 0 {
		return false
	}
	return cfg.TenantAccessTokenExpiresAt > now.Add(60*time.Second).Unix()
}

func cachedUserTokenValid(cfg *config.Config, now time.Time) bool {
	if cfg.UserAccessToken == "" || cfg.UserAccessTokenExpiresAt == 0 {
		return false
	}
	return cfg.UserAccessTokenExpiresAt > now.Add(60*time.Second).Unix()
}

func ensureTenantToken(ctx context.Context, state *appState) (string, error) {
	if err := requireCredentials(state.Config); err != nil {
		return "", err
	}
	if cachedTokenValid(state.Config, time.Now()) {
		return state.Config.TenantAccessToken, nil
	}
	if state.Verbose {
		fmt.Fprintln(state.Printer.Writer, "refreshing tenant access token")
	}
	sdk := state.SDK
	if sdk == nil {
		var err error
		sdk, err = larksdk.New(state.Config)
		if err != nil {
			return "", errors.New("auth client is required")
		}
		state.SDK = sdk
	}
	token, expiresIn, err := sdk.TenantAccessToken(ctx)
	if err != nil {
		return "", err
	}
	state.Config.TenantAccessToken = token
	state.Config.TenantAccessTokenExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second).Unix()
	if err := state.saveConfig(); err != nil {
		return "", err
	}
	return token, nil
}

func ensureUserToken(ctx context.Context, state *appState) (string, error) {
	if err := requireCredentials(state.Config); err != nil {
		return "", err
	}
	if cachedUserTokenValid(state.Config, time.Now()) {
		return state.Config.UserAccessToken, nil
	}
	if state.Config.RefreshToken == "" {
		return "", expireUserToken(state, errors.New("refresh token missing"))
	}
	if state.Verbose {
		fmt.Fprintln(state.Printer.Writer, "refreshing user access token")
	}
	sdk := state.SDK
	if sdk == nil {
		var err error
		sdk, err = larksdk.New(state.Config)
		if err != nil {
			return "", errors.New("auth client is required")
		}
		state.SDK = sdk
	}
	token, newRefreshToken, expiresIn, err := sdk.RefreshUserAccessToken(ctx, state.Config.RefreshToken)
	if err != nil {
		return "", expireUserToken(state, err)
	}
	state.Config.UserAccessToken = token
	state.Config.UserAccessTokenExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second).Unix()
	if newRefreshToken != "" {
		state.Config.RefreshToken = newRefreshToken
	}
	if err := state.saveConfig(); err != nil {
		return "", err
	}
	return token, nil
}

func expireUserToken(state *appState, cause error) error {
	state.Config.UserAccessToken = ""
	state.Config.UserAccessTokenExpiresAt = 0
	state.Config.RefreshToken = ""
	saveErr := state.saveConfig()
	base := fmt.Sprintf("user access token expired; run `%s`", userOAuthReloginCommand)
	var refreshErr *larksdk.RefreshAccessTokenError
	if errors.As(cause, &refreshErr) {
		msg := strings.ToLower(refreshErr.Msg)
		if strings.Contains(msg, "invalid") || strings.Contains(msg, "revok") || strings.Contains(msg, "expire") {
			base = fmt.Sprintf("user access token expired (refresh token revoked or expired); run `%s`", userOAuthReloginCommand)
		}
	}
	if cause != nil {
		var refreshErr *larksdk.RefreshAccessTokenError
		if errors.As(cause, &refreshErr) {
			msg := strings.ToLower(refreshErr.Msg)
			mentionsRefreshToken := strings.Contains(msg, "refresh_token") || strings.Contains(msg, "refresh token")
			looksRevoked := strings.Contains(msg, "invalid") || strings.Contains(msg, "expired") || strings.Contains(msg, "revoked")
			if mentionsRefreshToken && looksRevoked {
				base = fmt.Sprintf("refresh token revoked or expired; cleared cached credentials; run `%s`", userOAuthReloginCommand)
			}
		}
	}
	if saveErr != nil {
		if cause != nil {
			return fmt.Errorf("%s: %v; failed to clear cached token: %w", base, cause, saveErr)
		}
		return fmt.Errorf("%s: failed to clear cached token: %w", base, saveErr)
	}
	if cause != nil {
		return fmt.Errorf("%s: %w", base, cause)
	}
	return errors.New(base)
}

func execute() int {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
