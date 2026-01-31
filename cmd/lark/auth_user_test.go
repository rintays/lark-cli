package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"lark/internal/config"
)

func TestBuildUserAuthorizeURL(t *testing.T) {
	urlStr, err := buildUserAuthorizeURL("https://open.feishu.cn", "app-id", userOAuthRedirectURL, "state123", "offline_access", "", false)
	if err != nil {
		t.Fatalf("build authorize url: %v", err)
	}
	parsed, err := url.Parse(urlStr)
	if err != nil {
		t.Fatalf("parse authorize url: %v", err)
	}
	if parsed.Path != "/open-apis/authen/v1/authorize" {
		t.Fatalf("unexpected authorize path: %s", parsed.Path)
	}
	query := parsed.Query()
	if query.Get("client_id") != "app-id" {
		t.Fatalf("unexpected client_id: %s", query.Get("client_id"))
	}
	if query.Get("response_type") != "code" {
		t.Fatalf("unexpected response_type: %s", query.Get("response_type"))
	}
	if query.Get("redirect_uri") != userOAuthRedirectURL {
		t.Fatalf("unexpected redirect_uri: %s", query.Get("redirect_uri"))
	}
	if query.Get("state") != "state123" {
		t.Fatalf("unexpected state: %s", query.Get("state"))
	}
	if query.Get("scope") != "offline_access" {
		t.Fatalf("unexpected scope: %s", query.Get("scope"))
	}
	if query.Get("prompt") != "" {
		t.Fatalf("unexpected prompt: %s", query.Get("prompt"))
	}
	if query.Get("include_granted_scopes") != "" {
		t.Fatalf("unexpected include_granted_scopes: %s", query.Get("include_granted_scopes"))
	}
}

func TestBuildUserAuthorizeURLWithPrompt(t *testing.T) {
	urlStr, err := buildUserAuthorizeURL("https://open.feishu.cn", "app-id", userOAuthRedirectURL, "state123", "offline_access", "consent", false)
	if err != nil {
		t.Fatalf("build authorize url: %v", err)
	}
	parsed, err := url.Parse(urlStr)
	if err != nil {
		t.Fatalf("parse authorize url: %v", err)
	}
	if parsed.Query().Get("prompt") != "consent" {
		t.Fatalf("unexpected prompt: %s", parsed.Query().Get("prompt"))
	}
}

func TestBuildUserAuthorizeURLIncrementalScopes(t *testing.T) {
	scopeList := []string{"offline_access", "drive:drive", "calendar:calendar"}
	scopeValue := strings.Join(requestedUserOAuthScopes(scopeList, "offline_access drive:drive", true), " ")
	urlStr, err := buildUserAuthorizeURL("https://open.feishu.cn", "app-id", userOAuthRedirectURL, "state123", scopeValue, "", true)
	if err != nil {
		t.Fatalf("build authorize url: %v", err)
	}
	parsed, err := url.Parse(urlStr)
	if err != nil {
		t.Fatalf("parse authorize url: %v", err)
	}
	query := parsed.Query()
	if query.Get("include_granted_scopes") != "true" {
		t.Fatalf("expected include_granted_scopes, got %q", query.Get("include_granted_scopes"))
	}
	if query.Get("scope") != "offline_access calendar:calendar" {
		t.Fatalf("unexpected scope: %s", query.Get("scope"))
	}
}

func TestBuildUserAuthorizeURLNonIncrementalScopes(t *testing.T) {
	scopeList := []string{"offline_access", "drive:drive", "calendar:calendar"}
	scopeValue := strings.Join(requestedUserOAuthScopes(scopeList, "offline_access drive:drive", false), " ")
	urlStr, err := buildUserAuthorizeURL("https://open.feishu.cn", "app-id", userOAuthRedirectURL, "state123", scopeValue, "", false)
	if err != nil {
		t.Fatalf("build authorize url: %v", err)
	}
	parsed, err := url.Parse(urlStr)
	if err != nil {
		t.Fatalf("parse authorize url: %v", err)
	}
	query := parsed.Query()
	if query.Get("include_granted_scopes") != "" {
		t.Fatalf("unexpected include_granted_scopes: %s", query.Get("include_granted_scopes"))
	}
	if query.Get("scope") != "offline_access calendar:calendar drive:drive" {
		t.Fatalf("unexpected scope: %s", query.Get("scope"))
	}
}

func TestResolveUserOAuthScopesDefaultsToOfflineAccess(t *testing.T) {
	state := &appState{Config: config.Default()}
	scopes, source, err := resolveUserOAuthScopes(state, userOAuthScopeOptions{})
	if err != nil {
		t.Fatalf("resolve scopes: %v", err)
	}
	if len(scopes) != 1 || scopes[0] != defaultUserOAuthScope {
		t.Fatalf("unexpected scopes: %v", scopes)
	}
	if source != "default" {
		t.Fatalf("unexpected source: %s", source)
	}
}

func TestResolveUserOAuthScopesFromConfig(t *testing.T) {
	state := &appState{Config: &config.Config{UserScopes: []string{"offline_access", "drive:drive"}}}
	scopes, source, err := resolveUserOAuthScopes(state, userOAuthScopeOptions{})
	if err != nil {
		t.Fatalf("resolve scopes: %v", err)
	}
	if source != "config" {
		t.Fatalf("unexpected source: %s", source)
	}
	if strings.Join(scopes, " ") != "offline_access drive:drive" {
		t.Fatalf("unexpected scopes: %v", scopes)
	}
}

func TestResolveUserOAuthScopesFromServicesReadonly(t *testing.T) {
	state := &appState{Config: config.Default()}
	opts := userOAuthScopeOptions{
		Services:    []string{"drive"},
		ServicesSet: true,
		Readonly:    true,
	}
	scopes, source, err := resolveUserOAuthScopes(state, opts)
	if err != nil {
		t.Fatalf("resolve scopes: %v", err)
	}
	if source != "services" {
		t.Fatalf("unexpected source: %s", source)
	}
	if strings.Join(scopes, " ") != "offline_access drive:drive:readonly" {
		t.Fatalf("unexpected scopes: %v", scopes)
	}
}

func TestExchangeUserAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/authen/v2/oauth/token" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var payload map[string]string
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if payload["grant_type"] != "authorization_code" {
			t.Fatalf("unexpected grant_type: %s", payload["grant_type"])
		}
		if payload["client_id"] != "app-id" {
			t.Fatalf("unexpected client_id: %s", payload["client_id"])
		}
		if payload["client_secret"] != "app-secret" {
			t.Fatalf("unexpected client_secret: %s", payload["client_secret"])
		}
		if payload["code"] != "auth-code" {
			t.Fatalf("unexpected code: %s", payload["code"])
		}
		if payload["redirect_uri"] != userOAuthRedirectURL {
			t.Fatalf("unexpected redirect_uri: %s", payload["redirect_uri"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"access_token":"user-token","refresh_token":"refresh-token","expires_in":3600,"token_type":"Bearer","scope":"offline_access"}`)
	}))
	defer server.Close()

	token, err := exchangeUserAccessToken(context.Background(), server.Client(), server.URL, "app-id", "app-secret", "auth-code", userOAuthRedirectURL)
	if err != nil {
		t.Fatalf("exchange token: %v", err)
	}
	if token.AccessToken != "user-token" {
		t.Fatalf("unexpected access token: %s", token.AccessToken)
	}
	if token.RefreshToken != "refresh-token" {
		t.Fatalf("unexpected refresh token: %s", token.RefreshToken)
	}
	if token.ExpiresIn != 3600 {
		t.Fatalf("unexpected expires_in: %d", token.ExpiresIn)
	}
	if token.Scope != "offline_access" {
		t.Fatalf("unexpected scope: %s", token.Scope)
	}
}

func TestRequireUserRefreshToken(t *testing.T) {
	err := requireUserRefreshToken("")
	if err == nil {
		t.Fatalf("expected error for missing refresh_token")
	}
	if !strings.Contains(err.Error(), "offline access was not granted") {
		t.Fatalf("expected offline access hint, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), userOAuthReloginCommand) {
		t.Fatalf("expected re-run instruction, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "redirect URL/config") {
		t.Fatalf("expected redirect URL/config hint, got %q", err.Error())
	}

	if err := requireUserRefreshToken("refresh-token"); err != nil {
		t.Fatalf("unexpected error for refresh token: %v", err)
	}
}

func TestCanonicalScopeStringSortsAndDedupes(t *testing.T) {
	out := canonicalScopeString("drive.read drive.write drive.read")
	if out != "drive.read drive.write" {
		t.Fatalf("expected canonical scope, got %q", out)
	}
}

func TestScopesChangedWarningEmptyWhenNoPreviousScope(t *testing.T) {
	warning := scopesChangedWarning("", "offline_access")
	if warning != "" {
		t.Fatalf("expected no warning, got %q", warning)
	}
}

func TestScopesChangedWarningEmptyWhenSameScopeSet(t *testing.T) {
	warning := scopesChangedWarning("a b", "b a")
	if warning != "" {
		t.Fatalf("expected no warning, got %q", warning)
	}
}

func TestScopesChangedWarningIncludesReloginCommand(t *testing.T) {
	warning := scopesChangedWarning("offline_access", "offline_access contact:contact.base:readonly")
	if warning == "" {
		t.Fatalf("expected warning")
	}
	if !strings.Contains(strings.ToLower(warning), "scopes changed") {
		t.Fatalf("expected warning to mention scopes changed, got %q", warning)
	}
	if !strings.Contains(warning, userOAuthReloginCommand) {
		t.Fatalf("expected warning to include relogin command, got %q", warning)
	}
}
