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
	"im": {
		Name:               "im",
		TokenTypes:         []TokenType{TokenTenant, TokenUser},
		RequiredUserScopes: []string{"im:chat.group_info:readonly"},
		UserScopes: ServiceScopeSet{
			Full:     []string{"im:chat"},
			Readonly: []string{"im:chat.group_info:readonly"},
		},
	},
	"search-message": {Name: "search-message", TokenTypes: []TokenType{TokenUser}, RequiredUserScopes: []string{"im:message:readonly", "search:message"}, RequiresOffline: true},
	"search-user":    {Name: "search-user", TokenTypes: []TokenType{TokenUser}, RequiredUserScopes: []string{"contact:contact.base:readonly", "contact:user:search"}, RequiresOffline: true},
	"search-docs":    {Name: "search-docs", TokenTypes: []TokenType{TokenUser}, RequiredUserScopes: []string{"search:docs:read"}, RequiresOffline: true},

	// NOTE: "docs" is a legacy name used by existing commands; "docx" is the
	// user-facing/API surface name for the same capability. Keep them aligned.
	"drive": {Name: "drive", TokenTypes: []TokenType{TokenTenant, TokenUser}, RequiredUserScopes: []string{"drive:drive"}, UserScopes: ServiceScopeSet{Full: []string{"drive:drive"}, Readonly: []string{"drive:drive:readonly"}}, RequiresOffline: true},
	"docs": {
		Name:               "docs",
		TokenTypes:         []TokenType{TokenTenant, TokenUser},
		RequiredUserScopes: []string{"docx:document:readonly"},
		UserScopes: ServiceScopeSet{
			Full:     []string{"docx:document", "docx:document.block:convert", "docx:document:create", "docx:document:readonly", "docx:document:write_only"},
			Readonly: []string{"docx:document:readonly"},
		},
		RequiresOffline: true,
	},
	"docx": {
		Name:               "docx",
		TokenTypes:         []TokenType{TokenTenant, TokenUser},
		RequiredUserScopes: []string{"docx:document:readonly"},
		UserScopes: ServiceScopeSet{
			Full:     []string{"docx:document", "docx:document.block:convert", "docx:document:create", "docx:document:readonly", "docx:document:write_only"},
			Readonly: []string{"docx:document:readonly"},
		},
		RequiresOffline: true,
	},
	"sheets": {
		Name:               "sheets",
		TokenTypes:         []TokenType{TokenTenant, TokenUser},
		RequiredUserScopes: []string{"sheets:spreadsheet:read"},
		UserScopes: ServiceScopeSet{
			Full:     []string{"sheets:spreadsheet:create", "sheets:spreadsheet:read", "sheets:spreadsheet:write_only", "sheets:spreadsheet.meta:read"},
			Readonly: []string{"sheets:spreadsheet:readonly"},
		},
		RequiresOffline: true,
	},

	"calendar": {Name: "calendar", TokenTypes: []TokenType{TokenTenant, TokenUser}, RequiredUserScopes: []string{"calendar:calendar"}, UserScopes: ServiceScopeSet{Full: []string{"calendar:calendar"}, Readonly: []string{"calendar:calendar:readonly"}}},
	"task": {
		Name:               "task",
		TokenTypes:         []TokenType{TokenTenant, TokenUser},
		RequiredUserScopes: []string{"task:task:read"},
		UserScopes: ServiceScopeSet{
			Full:     []string{"task:task:write"},
			Readonly: []string{"task:task:read"},
		},
		RequiresOffline: true,
	},
	"task-write": {
		Name:               "task write",
		TokenTypes:         []TokenType{TokenTenant, TokenUser},
		RequiredUserScopes: []string{"task:task:write"},
		UserScopes: ServiceScopeSet{
			Full: []string{"task:task:write"},
		},
		RequiresOffline: true,
	},
	"tasklist": {
		Name:               "tasklist",
		TokenTypes:         []TokenType{TokenTenant, TokenUser},
		RequiredUserScopes: []string{"task:tasklist:read"},
		UserScopes: ServiceScopeSet{
			Full:     []string{"task:tasklist:write"},
			Readonly: []string{"task:tasklist:read"},
		},
		RequiresOffline: true,
	},
	"tasklist-write": {
		Name:               "tasklist write",
		TokenTypes:         []TokenType{TokenTenant, TokenUser},
		RequiredUserScopes: []string{"task:tasklist:write"},
		UserScopes: ServiceScopeSet{
			Full: []string{"task:tasklist:write"},
		},
		RequiresOffline: true,
	},
	"mail": {
		Name:       "mail",
		TokenTypes: []TokenType{TokenTenant, TokenUser},
		RequiredUserScopes: []string{
			"mail:user_mailbox.message:readonly",
			"mail:user_mailbox.message.subject:read",
			"mail:user_mailbox.message.address:read",
			"mail:user_mailbox.message.body:read",
		},
		UserScopes: ServiceScopeSet{
			Readonly: []string{
				"mail:user_mailbox.message:readonly",
				"mail:user_mailbox.message.subject:read",
				"mail:user_mailbox.message.address:read",
				"mail:user_mailbox.message.body:read",
			},
			Full: []string{
				"mail:user_mailbox.message:readonly",
				"mail:user_mailbox.message.subject:read",
				"mail:user_mailbox.message.address:read",
				"mail:user_mailbox.message.body:read",
				"mail:user_mailbox.message:send",
			},
		},
		RequiresOffline: true,
	},
	"mail-send":    {Name: "mail send", TokenTypes: []TokenType{TokenUser}, RequiredUserScopes: []string{"mail:user_mailbox.message:send"}, RequiresOffline: true},
	"mail-public":  {Name: "mail public", TokenTypes: []TokenType{TokenTenant}},
	"drive-export": {Name: "drive export", TokenTypes: []TokenType{TokenTenant, TokenUser}, RequiredUserScopes: []string{"drive:export:readonly"}, RequiresOffline: true},
	"wiki":         {Name: "wiki", TokenTypes: []TokenType{TokenTenant, TokenUser}, RequiredUserScopes: []string{"wiki:wiki"}, UserScopes: ServiceScopeSet{Full: []string{"wiki:wiki"}, Readonly: []string{"wiki:wiki:readonly"}}, RequiresOffline: true},
	"base":         {Name: "base", TokenTypes: []TokenType{TokenTenant}},
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
