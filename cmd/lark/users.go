package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

const maxUsersPageSize = 50

func newUsersCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "Manage users",
	}
	cmd.AddCommand(newUserGetCmd(state))
	cmd.AddCommand(newUsersSearchCmd(state))
	return cmd
}

func newUsersSearchCmd(state *appState) *cobra.Command {
	var email string
	var mobile string
	var name string
	var departmentID string

	cmd := &cobra.Command{
		Use:     "search",
		Aliases: []string{"list"},
		Short:   "Search users by email, mobile, or name",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			var users []larksdk.User
			switch {
			case email != "" || mobile != "":
				token, err := tokenFor(context.Background(), state, tokenTypesTenant)
				if err != nil {
					return err
				}
				result, err := state.SDK.BatchGetUserIDs(context.Background(), token, larksdk.BatchGetUserIDRequest{
					Emails:  nonEmptyList(email),
					Mobiles: nonEmptyList(mobile),
				})
				if err != nil {
					return err
				}
				users = result
			case name != "":
				token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
				if err != nil {
					return err
				}
				matches, err := searchUsersByName(context.Background(), state.SDK.ListUsersByDepartment, token, departmentID, name)
				if err != nil {
					return err
				}
				users = matches
			}

			payload := map[string]any{"users": users}
			lines := make([]string, 0, len(users))
			for _, user := range users {
				lines = append(lines, formatUserLine(user))
			}
			text := tableText([]string{"user_id", "name", "email", "mobile"}, lines, "no users found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "search by email")
	cmd.Flags().StringVar(&mobile, "mobile", "", "search by mobile")
	cmd.Flags().StringVar(&name, "name", "", "search by name")
	cmd.Flags().StringVar(&departmentID, "department-id", "0", "department ID for name search")
	cmd.MarkFlagsOneRequired("email", "mobile", "name")
	cmd.MarkFlagsMutuallyExclusive("email", "mobile", "name")

	return cmd
}

func nonEmptyList(value string) []string {
	if value == "" {
		return nil
	}
	return []string{value}
}

func searchUsersByName(ctx context.Context, listUsersByDepartment func(context.Context, string, larksdk.ListUsersByDepartmentRequest) (larksdk.ListUsersByDepartmentResult, error), token, departmentID, name string) ([]larksdk.User, error) {
	pageToken := ""
	matches := []larksdk.User{}
	needle := strings.ToLower(name)
	for {
		result, err := listUsersByDepartment(ctx, token, larksdk.ListUsersByDepartmentRequest{
			DepartmentID: departmentID,
			PageSize:     maxUsersPageSize,
			PageToken:    pageToken,
		})
		if err != nil {
			return nil, err
		}
		for _, user := range result.Items {
			if strings.Contains(strings.ToLower(user.Name), needle) {
				matches = append(matches, user)
			}
		}
		if !result.HasMore || result.PageToken == "" {
			break
		}
		pageToken = result.PageToken
	}
	return matches, nil
}

func formatUserLine(user larksdk.User) string {
	id := user.UserID
	if id == "" {
		id = user.OpenID
	}
	return fmt.Sprintf("%s\t%s\t%s\t%s", id, user.Name, user.Email, user.Mobile)
}
