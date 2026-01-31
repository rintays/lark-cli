package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newMeetingsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meetings",
		Short: "Manage meetings",
	}
	cmd.AddCommand(newMeetingInfoCmd(state))
	cmd.AddCommand(newMeetingListCmd(state))
	return cmd
}

func newMeetingInfoCmd(state *appState) *cobra.Command {
	var meetingID string
	var withParticipants bool
	var withMeetingAbility bool
	var userIDType string
	var queryMode int

	cmd := &cobra.Command{
		Use:   "info <meeting-id>",
		Short: "Show meeting details",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) > 0 {
				if meetingID != "" && meetingID != args[0] {
					return errors.New("meeting-id provided twice")
				}
				if err := cmd.Flags().Set("meeting-id", args[0]); err != nil {
					return err
				}
				return nil
			}
			if strings.TrimSpace(meetingID) == "" {
				return errors.New("meeting-id is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if queryMode < 0 || queryMode > 1 {
				return errors.New("query-mode must be 0 or 1")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			meeting, err := state.SDK.GetMeeting(context.Background(), token, larksdk.GetMeetingRequest{
				MeetingID:          meetingID,
				WithParticipants:   withParticipants,
				WithMeetingAbility: withMeetingAbility,
				UserIDType:         userIDType,
				QueryMode:          queryMode,
			})
			if err != nil {
				return err
			}
			payload := map[string]any{"meeting": meeting}
			text := tableTextRow(
				[]string{"meeting_id", "topic", "status", "start_time", "end_time"},
				[]string{
					meeting.ID,
					meeting.Topic,
					fmt.Sprintf("%d", meeting.Status),
					meeting.StartTime,
					meeting.EndTime,
				},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&meetingID, "meeting-id", "", "meeting ID (or provide as positional argument)")
	cmd.Flags().BoolVar(&withParticipants, "with-participants", false, "include participant list")
	cmd.Flags().BoolVar(&withMeetingAbility, "with-ability", false, "include meeting ability stats")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user ID type (user_id, union_id, open_id)")
	cmd.Flags().IntVar(&queryMode, "query-mode", 0, "query mode: 0 for meeting info, 1 for related artifacts")
	return cmd
}

func newMeetingListCmd(state *appState) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List meetings",
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				return errors.New("limit must be greater than 0")
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}

			meetings := make([]larksdk.MeetingListItem, 0, limit)
			pageToken := ""
			remaining := limit
			for {
				pageSize := remaining
				result, err := state.SDK.ListMeetings(context.Background(), token, larksdk.ListMeetingsRequest{
					PageSize:  pageSize,
					PageToken: pageToken,
				})
				if err != nil {
					return err
				}
				meetings = append(meetings, result.Items...)
				if len(meetings) >= limit || !result.HasMore {
					break
				}
				remaining = limit - len(meetings)
				pageToken = result.PageToken
				if pageToken == "" || remaining <= 0 {
					break
				}
			}
			if len(meetings) > limit {
				meetings = meetings[:limit]
			}
			payload := map[string]any{"meetings": meetings}
			lines := make([]string, 0, len(meetings))
			for _, meeting := range meetings {
				status := ""
				if meeting.Status != nil {
					status = fmt.Sprintf("%d", *meeting.Status)
				}
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", meeting.ID, meeting.Topic, status, meeting.StartTime, meeting.EndTime))
			}
			text := tableText([]string{"meeting_id", "topic", "status", "start_time", "end_time"}, lines, "no meetings found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "max number of meetings to return")
	return cmd
}
