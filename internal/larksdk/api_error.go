package larksdk

import "fmt"

// APIError is a structured error for Lark/Feishu OpenAPI responses.
//
// Why: Many SDK response objects include both Code and Msg. We want to preserve
// the numeric code in the error string so downstream heuristics can reliably
// distinguish insufficient OAuth scopes vs. document/wiki permissions, etc.
//
// The error string intentionally contains "code=<n>" so cmd/lark/user_scope_hint.go
// can parse it.
type APIError struct {
	Op   string
	Code int
	Msg  string
}

func (e APIError) Error() string {
	if e.Code != 0 {
		return fmt.Sprintf("%s failed (code=%d): %s", e.Op, e.Code, e.Msg)
	}
	return fmt.Sprintf("%s failed: %s", e.Op, e.Msg)
}

func apiError(op string, code int, msg string) error {
	return APIError{Op: op, Code: code, Msg: msg}
}
