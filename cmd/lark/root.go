package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"lark/internal/config"
	"lark/internal/larksdk"
	"lark/internal/output"
)

type appState struct {
	ConfigPath     string
	Config         *config.Config
	Profile        string
	JSON           bool
	Verbose        bool
	TokenType      string
	UserAccount    string
	Printer        output.Printer
	SDK            *larksdk.Client
	Platform       string
	BaseURL        string
	baseURLPersist string

	// Command is the invoked command path (space-separated, excluding the root
	// binary name). Example: "mail send".
	Command string
}

func newRootCmd() *cobra.Command {
	state := &appState{}
	cmd := &cobra.Command{
		Use:           "lark",
		Short:         "A Go CLI for Feishu/Lark inspired by gog",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			state.Command = canonicalCommandPath(cmd)
			if state.Profile == "" {
				state.Profile = strings.TrimSpace(os.Getenv("LARK_PROFILE"))
			}
			state.Profile = strings.TrimSpace(state.Profile)
			if state.Profile == "" || strings.EqualFold(state.Profile, "default") {
				state.Profile = "default"
			}
			if state.UserAccount == "" {
				state.UserAccount = strings.TrimSpace(os.Getenv("LARK_ACCOUNT"))
			}
			if state.ConfigPath == "" {
				path, err := config.DefaultPathForProfile(state.Profile)
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
			out := cmd.OutOrStdout()
			state.Printer = output.Printer{
				Writer: out,
				JSON:   state.JSON,
				Styled: output.AutoStyle(out) && !state.JSON,
			}
			sdkClient, err := larksdk.New(cfg)
			if err == nil {
				state.SDK = sdkClient
			} else if state.Verbose {
				fmt.Fprintf(state.Printer.Writer, "SDK disabled: %v\n", err)
			}
			return nil
		},
	}
	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		if errors.Is(err, pflag.ErrHelp) {
			return err
		}
		usage := ""
		if cmd != nil {
			usage = cmd.UsageString()
		}
		return usageErrorWithUsage(cmd, err.Error(), flagErrorHint(cmd, err), usage)
	})

	cmd.PersistentFlags().StringVar(&state.ConfigPath, "config", "", "config path (default: ~/.config/lark/config.json; uses profile path when --profile or LARK_PROFILE is set)")
	cmd.PersistentFlags().StringVar(&state.Profile, "profile", "", "config profile (env: LARK_PROFILE)")
	cmd.PersistentFlags().BoolVar(&state.JSON, "json", false, "output JSON")
	cmd.PersistentFlags().BoolVar(&state.Verbose, "verbose", false, "verbose output")
	cmd.PersistentFlags().StringVar(&state.TokenType, "token-type", "auto", "access token type (auto|tenant|user)")
	cmd.PersistentFlags().StringVar(&state.UserAccount, "account", "", "user account label (default: config default or LARK_ACCOUNT)")
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

func canonicalCommandPath(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	path := strings.TrimSpace(cmd.CommandPath())
	root := cmd.Root()
	if root != nil {
		name := strings.TrimSpace(root.Name())
		if name != "" {
			if path == name {
				return ""
			}
			prefix := name + " "
			if strings.HasPrefix(path, prefix) {
				return strings.TrimSpace(strings.TrimPrefix(path, name))
			}
		}
	}
	if strings.HasPrefix(path, "lark ") {
		return strings.TrimSpace(strings.TrimPrefix(path, "lark "))
	}
	return path
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

func cachedUserTokenValid(token userToken, now time.Time) bool {
	if token.AccessToken == "" || token.ExpiresAt == 0 {
		return false
	}
	return token.ExpiresAt > now.Add(60*time.Second).Unix()
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
	account := resolveUserAccountName(state)
	stored, ok, err := loadUserToken(state, account)
	if err != nil {
		return "", err
	}
	if ok && cachedUserTokenValid(stored, time.Now()) {
		return stored.AccessToken, nil
	}
	acct, _ := loadUserAccount(state.Config, account)
	refreshToken := stored.RefreshToken
	if refreshToken == "" {
		refreshToken = acct.RefreshTokenValue()
	}
	if refreshToken == "" {
		return "", expireUserToken(state, account, errors.New("refresh token missing"))
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
	token, newRefreshToken, expiresIn, err := sdk.RefreshUserAccessToken(ctx, refreshToken)
	if err != nil {
		return "", expireUserToken(state, account, err)
	}
	newToken := userToken{
		AccessToken:  token,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(expiresIn) * time.Second).Unix(),
	}
	if newRefreshToken != "" {
		newToken.RefreshToken = newRefreshToken
		if acct.UserRefreshTokenPayload != nil {
			if userTokenBackend(state.Config) == "file" {
				acct.UserRefreshTokenPayload.RefreshToken = newRefreshToken
			}
			acct.UserRefreshTokenPayload.CreatedAt = time.Now().Unix()
			saveUserAccount(state.Config, account, acct)
		}
	}
	if err := storeUserToken(state, account, newToken); err != nil {
		return "", err
	}
	if err := state.saveConfig(); err != nil {
		return "", err
	}
	return token, nil
}

func expireUserToken(state *appState, account string, cause error) error {
	if err := clearUserToken(state, account); err != nil {
		return err
	}
	saveErr := state.saveConfig()

	reloginCmd, note := userOAuthReloginRecommendation(state)
	suffix := ""
	if note != "" {
		suffix = "; " + note
	}

	accountNote := ""
	if strings.TrimSpace(account) != "" {
		accountNote = fmt.Sprintf(" for account %q", account)
	}
	base := fmt.Sprintf("user access token expired%s%s; run `%s`", accountNote, suffix, reloginCmd)
	var refreshErr *larksdk.RefreshAccessTokenError
	if errors.As(cause, &refreshErr) {
		msg := strings.ToLower(refreshErr.Msg)
		if strings.Contains(msg, "invalid") || strings.Contains(msg, "revok") || strings.Contains(msg, "expire") {
			base = fmt.Sprintf("user access token expired (refresh token revoked or expired)%s%s; run `%s`", accountNote, suffix, reloginCmd)
		}
	}
	if cause != nil {
		var refreshErr *larksdk.RefreshAccessTokenError
		if errors.As(cause, &refreshErr) {
			msg := strings.ToLower(refreshErr.Msg)
			mentionsRefreshToken := strings.Contains(msg, "refresh_token") || strings.Contains(msg, "refresh token")
			looksRevoked := strings.Contains(msg, "invalid") || strings.Contains(msg, "expired") || strings.Contains(msg, "revoked")
			if mentionsRefreshToken && looksRevoked {
				base = fmt.Sprintf("refresh token revoked or expired%s; cleared cached credentials%s; run `%s`", accountNote, suffix, reloginCmd)
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
	targetCmd := commandForArgs(cmd, os.Args[1:])
	if err := cmd.Execute(); err != nil {
		if targetCmd != nil && !isUsageError(err) && isRequiredFlagError(err) {
			err = usageErrorWithUsage(targetCmd, err.Error(), flagErrorHint(targetCmd, err), targetCmd.UsageString())
		}
		fmt.Fprintln(os.Stderr, output.FormatError(err, output.AutoStyle(os.Stderr)))
		return 1
	}
	return 0
}

func commandForArgs(cmd *cobra.Command, args []string) *cobra.Command {
	if cmd == nil {
		return nil
	}
	if len(args) == 0 {
		return cmd
	}
	found, _, err := cmd.Find(args)
	if err != nil {
		return cmd
	}
	if found == nil {
		return cmd
	}
	return found
}

func isUsageError(err error) bool {
	if err == nil {
		return false
	}
	var usageErr output.UsageError
	return errors.As(err, &usageErr)
}

func isRequiredFlagError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "required flag(s)")
}
