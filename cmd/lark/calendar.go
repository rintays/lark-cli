package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"lark/internal/larkapi"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			if start == "" || end == "" {
				return errors.New("start and end are required")
			}
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
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			resolvedCalendarID, err := resolveCalendarID(context.Background(), state, token, calendarID)
			if err != nil {
				return err
			}
			result, err := state.Client.ListCalendarEvents(context.Background(), token, larkapi.ListCalendarEventsRequest{
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
			text := "no events found"
			if len(lines) > 0 {
				text = strings.Join(lines, "\n")
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&start, "start", "", "start time (RFC3339)")
	cmd.Flags().StringVar(&end, "end", "", "end time (RFC3339)")
	cmd.Flags().StringVar(&calendarID, "calendar-id", "", "calendar ID (default: primary)")
	cmd.Flags().IntVar(&limit, "limit", 50, "max number of events to return")

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
		RunE: func(cmd *cobra.Command, args []string) error {
			if summary == "" {
				return errors.New("summary is required")
			}
			if start == "" || end == "" {
				return errors.New("start and end are required")
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
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			resolvedCalendarID, err := resolveCalendarID(context.Background(), state, token, calendarID)
			if err != nil {
				return err
			}
			event, err := state.Client.CreateCalendarEvent(context.Background(), token, larkapi.CreateCalendarEventRequest{
				CalendarID:  resolvedCalendarID,
				Summary:     summary,
				Description: description,
				StartTime:   startTime.Unix(),
				EndTime:     endTime.Unix(),
			})
			if err != nil {
				return err
			}
			attendeeRecords := make([]larkapi.CalendarEventAttendee, 0, len(attendees))
			for _, email := range attendees {
				if email == "" {
					continue
				}
				attendeeRecords = append(attendeeRecords, larkapi.CalendarEventAttendee{
					Type:            "third_party",
					ThirdPartyEmail: email,
				})
			}
			if len(attendeeRecords) > 0 {
				if err := state.Client.CreateCalendarEventAttendees(context.Background(), token, larkapi.CreateCalendarEventAttendeesRequest{
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
			text := fmt.Sprintf("%s\t%s\t%s\t%s", event.EventID, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339), event.Summary)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar-id", "", "calendar ID (default: primary)")
	cmd.Flags().StringVar(&start, "start", "", "start time (RFC3339)")
	cmd.Flags().StringVar(&end, "end", "", "end time (RFC3339)")
	cmd.Flags().StringVar(&summary, "summary", "", "event summary")
	cmd.Flags().StringVar(&description, "description", "", "event description")
	cmd.Flags().StringArrayVar(&attendees, "attendee", nil, "attendee email (repeatable)")

	return cmd
}

func resolveCalendarID(ctx context.Context, state *appState, token, calendarID string) (string, error) {
	if calendarID != "" {
		return calendarID, nil
	}
	cal, err := state.Client.PrimaryCalendar(ctx, token)
	if err != nil {
		return "", err
	}
	if cal.CalendarID == "" {
		return "", errors.New("primary calendar id is empty")
	}
	return cal.CalendarID, nil
}

func formatEventTime(eventTime larkapi.CalendarEventTime) string {
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
