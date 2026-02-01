package larksdk

import (
	"fmt"
	"strings"
)

// userOAuthReloginCommand mirrors the CLI remediation command used elsewhere.
// Kept here (internal package) so larksdk can append actionable guidance.
const userOAuthReloginCommand = "lark auth user login --scopes \"offline_access\" --force-consent"

// withInsufficientScopeRemediation appends a best-effort remediation hint when the
// server message suggests missing permission/scope.
func withInsufficientScopeRemediation(err error, serverMsg string) error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(serverMsg)
	if msg == "" {
		msg = strings.ToLower(err.Error())
	}
	// Heuristic: SDK/OpenAPI errors vary by endpoint; keep this conservative.
	if !(strings.Contains(msg, "insufficient") || strings.Contains(msg, "scope") || strings.Contains(msg, "permission") || strings.Contains(msg, "forbidden")) {
		return err
	}
	return fmt.Errorf("%w; this may be due to missing permission/scope; try re-authorizing with: `%s`", err, userOAuthReloginCommand)
}
