package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

const maxUserSearchPageSize = 200

func newUsersCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "Manage users",
		Long: `Users are people in your tenant directory.

- User IDs: user_id (tenant-scoped), open_id (app-scoped), union_id (cross-app).
- Use search to resolve IDs before calling other APIs.`,
	}
	cmd.AddCommand(newUserInfoCmd(state))
	cmd.AddCommand(newUsersSearchCmd(state))
	return cmd
}

func newUsersSearchCmd(state *appState) *cobra.Command {
	var query string
	var limit int
	var pages int

	cmd := &cobra.Command{
		Use:     "search <search_query>",
		Aliases: []string{"list"},
		Short:   "Search users by keyword",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				return errors.New("search_query is required")
			}
			query = strings.TrimSpace(args[0])
			if query == "" {
				return errors.New("search_query is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			if limit <= 0 {
				return errors.New("limit must be greater than 0")
			}
			if pages <= 0 {
				return errors.New("pages must be greater than 0")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesUser)
			if err != nil {
				return err
			}

			users := make([]larksdk.User, 0, limit)
			pageToken := ""
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
				result, err := state.SDK.SearchUsers(context.Background(), token, larksdk.SearchUsersRequest{
					Query:     query,
					PageSize:  pageSize,
					PageToken: pageToken,
				})
				if err != nil {
					return withUserScopeHintForCommand(state, err)
				}
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

			payload := map[string]any{"users": users}
			lines := make([]string, 0, len(users))
			for _, user := range users {
				lines = append(lines, formatUserSearchLine(user))
			}
			text := tableText([]string{"user_id", "name", "open_id", "departments"}, lines, "no users found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 50, "max number of users to return")
	cmd.Flags().IntVar(&pages, "pages", 1, "max number of pages to fetch")

	return cmd
}

func formatUserSearchLine(user larksdk.User) string {
	id := user.UserID
	if id == "" {
		id = user.OpenID
	}
	departments := strings.Join(user.DepartmentIDs, ",")
	return fmt.Sprintf("%s\t%s\t%s\t%s", id, user.Name, user.OpenID, departments)
}

func formatUserLine(user larksdk.User) string {
	id := user.UserID
	if id == "" {
		id = user.OpenID
	}
	return fmt.Sprintf("%s\t%s\t%s\t%s", id, user.Name, user.Email, user.Mobile)
}
