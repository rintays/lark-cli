package testutil

import (
	"os"
	"testing"
)

// RequireEnv returns the value of an env var or skips the test if it's empty.
func RequireEnv(t *testing.T, key string) string {
	t.Helper()
	val := os.Getenv(key)
	if val == "" {
		t.Skipf("missing %s", key)
	}
	return val
}
