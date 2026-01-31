package larksdk

import (
	"fmt"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

func formatCodeError(prefix string, codeErr larkcore.CodeError, apiResp *larkcore.ApiResp) error {
	msg := strings.TrimSpace(codeErr.Msg)
	if msg == "" {
		msg = "request failed"
	}
	detail := formatCodeErrorDetail(codeErr)
	if detail != "" {
		msg = fmt.Sprintf("%s (%s)", msg, detail)
	}
	if apiResp != nil {
		if reqID := apiResp.RequestId(); reqID != "" {
			msg = fmt.Sprintf("%s [request_id=%s]", msg, reqID)
		}
	}
	if prefix == "" {
		return fmt.Errorf("%s", msg)
	}
	return fmt.Errorf("%s: %s", prefix, msg)
}

func formatCodeErrorDetail(codeErr larkcore.CodeError) string {
	if codeErr.Err == nil {
		return ""
	}
	parts := make([]string, 0, 4)
	for _, detail := range codeErr.Err.Details {
		if detail == nil {
			continue
		}
		entry := strings.TrimSpace(strings.Join(nonEmpty(detail.Key, detail.Value), "="))
		if entry != "" {
			parts = append(parts, fmt.Sprintf("detail=%s", entry))
		}
	}
	for _, violation := range codeErr.Err.FieldViolations {
		if violation == nil {
			continue
		}
		segment := joinNonEmpty("field", violation.Field, "value", violation.Value, "desc", violation.Description)
		if segment != "" {
			parts = append(parts, fmt.Sprintf("field_violation{%s}", segment))
		}
	}
	for _, violation := range codeErr.Err.PermissionViolations {
		if violation == nil {
			continue
		}
		segment := joinNonEmpty("type", violation.Type, "subject", violation.Subject, "desc", violation.Description)
		if segment != "" {
			parts = append(parts, fmt.Sprintf("permission_violation{%s}", segment))
		}
	}
	return strings.Join(parts, "; ")
}

func joinNonEmpty(pairs ...string) string {
	if len(pairs)%2 != 0 {
		return ""
	}
	parts := make([]string, 0, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		key := strings.TrimSpace(pairs[i])
		value := strings.TrimSpace(pairs[i+1])
		if key == "" || value == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(parts, " ")
}

func nonEmpty(values ...string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		out = append(out, value)
	}
	return out
}
