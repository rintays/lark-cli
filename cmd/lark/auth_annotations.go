package main

import (
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/authregistry"
)

const authServicesAnnotationKey = "auth.services"

func annotateAuthServices(cmd *cobra.Command, services ...string) {
	if cmd == nil || len(services) == 0 {
		return
	}
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[authServicesAnnotationKey] = strings.Join(services, ",")
}

func registerAuthServices(root *cobra.Command) {
	if root == nil {
		return
	}
	mapping := map[string][]string{}
	var walk func(cmd *cobra.Command)
	walk = func(cmd *cobra.Command) {
		if cmd == nil {
			return
		}
		services := parseAuthServicesAnnotation(cmd)
		if len(services) > 0 {
			if path := canonicalCommandPath(cmd); path != "" {
				mapping[path] = services
			}
		}
		for _, child := range cmd.Commands() {
			walk(child)
		}
	}
	walk(root)
	if len(mapping) == 0 {
		return
	}
	authregistry.SetCommandServiceMap(mapping)
}

func parseAuthServicesAnnotation(cmd *cobra.Command) []string {
	if cmd == nil || cmd.Annotations == nil {
		return nil
	}
	raw := strings.TrimSpace(cmd.Annotations[authServicesAnnotationKey])
	if raw == "" {
		return nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case ',', ' ', '\t', '\n', ';':
			return true
		default:
			return false
		}
	})
	if len(parts) == 0 {
		return nil
	}
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		part = strings.ToLower(strings.TrimSpace(part))
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		out = append(out, part)
	}
	return out
}
