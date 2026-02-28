package authregistry

import "sort"

// TokenType describes the type(s) of access token a given service may require.
//
// This is an internal registry used to drive a services→scopes→credentials auth UX
// and should remain stable/deterministic.
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

// ServiceDef is a service definition for auth + capability mapping.
//
// Services are capabilities (drive/docs/mail/...) rather than concrete CLI
// commands. Commands map to services at runtime (future work).
type ServiceDef struct {
	Name string `json:"name"`

	// TokenTypes declares which tokens a service may require.
	TokenTypes []TokenType `json:"token_types,omitempty"`

	// RequiredUserScopes is the minimal required user OAuth scope set (no variants).
	//
	// This is used by the services→scopes model and should remain
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
		RequiredUserScopes: []string{"im:chat:read"},
		UserScopes: ServiceScopeSet{
			Full:     []string{"im:chat:read", "im:chat.members:read"},
			Readonly: []string{"im:chat:read"},
		},
	},
	"search-message": {Name: "search-message", TokenTypes: []TokenType{TokenUser}, RequiredUserScopes: []string{"im:message:readonly", "search:message"}, RequiresOffline: true},
	"search-user": {
		Name:       "search-user",
		TokenTypes: []TokenType{TokenUser},
		RequiredUserScopes: []string{
			"contact:contact.base:readonly",
			"contact:user.employee_id:readonly",
			"contact:user:search",
		},
		RequiresOffline: true,
	},
	"search-docs": {Name: "search-docs", TokenTypes: []TokenType{TokenUser}, RequiredUserScopes: []string{"search:docs:read"}, RequiresOffline: true},

	// NOTE: "docs" is a legacy name used by existing commands; "docx" is the
	// user-facing/API surface name for the same capability. Keep them aligned.

	// Drive ("My Space") uses dedicated scopes per operation. Avoid the legacy
	// broad drive:drive / drive:drive:readonly scopes.
	"drive-search": {
		Name:               "drive search",
		TokenTypes:         []TokenType{TokenUser},
		RequiredUserScopes: []string{"drive:drive.search:readonly"},
		UserScopes: ServiceScopeSet{
			// Search is read-only.
			Full:     []string{"drive:drive.search:readonly"},
			Readonly: []string{"drive:drive.search:readonly"},
		},
		RequiresOffline: true,
	},
	"drive-metadata": {
		Name:       "drive metadata",
		TokenTypes: []TokenType{TokenTenant, TokenUser},
		// drive:drive.metadata:readonly covers file/folder metadata; space:document:retrieve
		// is required by list endpoints.
		RequiredUserScopes: []string{"drive:drive.metadata:readonly", "space:document:retrieve"},
		UserScopes: ServiceScopeSet{
			Full:     []string{"drive:drive.metadata:readonly", "space:document:retrieve"},
			Readonly: []string{"drive:drive.metadata:readonly", "space:document:retrieve"},
		},
		RequiresOffline: true,
	},
	"drive-download": {
		Name:               "drive download",
		TokenTypes:         []TokenType{TokenTenant, TokenUser},
		RequiredUserScopes: []string{"drive:file:download"},
		UserScopes: ServiceScopeSet{
			Full:     []string{"drive:file:download"},
			Readonly: []string{"drive:file:download"},
		},
		RequiresOffline: true,
	},
	"drive-upload": {
		Name:               "drive upload",
		TokenTypes:         []TokenType{TokenTenant, TokenUser},
		RequiredUserScopes: []string{"drive:file:upload"},
		UserScopes: ServiceScopeSet{
			Full: []string{"drive:file:upload"},
		},
		RequiresOffline: true,
	},
	"drive-permissions": {
		Name:       "drive permissions",
		TokenTypes: []TokenType{TokenTenant, TokenUser},
		// Collaborator management + public permission settings.
		RequiredUserScopes: []string{
			"docs:permission.member:create",
			"docs:permission.member:delete",
			"docs:permission.member:retrieve",
			"docs:permission.member:update",
			"docs:permission.setting:write_only",
		},
		UserScopes: ServiceScopeSet{
			Full: []string{
				"docs:permission.member:create",
				"docs:permission.member:delete",
				"docs:permission.member:retrieve",
				"docs:permission.member:update",
				"docs:permission.setting:write_only",
			},
			Readonly: []string{"docs:permission.member:retrieve"},
		},
		RequiresOffline: true,
	},
	"drive-comment-read": {
		Name:               "drive comment read",
		TokenTypes:         []TokenType{TokenTenant, TokenUser},
		RequiredUserScopes: []string{"docs:document.comment:read"},
		UserScopes: ServiceScopeSet{
			Full:     []string{"docs:document.comment:read"},
			Readonly: []string{"docs:document.comment:read"},
		},
		RequiresOffline: true,
	},
	"drive-comment-write": {
		Name:       "drive comment write",
		TokenTypes: []TokenType{TokenTenant, TokenUser},
		RequiredUserScopes: []string{
			"docs:document.comment:create",
			"docs:document.comment:update",
		},
		UserScopes: ServiceScopeSet{
			Full: []string{
				"docs:document.comment:create",
				"docs:document.comment:update",
			},
		},
		RequiresOffline: true,
	},
	"docs": {
		Name:               "docs",
		TokenTypes:         []TokenType{TokenTenant, TokenUser},
		RequiredUserScopes: []string{"docx:document:readonly"},
		UserScopes: ServiceScopeSet{
			Full:     []string{"docx:document.block:convert", "docx:document:create", "docx:document:readonly", "docx:document:write_only"},
			Readonly: []string{"docx:document:readonly"},
		},
		RequiresOffline: true,
	},
	"docx": {
		Name:               "docx",
		TokenTypes:         []TokenType{TokenTenant, TokenUser},
		RequiredUserScopes: []string{"docx:document:readonly"},
		UserScopes: ServiceScopeSet{
			Full:     []string{"docx:document.block:convert", "docx:document:create", "docx:document:readonly", "docx:document:write_only"},
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
			// Tasklist APIs require the read scope even when write is granted.
			// Request both to avoid confusing "missing scope" errors.
			Full:     []string{"task:tasklist:read", "task:tasklist:write"},
			Readonly: []string{"task:tasklist:read"},
		},
		RequiresOffline: true,
	},
	"tasklist-write": {
		Name:               "tasklist write",
		TokenTypes:         []TokenType{TokenTenant, TokenUser},
		RequiredUserScopes: []string{"task:tasklist:write"},
		UserScopes: ServiceScopeSet{
			// Keep compatibility for callers that explicitly ask for write-only,
			// but include read to ensure list/read endpoints work.
			Full: []string{"task:tasklist:read", "task:tasklist:write"},
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
	"vc-meeting": {
		Name:       "vc meeting",
		TokenTypes: []TokenType{TokenUser},
		// NOTE: Some Feishu VC privilege strings appear in error messages but do not
		// surface as selectable OAuth scopes in the developer console. Keep the
		// registry limited to confirmed OAuth scopes.
		RequiredUserScopes: []string{
			"vc:meeting:readonly",
		},
		UserScopes: ServiceScopeSet{
			Readonly: []string{
				"vc:meeting:readonly",
			},
		},
		RequiresOffline: true,
	},
	"base": {Name: "base", TokenTypes: []TokenType{TokenTenant}},
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
