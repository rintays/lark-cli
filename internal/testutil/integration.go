package testutil

import (
	"os"
	"testing"
)

// RequireIntegration skips when integration tests are not explicitly enabled.
func RequireIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("LARK_INTEGRATION") == "" && os.Getenv("INTEGRATION") == "" {
		t.Skip("integration tests disabled; set LARK_INTEGRATION=1 or INTEGRATION=1")
	}
}
