package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"lark/internal/config"
)

const (
	userOAuthListenAddr     = "localhost:17653"
	userOAuthCallbackPath   = "/oauth/callback"
	userOAuthRedirectURL    = "http://localhost:17653/oauth/callback"
	defaultUserOAuthScope   = "offline_access"
	userOAuthReloginCommand = "lark auth user login --scopes \"offline_access\" --force-consent"
)

type userOAuthToken struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type oauthCallbackResult struct {
	code string
	err  error
}

func newAuthUserCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage user OAuth credentials",
	}
	cmd.AddCommand(newAuthUserLoginCmd(state))
	cmd.AddCommand(newAuthUserStatusCmd(state))
	cmd.AddCommand(newAuthUserScopesCmd(state))
	cmd.AddCommand(newAuthUserServicesCmd(state))
	return cmd
}

func newAuthUserLoginCmd(state *appState) *cobra.Command {
	var scopes string
	var legacyScope string
	var services []string
	var readonly bool
	var driveScope string
	var forceConsent bool
	var incremental bool
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in with user OAuth and store tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireCredentials(state.Config); err != nil {
				return err
			}
			scopeSet := cmd.Flags().Changed("scopes")
			legacyScopeSet := cmd.Flags().Changed("scope")
			if scopeSet && legacyScopeSet {
				return errors.New("--scope and --scopes cannot be used together")
			}
			if legacyScopeSet {
				scopes = legacyScope
				scopeSet = true
			}
			if scopeSet {
				if cmd.Flags().Changed("services") || cmd.Flags().Changed("readonly") || cmd.Flags().Changed("drive-scope") {
					return errors.New("--scopes cannot be combined with --services, --readonly, or --drive-scope")
				}
			}

			scopeOpts := userOAuthScopeOptions{
				Scopes:        scopes,
				ScopesSet:     scopeSet,
				Services:      parseServicesList(services),
				ServicesSet:   cmd.Flags().Changed("services"),
				Readonly:      readonly,
				DriveScope:    driveScope,
				DriveScopeSet: cmd.Flags().Changed("drive-scope"),
			}
			scopeList, _, err := resolveUserOAuthScopes(state, scopeOpts)
			if err != nil {
				return err
			}
			requestedScopes := requestedUserOAuthScopes(scopeList, state.Config.UserAccessTokenScope, incremental)
			scopeValue := strings.Join(requestedScopes, " ")

			authState, err := newOAuthState()
			if err != nil {
				return err
			}
			authorizeURL, err := buildUserAuthorizeURL(state.Config.BaseURL, state.Config.AppID, userOAuthRedirectURL, authState, scopeValue, userOAuthPrompt(forceConsent), incremental)
			if err != nil {
				return err
			}
			server := &http.Server{Addr: userOAuthListenAddr}
			resultCh := make(chan oauthCallbackResult, 1)
			server.Handler = oauthCallbackHandler(authState, resultCh, server)
			listener, err := net.Listen("tcp", userOAuthListenAddr)
			if err != nil {
				return fmt.Errorf("listen on %s: %w", userOAuthListenAddr, err)
			}
			defer func() {
				_ = server.Shutdown(context.Background())
			}()
			go func() {
				_ = server.Serve(listener)
			}()

			if err := openBrowser(authorizeURL); err != nil {
				fmt.Fprintf(state.Printer.Writer, "Open this URL in your browser:\n%s\n", authorizeURL)
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()
			var result oauthCallbackResult
			select {
			case result = <-resultCh:
				if result.err != nil {
					return result.err
				}
			case <-ctx.Done():
				return fmt.Errorf("timed out waiting for OAuth callback")
			}

			tokens, err := exchangeUserAccessToken(ctx, nil, state.Config.BaseURL, state.Config.AppID, state.Config.AppSecret, result.code, userOAuthRedirectURL)
			if err != nil {
				return err
			}
			if err := requireUserRefreshToken(tokens.RefreshToken); err != nil {
				return err
			}
			grantedScope := canonicalScopeString(tokens.Scope)
			warning := ""
			if grantedScope != "" {
				warning = scopesChangedWarning(state.Config.UserAccessTokenScope, grantedScope)
				state.Config.UserAccessTokenScope = grantedScope
			}
			if strings.EqualFold(state.Config.KeyringBackend, "keychain") {
				return errors.New("keychain backend is not implemented yet; please use keyring_backend=file (or set LARK_KEYRING_BACKEND=file)")
			}
			now := time.Now()
			state.Config.UserAccessToken = tokens.AccessToken
			state.Config.RefreshToken = tokens.RefreshToken
			state.Config.UserAccessTokenExpiresAt = now.Add(time.Duration(tokens.ExpiresIn) * time.Second).Unix()
			state.Config.UserScopes = scopeList
			refreshPayload := &config.UserRefreshTokenPayload{
				RefreshToken: tokens.RefreshToken,
				CreatedAt:    now.Unix(),
			}
			if scopeOpts.ServicesSet {
				refreshPayload.Services = scopeOpts.Services
			}
			if grantedScope != "" {
				refreshPayload.Scopes = grantedScope
			}
			state.Config.UserRefreshTokenPayload = refreshPayload
			if err := state.saveConfig(); err != nil {
				return err
			}

			payload := map[string]any{
				"config_path":                  state.ConfigPath,
				"user_access_token_expires_at": state.Config.UserAccessTokenExpiresAt,
			}
			if grantedScope != "" {
				payload["user_access_token_scope"] = grantedScope
			}
			if warning != "" {
				payload["warning"] = warning
			}

			message := fmt.Sprintf("saved user OAuth tokens to %s", state.ConfigPath)
			if warning != "" {
				message = fmt.Sprintf("%s\n\nWARNING: %s", message, warning)
			}
			return state.Printer.Print(payload, message)
		},
	}

	cmd.Flags().StringVar(&legacyScope, "scope", "", "OAuth scopes (space-separated) (deprecated: use --scopes)")
	cmd.Flags().StringVar(&scopes, "scopes", "", "OAuth scopes (space/comma-separated)")
	cmd.Flags().StringSliceVar(&services, "services", nil, "OAuth service set (comma-separated, use `lark auth user services`)")
	cmd.Flags().BoolVar(&readonly, "readonly", false, "use read-only OAuth scopes for selected services")
	cmd.Flags().StringVar(&driveScope, "drive-scope", "", "drive scope (full or readonly)")
	cmd.Flags().BoolVar(&forceConsent, "force-consent", false, "force the consent screen during OAuth")
	cmd.Flags().BoolVar(&incremental, "incremental", true, "use incremental OAuth (include granted scopes; set --incremental=false to request full scopes)")
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Minute, "timeout waiting for OAuth callback")

	return cmd
}

func newOAuthState() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func canonicalScopeString(scope string) string {
	fields := strings.Fields(scope)
	if len(fields) == 0 {
		return ""
	}
	sort.Strings(fields)
	out := fields[:0]
	for _, s := range fields {
		if len(out) == 0 || s != out[len(out)-1] {
			out = append(out, s)
		}
	}
	return strings.Join(out, " ")
}

func scopesChangedWarning(previousScope, newScope string) string {
	prev := canonicalScopeString(previousScope)
	next := canonicalScopeString(newScope)
	if prev == "" || next == "" || prev == next {
		return ""
	}
	return fmt.Sprintf(
		"OAuth scopes changed since last login.\nPrevious scopes: %s\nNew scopes: %s\n\nIf you intended to change scopes, you may need to force the consent screen so Lark re-grants scopes and issues an updated refresh token.\nRe-run: `%s`",
		prev,
		next,
		userOAuthReloginCommand,
	)
}

func requireUserRefreshToken(refreshToken string) error {
	if refreshToken != "" {
		return nil
	}
	return fmt.Errorf("offline access was not granted: refresh_token missing from OAuth response; re-run with: `%s`; check console redirect URL/config", userOAuthReloginCommand)
}

func userOAuthPrompt(forceConsent bool) string {
	if forceConsent {
		return "consent"
	}
	return ""
}
func buildUserAuthorizeURL(baseURL, appID, redirectURI, state, scope, prompt string, includeGrantedScopes bool) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	base.Path = "/open-apis/authen/v1/authorize"
	query := base.Query()
	query.Set("client_id", appID)
	query.Set("response_type", "code")
	query.Set("redirect_uri", redirectURI)
	if state != "" {
		query.Set("state", state)
	}
	if scope != "" {
		query.Set("scope", scope)
	}
	if prompt != "" {
		query.Set("prompt", prompt)
	}
	if includeGrantedScopes {
		query.Set("include_granted_scopes", "true")
	}
	base.RawQuery = query.Encode()
	return base.String(), nil
}

func exchangeUserAccessToken(ctx context.Context, httpClient *http.Client, baseURL, appID, appSecret, code, redirectURI string) (userOAuthToken, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	endpoint, err := buildUserTokenURL(baseURL)
	if err != nil {
		return userOAuthToken{}, err
	}
	payload := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     appID,
		"client_secret": appSecret,
		"code":          code,
		"redirect_uri":  redirectURI,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return userOAuthToken{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return userOAuthToken{}, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := httpClient.Do(req)
	if err != nil {
		return userOAuthToken{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return userOAuthToken{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return userOAuthToken{}, fmt.Errorf("token exchange failed: %s", strings.TrimSpace(string(data)))
	}
	var parsed userOAuthToken
	if err := json.Unmarshal(data, &parsed); err != nil {
		return userOAuthToken{}, err
	}
	if parsed.Error != "" {
		return userOAuthToken{}, fmt.Errorf("token exchange failed: %s", oauthErrorMessage(parsed))
	}
	if parsed.AccessToken == "" {
		return userOAuthToken{}, errors.New("token exchange failed: missing access_token")
	}
	return parsed, nil
}

func buildUserTokenURL(baseURL string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	base.Path = "/open-apis/authen/v2/oauth/token"
	return base.String(), nil
}

func oauthCallbackHandler(state string, resultCh chan<- oauthCallbackResult, server *http.Server) http.Handler {
	mux := http.NewServeMux()
	var once sync.Once
	mux.HandleFunc(userOAuthCallbackPath, func(w http.ResponseWriter, r *http.Request) {
		result := oauthCallbackResult{}
		query := r.URL.Query()
		if state != "" && query.Get("state") != state {
			result.err = errors.New("oauth state mismatch")
			writeOAuthError(w, "OAuth state mismatch")
			finalizeOAuthResult(server, &once, resultCh, result)
			return
		}
		if errValue := query.Get("error"); errValue != "" {
			message := errValue
			if desc := query.Get("error_description"); desc != "" {
				message = fmt.Sprintf("%s: %s", errValue, desc)
			}
			result.err = fmt.Errorf("oauth error: %s", message)
			writeOAuthError(w, message)
			finalizeOAuthResult(server, &once, resultCh, result)
			return
		}
		code := query.Get("code")
		if code == "" {
			result.err = errors.New("oauth callback missing code")
			writeOAuthError(w, "Missing authorization code")
			finalizeOAuthResult(server, &once, resultCh, result)
			return
		}
		result.code = code
		writeOAuthSuccess(w)
		finalizeOAuthResult(server, &once, resultCh, result)
	})
	return mux
}

func finalizeOAuthResult(server *http.Server, once *sync.Once, resultCh chan<- oauthCallbackResult, result oauthCallbackResult) {
	once.Do(func() {
		select {
		case resultCh <- result:
		default:
		}
		go func() {
			_ = server.Shutdown(context.Background())
		}()
	})
}

func writeOAuthSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = io.WriteString(w, "Login complete. You can close this window.")
}

func writeOAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = io.WriteString(w, fmt.Sprintf("Login failed: %s", message))
}

func oauthErrorMessage(token userOAuthToken) string {
	if token.ErrorDescription != "" {
		return token.ErrorDescription
	}
	return token.Error
}

func openBrowser(targetURL string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", targetURL).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL).Start()
	default:
		return exec.Command("xdg-open", targetURL).Start()
	}
}
