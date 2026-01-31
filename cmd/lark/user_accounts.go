package main

import (
	"os"
	"sort"
	"strings"

	"lark/internal/config"
)

const defaultUserAccountName = "default"

func resolveUserAccountName(state *appState) string {
	if state == nil {
		return defaultUserAccountName
	}
	name := strings.TrimSpace(state.UserAccount)
	if name == "" {
		name = strings.TrimSpace(os.Getenv("LARK_ACCOUNT"))
	}
	if name == "" && state.Config != nil {
		name = strings.TrimSpace(state.Config.DefaultUserAccount)
	}
	if name == "" {
		name = defaultUserAccountName
	}
	return name
}

func normalizeAccountName(raw string) string {
	return strings.TrimSpace(raw)
}

func listUserAccountNames(cfg *config.Config) []string {
	if cfg == nil {
		return nil
	}
	seen := map[string]struct{}{}
	for name := range cfg.UserAccounts {
		if strings.TrimSpace(name) == "" {
			continue
		}
		seen[name] = struct{}{}
	}
	if cfg.DefaultUserAccount != "" {
		seen[cfg.DefaultUserAccount] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func loadUserAccount(cfg *config.Config, name string) (config.UserAccount, bool) {
	if cfg == nil {
		return config.UserAccount{}, false
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return config.UserAccount{}, false
	}
	if cfg.UserAccounts != nil {
		if acct, ok := cfg.UserAccounts[name]; ok && acct != nil {
			return *acct, true
		}
	}
	return config.UserAccount{}, false
}

func ensureUserAccount(cfg *config.Config, name string) config.UserAccount {
	if acct, ok := loadUserAccount(cfg, name); ok {
		return acct
	}
	return config.UserAccount{}
}

func saveUserAccount(cfg *config.Config, name string, acct config.UserAccount) {
	if cfg == nil {
		return
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	if cfg.UserAccounts == nil {
		cfg.UserAccounts = map[string]*config.UserAccount{}
	}
	copy := acct
	cfg.UserAccounts[name] = &copy
}

func clearUserAccountTokens(cfg *config.Config, name string) bool {
	if cfg == nil {
		return false
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	acct, ok := loadUserAccount(cfg, name)
	if !ok {
		return false
	}
	acct.UserAccessToken = ""
	acct.RefreshToken = ""
	acct.UserAccessTokenExpiresAt = 0
	saveUserAccount(cfg, name, acct)
	return true
}

func deleteUserAccount(cfg *config.Config, name string) {
	if cfg == nil {
		return
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	if cfg.UserAccounts != nil {
		delete(cfg.UserAccounts, name)
	}
}
