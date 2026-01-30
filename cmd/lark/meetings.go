package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newMeetingsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meetings",
		Short: "Manage meetings",
	}
	cmd.AddCommand(newMeetingGetCmd(state))
	return cmd
}

func newMeetingGetCmd(state *appState) *cobra.Command {
	var meetingID string
	var withParticipants bool
	var withMeetingAbility bool
	var userIDType string
	var queryMode int

	cmd := &cobra.Command{
		Use:   "get <meeting-id>",
		Short: "Get meeting details",
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
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if queryMode < 0 || queryMode > 1 {
				return errors.New("query-mode must be 0 or 1")
			}
			token, err := ensureTenantToken(context.Background(), state)
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
			text := fmt.Sprintf("%s\t%s\t%d\t%s\t%s", meeting.ID, meeting.Topic, meeting.Status, meeting.StartTime, meeting.EndTime)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&meetingID, "meeting-id", "", "meeting ID (or provide as positional argument)")
	cmd.Flags().BoolVar(&withParticipants, "with-participants", false, "include participant list")
	cmd.Flags().BoolVar(&withMeetingAbility, "with-ability", false, "include meeting ability stats")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user ID type (user_id, union_id, open_id)")
	cmd.Flags().IntVar(&queryMode, "query-mode", 0, "query mode: 0 for meeting info, 1 for related artifacts")
	_ = cmd.MarkFlagRequired("meeting-id")
	return cmd
}
