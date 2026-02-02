package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var receiveIDTypeValues = []string{"chat_id", "open_id", "user_id", "email"}
var driveMemberTypeValues = []string{"openid", "userid", "email", "openchat", "opendepartmentid"}
var drivePermValues = []string{"view", "edit", "full_access"}
var drivePermTypeValues = []string{"container", "single_page"}

func normalizeReceiveIDType(raw string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "chat_id", "chatid":
		return "chat_id", true
	case "open_id", "openid":
		return "open_id", true
	case "user_id", "userid":
		return "user_id", true
	case "email":
		return "email", true
	default:
		return "", false
	}
}

func normalizeDriveMemberType(raw string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "open_id", "openid":
		return "openid", true
	case "user_id", "userid":
		return "userid", true
	case "email":
		return "email", true
	case "openchat":
		return "openchat", true
	case "opendepartmentid", "open_department_id", "opendepartment_id":
		return "opendepartmentid", true
	default:
		return "", false
	}
}

func validateOneOf(cmd *cobra.Command, label, value string, allowed []string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	for _, entry := range allowed {
		if value == entry {
			return nil
		}
	}
	return usageErrorWithUsage(cmd, fmt.Sprintf("%s must be one of %s", label, strings.Join(allowed, ", ")), "", cmd.UsageString())
}

func registerEnumCompletion(cmd *cobra.Command, flag string, values []string) {
	if cmd == nil {
		return
	}
	_ = cmd.RegisterFlagCompletionFunc(flag, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return values, cobra.ShellCompDirectiveNoFileComp
	})
}
