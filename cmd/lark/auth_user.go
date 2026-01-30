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
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"lark/internal/config"
)

const (
	userOAuthListenAddr   = "localhost:17653"
	userOAuthCallbackPath = "/oauth/callback"
	userOAuthRedirectURL  = "http://localhost:17653/oauth/callback"
	defaultUserOAuthScope = "offline_access"
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
	return cmd
}

func newAuthUserLoginCmd(state *appState) *cobra.Command {
	var scope string
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in with user OAuth and store tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireCredentials(state.Config); err != nil {
				return err
			}
			authState, err := newOAuthState()
			if err != nil {
				return err
			}
			authorizeURL, err := buildUserAuthorizeURL(state.Config.BaseURL, state.Config.AppID, userOAuthRedirectURL, authState, userOAuthScope(scope, cmd.Flags().Changed("scope")))
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
			state.Config.UserAccessToken = tokens.AccessToken
			state.Config.RefreshToken = tokens.RefreshToken
			state.Config.UserAccessTokenExpiresAt = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second).Unix()
			if err := config.Save(state.ConfigPath, state.Config); err != nil {
				return err
			}

			payload := map[string]any{
				"config_path":                  state.ConfigPath,
				"user_access_token_expires_at": state.Config.UserAccessTokenExpiresAt,
			}
			return state.Printer.Print(payload, fmt.Sprintf("saved user OAuth tokens to %s", state.ConfigPath))
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "OAuth scopes (space-separated)")
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

func userOAuthScope(scope string, scopeSet bool) string {
	if !scopeSet {
		return defaultUserOAuthScope
	}
	return scope
}

func requireUserRefreshToken(refreshToken string) error {
	if refreshToken != "" {
		return nil
	}
	return errors.New("offline access was not granted: refresh_token missing from OAuth response; re-run with: lark auth user login --scope offline_access; check console redirect URL/config")
}

func buildUserAuthorizeURL(baseURL, appID, redirectURI, state, scope string) (string, error) {
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
