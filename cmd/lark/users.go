package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

const maxUserSearchPageSize = 200

func newUsersCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "Search and inspect directory users",
		Long: `Users are people in your tenant directory.

IDs: user_id (tenant-scoped), open_id (app-scoped), union_id (cross-app).
Tip: use search to resolve IDs before calling other APIs.`,
		Example: `  lark users search "Ada"
  lark users info <user_id>`,
	}
	cmd.AddCommand(newUserInfoCmd(state))
	cmd.AddCommand(newUsersSearchCmd(state))
	return cmd
}

func newUsersSearchCmd(state *appState) *cobra.Command {
	var query string
	var email string
	var limit int
	var pages int

	cmd := &cobra.Command{
		Use:     "search <search_query>",
		Aliases: []string{"list"},
		Short:   "Search users by keyword",
		Example: `  lark users search "Ada"
  lark users search --email "ada@example.com"`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			email = strings.TrimSpace(email)
			if len(args) == 0 && email == "" {
				return usageError(cmd, "search_query is required", `Examples:
  lark users search "Ada"
  lark users search --email "ada@example.com"`)
			}
			if len(args) > 0 && email != "" {
				return usageError(cmd, "search_query and --email cannot be used together", `Examples:
  lark users search "Ada"
  lark users search --email "ada@example.com"`)
			}
			if len(args) > 0 {
				query = strings.TrimSpace(args[0])
				if query == "" {
					return usageError(cmd, "search_query is required", `Examples:
  lark users search "Ada"
  lark users search --email "ada@example.com"`)
				}
			} else {
				query = email
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := requireSDK(state); err != nil {
				return err
			}
			if limit <= 0 {
				return flagUsage(cmd, "limit must be greater than 0")
			}
			if pages <= 0 {
				return flagUsage(cmd, "pages must be greater than 0")
			}
			ctx := cmd.Context()
			token, err := tokenFor(ctx, state, tokenTypesUser)
			if err != nil {
				return err
			}

			users := make([]larksdk.User, 0, limit)
			pageToken := ""
			nextPageToken := ""
			hasMore := false
			pageCount := 0
			remaining := limit
			for {
				if pageCount >= pages {
					break
				}
				pageCount++
				pageSize := remaining
				if pageSize > maxUserSearchPageSize {
					pageSize = maxUserSearchPageSize
				}
				if pageSize <= 0 {
					break
				}
				result, err := state.SDK.SearchUsers(ctx, token, larksdk.SearchUsersRequest{
					Query:     query,
					PageSize:  pageSize,
					PageToken: pageToken,
				})
				if err != nil {
					return withUserScopeHintForCommand(state, err)
				}
				nextPageToken = result.PageToken
				hasMore = result.HasMore
				users = append(users, result.Users...)
				if len(users) >= limit || !result.HasMore {
					break
				}
				remaining = limit - len(users)
				pageToken = result.PageToken
				if strings.TrimSpace(pageToken) == "" {
					break
				}
			}
			if len(users) > limit {
				users = users[:limit]
			}

			payload := map[string]any{
				"users":           users,
				"has_more":        hasMore,
				"next_page_token": nextPageToken,
			}
			lines := make([]string, 0, len(users))
			for _, user := range users {
				lines = append(lines, formatUserSearchLine(user))
			}
			text := tableText([]string{"user_id", "name", "email", "departments"}, lines, "no users found")
			return state.Printer.Print(payload, text)
		},
	}
	annotateAuthServices(cmd, "search-user")

	cmd.Flags().StringVar(&email, "email", "", "search by exact email address")
	cmd.Flags().IntVar(&limit, "limit", 50, "max number of users to return")
	cmd.Flags().IntVar(&pages, "pages", 1, "max number of pages to fetch")

	return cmd
}

func formatUserSearchLine(user larksdk.User) string {
	id := user.UserID
	if id == "" {
		id = user.OpenID
	}
	email := user.Email
	if email == "" {
		email = user.EnterpriseEmail
	}
	departments := formatUserDepartments(user)
	return fmt.Sprintf("%s\t%s\t%s\t%s", id, user.Name, email, departments)
}

func formatUserLine(user larksdk.User) string {
	id := user.UserID
	if id == "" {
		id = user.OpenID
	}
	return fmt.Sprintf("%s\t%s\t%s\t%s", id, user.Name, user.Email, user.Mobile)
}

func formatUserDepartments(user larksdk.User) string {
	if len(user.Departments) > 0 {
		parts := make([]string, 0, len(user.Departments))
		for _, dept := range user.Departments {
			if dept.ID == "" {
				continue
			}
			if dept.Name != "" {
				parts = append(parts, fmt.Sprintf("%s(%s)", dept.Name, dept.ID))
				continue
			}
			parts = append(parts, dept.ID)
		}
		return strings.Join(parts, ",")
	}
	return strings.Join(user.DepartmentIDs, ",")
}
