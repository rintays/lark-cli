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
)

func TestBuildUserAuthorizeURL(t *testing.T) {
	urlStr, err := buildUserAuthorizeURL("https://open.feishu.cn", "app-id", userOAuthRedirectURL, "state123", "offline_access")
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
}

func TestUserOAuthScopeDefaultsToOfflineAccess(t *testing.T) {
	scope := userOAuthScope("", false)
	if scope != defaultUserOAuthScope {
		t.Fatalf("expected default scope %q, got %q", defaultUserOAuthScope, scope)
	}
}

func TestUserOAuthScopeRespectsExplicitScope(t *testing.T) {
	scope := userOAuthScope("contact:contact.base:readonly", true)
	if scope != "contact:contact.base:readonly" {
		t.Fatalf("expected explicit scope, got %q", scope)
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
}

func TestRequireUserRefreshToken(t *testing.T) {
	err := requireUserRefreshToken("")
	if err == nil {
		t.Fatalf("expected error for missing refresh_token")
	}
	if !strings.Contains(err.Error(), "offline access was not granted") {
		t.Fatalf("expected offline access hint, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "lark auth user login --scope offline_access") {
		t.Fatalf("expected re-run instruction, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "redirect URL/config") {
		t.Fatalf("expected redirect URL/config hint, got %q", err.Error())
	}

	if err := requireUserRefreshToken("refresh-token"); err != nil {
		t.Fatalf("unexpected error for refresh token: %v", err)
	}
}
