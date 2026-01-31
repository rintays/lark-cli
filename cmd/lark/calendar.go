package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newCalendarCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calendar",
		Short: "Manage calendar events",
	}
	cmd.AddCommand(newCalendarListCmd(state))
	cmd.AddCommand(newCalendarCreateCmd(state))
	return cmd
}

func newCalendarListCmd(state *appState) *cobra.Command {
	var start string
	var end string
	var calendarID string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List events in a time range",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				return errors.New("limit must be greater than 0")
			}
			startTime, err := time.Parse(time.RFC3339, start)
			if err != nil {
				return fmt.Errorf("invalid start time: %w", err)
			}
			endTime, err := time.Parse(time.RFC3339, end)
			if err != nil {
				return fmt.Errorf("invalid end time: %w", err)
			}
			if !endTime.After(startTime) {
				return errors.New("end time must be after start time")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			resolvedCalendarID, err := resolveCalendarID(context.Background(), state, token, calendarID)
			if err != nil {
				return err
			}
			result, err := state.SDK.ListCalendarEvents(context.Background(), token, larksdk.ListCalendarEventsRequest{
				CalendarID: resolvedCalendarID,
				StartTime:  strconv.FormatInt(startTime.Unix(), 10),
				EndTime:    strconv.FormatInt(endTime.Unix(), 10),
				PageSize:   limit,
			})
			if err != nil {
				return err
			}
			events := result.Items
			if len(events) > limit {
				events = events[:limit]
			}
			payload := map[string]any{
				"calendar_id": resolvedCalendarID,
				"events":      events,
			}
			lines := make([]string, 0, len(events))
			for _, event := range events {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s", event.EventID, formatEventTime(event.StartTime), formatEventTime(event.EndTime), event.Summary))
			}
			text := tableText([]string{"event_id", "start_time", "end_time", "summary"}, lines, "no events found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&start, "start", "", "start time (RFC3339)")
	cmd.Flags().StringVar(&end, "end", "", "end time (RFC3339)")
	cmd.Flags().StringVar(&calendarID, "calendar-id", "", "calendar ID (default: primary)")
	cmd.Flags().IntVar(&limit, "limit", 50, "max number of events to return")
	_ = cmd.MarkFlagRequired("start")
	_ = cmd.MarkFlagRequired("end")

	return cmd
}

func newCalendarCreateCmd(state *appState) *cobra.Command {
	var start string
	var end string
	var calendarID string
	var summary string
	var description string
	var attendees []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a calendar event",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			startTime, err := time.Parse(time.RFC3339, start)
			if err != nil {
				return fmt.Errorf("invalid start time: %w", err)
			}
			endTime, err := time.Parse(time.RFC3339, end)
			if err != nil {
				return fmt.Errorf("invalid end time: %w", err)
			}
			if !endTime.After(startTime) {
				return errors.New("end time must be after start time")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			resolvedCalendarID, err := resolveCalendarID(context.Background(), state, token, calendarID)
			if err != nil {
				return err
			}
			event, err := state.SDK.CreateCalendarEvent(context.Background(), token, larksdk.CreateCalendarEventRequest{
				CalendarID:  resolvedCalendarID,
				Summary:     summary,
				Description: description,
				StartTime:   startTime.Unix(),
				EndTime:     endTime.Unix(),
			})
			if err != nil {
				return err
			}
			attendeeRecords := make([]larksdk.CalendarEventAttendee, 0, len(attendees))
			for _, email := range attendees {
				if email == "" {
					continue
				}
				attendeeRecords = append(attendeeRecords, larksdk.CalendarEventAttendee{
					Type:            "third_party",
					ThirdPartyEmail: email,
				})
			}
			if len(attendeeRecords) > 0 {
				if err := state.SDK.CreateCalendarEventAttendees(context.Background(), token, larksdk.CreateCalendarEventAttendeesRequest{
					CalendarID: resolvedCalendarID,
					EventID:    event.EventID,
					Attendees:  attendeeRecords,
				}); err != nil {
					return err
				}
			}
			payload := map[string]any{
				"calendar_id": resolvedCalendarID,
				"event":       event,
				"attendees":   attendees,
			}
			text := tableTextRow(
				[]string{"event_id", "start_time", "end_time", "summary"},
				[]string{event.EventID, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339), event.Summary},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar-id", "", "calendar ID (default: primary)")
	cmd.Flags().StringVar(&start, "start", "", "start time (RFC3339)")
	cmd.Flags().StringVar(&end, "end", "", "end time (RFC3339)")
	cmd.Flags().StringVar(&summary, "summary", "", "event summary")
	cmd.Flags().StringVar(&description, "description", "", "event description")
	cmd.Flags().StringArrayVar(&attendees, "attendee", nil, "attendee email (repeatable)")
	_ = cmd.MarkFlagRequired("summary")
	_ = cmd.MarkFlagRequired("start")
	_ = cmd.MarkFlagRequired("end")

	return cmd
}

func resolveCalendarID(ctx context.Context, state *appState, token, calendarID string) (string, error) {
	if calendarID != "" {
		return calendarID, nil
	}
	if state.SDK == nil {
		return "", errors.New("sdk client is required")
	}
	cal, err := state.SDK.PrimaryCalendar(ctx, token)
	if err != nil {
		return "", err
	}
	if cal.CalendarID == "" {
		return "", errors.New("primary calendar id is empty")
	}
	return cal.CalendarID, nil
}

func formatEventTime(eventTime larksdk.CalendarEventTime) string {
	if eventTime.Timestamp != "" {
		seconds, err := strconv.ParseInt(eventTime.Timestamp, 10, 64)
		if err != nil {
			return eventTime.Timestamp
		}
		return time.Unix(seconds, 0).UTC().Format(time.RFC3339)
	}
	if eventTime.Date != "" {
		return eventTime.Date
	}
	return ""
}
