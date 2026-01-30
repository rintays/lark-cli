package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"lark/internal/larkapi"
)

func newMeetingCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meeting",
		Short: "Get meeting details",
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
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				if meetingID != "" && meetingID != args[0] {
					return errors.New("meeting-id provided twice")
				}
				meetingID = args[0]
			}
			if meetingID == "" {
				return errors.New("meeting-id is required")
			}
			if queryMode < 0 || queryMode > 1 {
				return errors.New("query-mode must be 0 or 1")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			meeting, err := state.Client.GetMeeting(context.Background(), token, larkapi.GetMeetingRequest{
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
	return cmd
}
