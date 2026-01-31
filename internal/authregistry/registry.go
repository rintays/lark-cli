package authregistry

import "sort"

// TokenType describes the type(s) of access token a given service may require.
//
// This is an internal registry used to drive gog-style auth UX (services → scopes
// → credentials) and should remain stable/deterministic.
//
// NOTE: Not all services currently declare OAuth scopes; those will be filled
// in incrementally as we learn the correct Feishu/Lark scope strings.
type TokenType string

const (
	TokenTenant TokenType = "tenant"
	TokenUser   TokenType = "user"
)

// ServiceScopeSet defines OAuth scope variants for a service.
//
// In the future we may add more variants (e.g. drive file/full levels).
type ServiceScopeSet struct {
	Full     []string `json:"full,omitempty"`
	Readonly []string `json:"readonly,omitempty"`
}

// ServiceDef is a gog-style service definition.
//
// Services are capabilities (drive/docs/mail/...) rather than concrete CLI
// commands. Commands map to services at runtime (future work).
type ServiceDef struct {
	Name string `json:"name"`

	// TokenTypes declares which tokens a service may require.
	TokenTypes []TokenType `json:"token_types,omitempty"`

	// RequiredUserScopes is the minimal required user OAuth scope set (no variants).
	//
	// This is used by the gog-style “services → scopes” model and should remain
	// stable/deterministic.
	RequiredUserScopes []string `json:"required_user_scopes,omitempty"`

	// UserScopes defines OAuth scope variants for user OAuth auth flows.
	//
	// TODO: migrate commands to use RequiredUserScopes + variants derived from it.
	UserScopes ServiceScopeSet `json:"user_scopes,omitempty"`

	RequiresOffline bool `json:"requires_offline,omitempty"`
}

// Registry is the fixed set of known services.
//
// Keep this list stable and append-only where possible.
var Registry = map[string]ServiceDef{
	"im":       {Name: "im", TokenTypes: []TokenType{TokenTenant}},
	"drive":    {Name: "drive", TokenTypes: []TokenType{TokenTenant, TokenUser}, UserScopes: ServiceScopeSet{Full: []string{"drive:drive"}, Readonly: []string{"drive:drive:readonly"}}, RequiresOffline: true},
	"docs":     {Name: "docs", TokenTypes: []TokenType{TokenTenant, TokenUser}, UserScopes: ServiceScopeSet{Full: []string{"drive:drive"}, Readonly: []string{"drive:drive:readonly"}}, RequiresOffline: true},
	"sheets":   {Name: "sheets", TokenTypes: []TokenType{TokenTenant, TokenUser}, UserScopes: ServiceScopeSet{Full: []string{"drive:drive"}, Readonly: []string{"drive:drive:readonly"}}, RequiresOffline: true},
	"calendar": {Name: "calendar", TokenTypes: []TokenType{TokenTenant}},
	"mail":     {Name: "mail", TokenTypes: []TokenType{TokenTenant, TokenUser}, RequiresOffline: true},
	"wiki":     {Name: "wiki", TokenTypes: []TokenType{TokenTenant, TokenUser}, RequiresOffline: true},
	"base":     {Name: "base", TokenTypes: []TokenType{TokenTenant}},
	"contacts": {Name: "contacts", TokenTypes: []TokenType{TokenTenant}},
	"meetings": {Name: "meetings", TokenTypes: []TokenType{TokenTenant}},
	"minutes":  {Name: "minutes", TokenTypes: []TokenType{TokenTenant}},
}

// AllServiceNames returns all known service names in stable-sorted order.
func AllServiceNames() []string {
	out := make([]string, 0, len(Registry))
	for name := range Registry {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
