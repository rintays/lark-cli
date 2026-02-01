package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newCalendarCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "calendars",
		Aliases: []string{"calendar"},
		Short:   "Manage calendars (events)",
		Long: `Calendars contain events scheduled for users.

- calendar_id identifies a calendar (default: primary).
- Events have event_id plus start/end times.
- list/search operate on time ranges; create/update manage event details.

Canonical command name: calendars (alias: calendar).`,
	}
	cmd.AddCommand(newCalendarListCmd(state))
	cmd.AddCommand(newCalendarCreateCmd(state))
	cmd.AddCommand(newCalendarSearchCmd(state))
	cmd.AddCommand(newCalendarGetCmd(state))
	cmd.AddCommand(newCalendarUpdateCmd(state))
	cmd.AddCommand(newCalendarDeleteCmd(state))
	return cmd
}

func newCalendarListCmd(state *appState) *cobra.Command {
	var start string
	var end string
	var calendarID string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List events",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				return errors.New("limit must be greater than 0")
			}
			var startTime time.Time
			var endTime time.Time
			if start != "" || end != "" {
				if start == "" || end == "" {
					return errors.New("start and end must be provided together")
				}
				parsedStart, err := time.Parse(time.RFC3339, start)
				if err != nil {
					return fmt.Errorf("invalid start time: %w", err)
				}
				parsedEnd, err := time.Parse(time.RFC3339, end)
				if err != nil {
					return fmt.Errorf("invalid end time: %w", err)
				}
				if !parsedEnd.After(parsedStart) {
					return errors.New("end time must be after start time")
				}
				startTime = parsedStart
				endTime = parsedEnd
			}
			token, tokenType, err := resolveAccessToken(context.Background(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			resolvedCalendarID, err := resolveCalendarID(context.Background(), state, token, tokenType, calendarID)
			if err != nil {
				return err
			}
			events := make([]larksdk.CalendarEvent, 0, limit)
			pageToken := ""
			remaining := limit
			for {
				pageSize := remaining
				req := larksdk.ListCalendarEventsRequest{
					CalendarID: resolvedCalendarID,
					PageSize:   pageSize,
					PageToken:  pageToken,
				}
				if start != "" {
					req.StartTime = strconv.FormatInt(startTime.Unix(), 10)
					req.EndTime = strconv.FormatInt(endTime.Unix(), 10)
				}
				result, err := state.SDK.ListCalendarEvents(context.Background(), token, req)
				if err != nil {
					return err
				}
				events = append(events, result.Items...)
				if len(events) >= limit || !result.HasMore {
					break
				}
				remaining = limit - len(events)
				if remaining <= 0 || result.PageToken == "" {
					break
				}
				pageToken = result.PageToken
			}
			if len(events) > limit {
				events = events[:limit]
			}
			payload := map[string]any{
				"calendar_id": resolvedCalendarID,
				"events":      events,
			}
			lines := make([]string, 0, len(events))
			for _, event := range events {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", event.EventID, formatEventTime(event.StartTime), formatEventTime(event.EndTime), event.Summary, event.Status))
			}
			text := tableText([]string{"event_id", "start_time", "end_time", "summary", "status"}, lines, "no events found")
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
	var bodyJSON string
	var bodyFile string
	var userIDType string
	var idempotencyKey string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a calendar event",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			extra, err := parseCalendarExtra(bodyJSON, bodyFile)
			if err != nil {
				return err
			}
			if summary == "" && (extra == nil || extra["summary"] == nil) {
				return errors.New("summary is required")
			}
			if start != "" || end != "" {
				if start == "" || end == "" {
					return errors.New("start and end times are required")
				}
			} else if extra == nil {
				return errors.New("start and end times are required")
			} else {
				startExtra := extra["start_time"] != nil
				endExtra := extra["end_time"] != nil
				if startExtra != endExtra {
					return errors.New("start_time and end_time must be provided together")
				}
			}
			var startTime time.Time
			var endTime time.Time
			if start != "" {
				startTime, err = time.Parse(time.RFC3339, start)
				if err != nil {
					return fmt.Errorf("invalid start time: %w", err)
				}
				endTime, err = time.Parse(time.RFC3339, end)
				if err != nil {
					return fmt.Errorf("invalid end time: %w", err)
				}
				if !endTime.After(startTime) {
					return errors.New("end time must be after start time")
				}
			}
			token, tokenType, err := resolveAccessToken(context.Background(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			resolvedCalendarID, err := resolveCalendarID(context.Background(), state, token, tokenType, calendarID)
			if err != nil {
				return err
			}
			req := larksdk.CreateCalendarEventRequest{
				CalendarID:     resolvedCalendarID,
				Summary:        summary,
				Description:    description,
				IdempotencyKey: idempotencyKey,
				UserIDType:     userIDType,
				Extra:          extra,
			}
			if start != "" {
				req.StartTime = startTime.Unix()
				req.EndTime = endTime.Unix()
			}
			event, err := state.SDK.CreateCalendarEvent(context.Background(), token, req)
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
			startText := formatEventTime(event.StartTime)
			if startText == "" && !startTime.IsZero() {
				startText = startTime.Format(time.RFC3339)
			}
			endText := formatEventTime(event.EndTime)
			if endText == "" && !endTime.IsZero() {
				endText = endTime.Format(time.RFC3339)
			}
			text := tableTextRow(
				[]string{"event_id", "start_time", "end_time", "summary", "status"},
				[]string{event.EventID, startText, endText, event.Summary, event.Status},
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
	cmd.Flags().StringVar(&bodyJSON, "body-json", "", "JSON object of additional event fields")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "path to JSON file of additional event fields")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user id type (open_id|union_id|user_id)")
	cmd.Flags().StringVar(&idempotencyKey, "idempotency-key", "", "idempotency key for event creation")
	cmd.MarkFlagsMutuallyExclusive("body-json", "body-file")

	return cmd
}

func newCalendarSearchCmd(state *appState) *cobra.Command {
	var query string
	var start string
	var end string
	var calendarID string
	var limit int
	var userIDs []string
	var roomIDs []string
	var chatIDs []string

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search events in a calendar",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				return errors.New("limit must be greater than 0")
			}
			var startTime time.Time
			var endTime time.Time
			if start != "" || end != "" {
				if start == "" || end == "" {
					return errors.New("start and end must be provided together")
				}
				parsedStart, err := time.Parse(time.RFC3339, start)
				if err != nil {
					return fmt.Errorf("invalid start time: %w", err)
				}
				parsedEnd, err := time.Parse(time.RFC3339, end)
				if err != nil {
					return fmt.Errorf("invalid end time: %w", err)
				}
				if !parsedEnd.After(parsedStart) {
					return errors.New("end time must be after start time")
				}
				startTime = parsedStart
				endTime = parsedEnd
			}

			token, tokenType, err := resolveAccessToken(context.Background(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			resolvedCalendarID, err := resolveCalendarID(context.Background(), state, token, tokenType, calendarID)
			if err != nil {
				return err
			}
			events := make([]larksdk.CalendarEvent, 0, limit)
			pageToken := ""
			remaining := limit
			for {
				pageSize := remaining
				req := larksdk.SearchCalendarEventsRequest{
					CalendarID: resolvedCalendarID,
					Query:      query,
					UserIDs:    userIDs,
					RoomIDs:    roomIDs,
					ChatIDs:    chatIDs,
					PageSize:   pageSize,
					PageToken:  pageToken,
				}
				if start != "" {
					req.StartTime = strconv.FormatInt(startTime.Unix(), 10)
					req.EndTime = strconv.FormatInt(endTime.Unix(), 10)
				}
				result, err := state.SDK.SearchCalendarEvents(context.Background(), token, req)
				if err != nil {
					return err
				}
				events = append(events, result.Items...)
				if len(events) >= limit || result.PageToken == "" || result.PageToken == pageToken {
					pageToken = result.PageToken
					break
				}
				remaining = limit - len(events)
				if remaining <= 0 {
					pageToken = result.PageToken
					break
				}
				pageToken = result.PageToken
			}
			if len(events) > limit {
				events = events[:limit]
			}
			payload := map[string]any{
				"calendar_id": resolvedCalendarID,
				"events":      events,
				"page_token":  pageToken,
			}
			lines := make([]string, 0, len(events))
			for _, event := range events {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", event.EventID, formatEventTime(event.StartTime), formatEventTime(event.EndTime), event.Summary, event.Status))
			}
			text := tableText([]string{"event_id", "start_time", "end_time", "summary", "status"}, lines, "no events found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "search query")
	cmd.Flags().StringVar(&start, "start", "", "start time (RFC3339)")
	cmd.Flags().StringVar(&end, "end", "", "end time (RFC3339)")
	cmd.Flags().StringVar(&calendarID, "calendar-id", "", "calendar ID (default: primary)")
	cmd.Flags().IntVar(&limit, "limit", 20, "max number of events to return")
	cmd.Flags().StringArrayVar(&userIDs, "user-id", nil, "filter by attendee user id (repeatable)")
	cmd.Flags().StringArrayVar(&roomIDs, "room-id", nil, "filter by room id (repeatable)")
	cmd.Flags().StringArrayVar(&chatIDs, "chat-id", nil, "filter by chat id (repeatable)")
	_ = cmd.MarkFlagRequired("query")

	return cmd
}

func newCalendarGetCmd(state *appState) *cobra.Command {
	var calendarID string
	var eventID string
	var needMeetingSettings bool
	var needAttendee bool
	var maxAttendeeNum int
	var userIDType string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get calendar event details",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			token, tokenType, err := resolveAccessToken(context.Background(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			resolvedCalendarID, err := resolveCalendarID(context.Background(), state, token, tokenType, calendarID)
			if err != nil {
				return err
			}
			attendeeFlagChanged := cmd.Flags().Changed("need-attendee")
			meetingFlagChanged := cmd.Flags().Changed("need-meeting-settings")
			maxAttendeeChanged := cmd.Flags().Changed("max-attendee-num")
			req := larksdk.GetCalendarEventRequest{
				CalendarID: resolvedCalendarID,
				EventID:    eventID,
				UserIDType: userIDType,
			}
			if meetingFlagChanged || needMeetingSettings {
				value := needMeetingSettings
				req.NeedMeetingSettings = &value
			}
			if attendeeFlagChanged || needAttendee {
				value := needAttendee
				req.NeedAttendee = &value
				if value && maxAttendeeNum > 0 {
					req.MaxAttendeeNum = &maxAttendeeNum
				} else if maxAttendeeChanged {
					valueNum := maxAttendeeNum
					req.MaxAttendeeNum = &valueNum
				}
			}
			event, err := state.SDK.GetCalendarEvent(context.Background(), token, req)
			var extraErr error
			if err != nil && (needAttendee || needMeetingSettings) && !attendeeFlagChanged && !meetingFlagChanged {
				extraErr = err
				fallbackReq := larksdk.GetCalendarEventRequest{
					CalendarID: resolvedCalendarID,
					EventID:    eventID,
					UserIDType: userIDType,
				}
				event, err = state.SDK.GetCalendarEvent(context.Background(), token, fallbackReq)
			}
			if err != nil {
				return err
			}
			payload := map[string]any{
				"calendar_id": resolvedCalendarID,
				"event":       event,
			}
			if extraErr != nil {
				payload["extra_error"] = extraErr.Error()
			}
			text := formatCalendarEventInfo(event, extraErr)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar-id", "", "calendar ID (default: primary)")
	cmd.Flags().StringVar(&eventID, "event-id", "", "event ID")
	cmd.Flags().BoolVar(&needAttendee, "need-attendee", true, "include attendee info (requires permission)")
	cmd.Flags().BoolVar(&needMeetingSettings, "need-meeting-settings", true, "include meeting settings for VC events")
	cmd.Flags().IntVar(&maxAttendeeNum, "max-attendee-num", 100, "max number of attendees to return (only when --need-attendee)")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "open_id", "user ID type (open_id, union_id, user_id)")
	_ = cmd.MarkFlagRequired("event-id")

	return cmd
}

func newCalendarUpdateCmd(state *appState) *cobra.Command {
	var calendarID string
	var eventID string
	var summary string
	var description string
	var start string
	var end string
	var bodyJSON string
	var bodyFile string
	var userIDType string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a calendar event",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			extra, err := parseCalendarExtra(bodyJSON, bodyFile)
			if err != nil {
				return err
			}
			if summary == "" && description == "" && start == "" && end == "" && extra == nil {
				return errors.New("at least one field must be provided")
			}
			var startTime time.Time
			var endTime time.Time
			if start != "" || end != "" {
				if start == "" || end == "" {
					return errors.New("start and end must be provided together")
				}
				parsedStart, err := time.Parse(time.RFC3339, start)
				if err != nil {
					return fmt.Errorf("invalid start time: %w", err)
				}
				parsedEnd, err := time.Parse(time.RFC3339, end)
				if err != nil {
					return fmt.Errorf("invalid end time: %w", err)
				}
				if !parsedEnd.After(parsedStart) {
					return errors.New("end time must be after start time")
				}
				startTime = parsedStart
				endTime = parsedEnd
			} else if extra != nil {
				startExtra := extra["start_time"] != nil
				endExtra := extra["end_time"] != nil
				if startExtra != endExtra {
					return errors.New("start_time and end_time must be provided together")
				}
			}

			token, tokenType, err := resolveAccessToken(context.Background(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			resolvedCalendarID, err := resolveCalendarID(context.Background(), state, token, tokenType, calendarID)
			if err != nil {
				return err
			}
			req := larksdk.UpdateCalendarEventRequest{
				CalendarID:  resolvedCalendarID,
				EventID:     eventID,
				Summary:     summary,
				Description: description,
				UserIDType:  userIDType,
				Extra:       extra,
			}
			if start != "" {
				startUnix := startTime.Unix()
				endUnix := endTime.Unix()
				req.StartTime = &startUnix
				req.EndTime = &endUnix
			}
			event, err := state.SDK.UpdateCalendarEvent(context.Background(), token, req)
			if err != nil {
				return err
			}
			payload := map[string]any{
				"calendar_id": resolvedCalendarID,
				"event":       event,
			}
			text := tableTextRow(
				[]string{"event_id", "start_time", "end_time", "summary", "status"},
				[]string{event.EventID, formatEventTime(event.StartTime), formatEventTime(event.EndTime), event.Summary, event.Status},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar-id", "", "calendar ID (default: primary)")
	cmd.Flags().StringVar(&eventID, "event-id", "", "event ID")
	cmd.Flags().StringVar(&summary, "summary", "", "event summary")
	cmd.Flags().StringVar(&description, "description", "", "event description")
	cmd.Flags().StringVar(&start, "start", "", "start time (RFC3339)")
	cmd.Flags().StringVar(&end, "end", "", "end time (RFC3339)")
	cmd.Flags().StringVar(&bodyJSON, "body-json", "", "JSON object of additional event fields")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "path to JSON file of additional event fields")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user id type (open_id|union_id|user_id)")
	cmd.MarkFlagsMutuallyExclusive("body-json", "body-file")
	_ = cmd.MarkFlagRequired("event-id")

	return cmd
}

func newCalendarDeleteCmd(state *appState) *cobra.Command {
	var calendarID string
	var eventID string
	var notify bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a calendar event",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			token, tokenType, err := resolveAccessToken(context.Background(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			resolvedCalendarID, err := resolveCalendarID(context.Background(), state, token, tokenType, calendarID)
			if err != nil {
				return err
			}
			result, err := state.SDK.DeleteCalendarEvent(context.Background(), token, larksdk.DeleteCalendarEventRequest{
				CalendarID:       resolvedCalendarID,
				EventID:          eventID,
				NeedNotification: notify,
			})
			if err != nil {
				return err
			}
			payload := map[string]any{
				"calendar_id": resolvedCalendarID,
				"event_id":    result.EventID,
				"deleted":     result.Deleted,
			}
			text := tableTextRow(
				[]string{"event_id", "deleted"},
				[]string{result.EventID, strconv.FormatBool(result.Deleted)},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar-id", "", "calendar ID (default: primary)")
	cmd.Flags().StringVar(&eventID, "event-id", "", "event ID")
	cmd.Flags().BoolVar(&notify, "notify", true, "notify attendees about deletion")
	_ = cmd.MarkFlagRequired("event-id")

	return cmd
}

func resolveCalendarID(ctx context.Context, state *appState, token string, accessType tokenType, calendarID string) (string, error) {
	if calendarID != "" {
		return calendarID, nil
	}
	if state.SDK == nil {
		return "", errors.New("sdk client is required")
	}
	var cal larksdk.Calendar
	var err error
	switch accessType {
	case tokenTypeTenant:
		cal, err = state.SDK.PrimaryCalendar(ctx, token)
	case tokenTypeUser:
		cal, err = state.SDK.PrimaryCalendarWithUserToken(ctx, token)
	default:
		return "", fmt.Errorf("unsupported token type %s", accessType)
	}
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

func formatCalendarEventInfo(event larksdk.CalendarEvent, extraErr error) string {
	rows := [][]string{
		{"event_id", infoValue(event.EventID)},
		{"organizer_calendar_id", infoValue(event.OrganizerCalendarID)},
		{"summary", infoValue(event.Summary)},
		{"description", infoValue(event.Description)},
		{"start_time", infoValue(formatEventTime(event.StartTime))},
		{"start_time.timestamp", infoValue(event.StartTime.Timestamp)},
		{"start_time.timezone", infoValue(event.StartTime.Timezone)},
		{"start_time.date", infoValue(event.StartTime.Date)},
		{"end_time", infoValue(formatEventTime(event.EndTime))},
		{"end_time.timestamp", infoValue(event.EndTime.Timestamp)},
		{"end_time.timezone", infoValue(event.EndTime.Timezone)},
		{"end_time.date", infoValue(event.EndTime.Date)},
		{"status", infoValue(event.Status)},
		{"visibility", infoValue(event.Visibility)},
		{"attendee_ability", infoValue(event.AttendeeAbility)},
		{"free_busy_status", infoValue(event.FreeBusyStatus)},
		{"recurrence", infoValue(event.Recurrence)},
		{"is_exception", infoValueBoolPtr(event.IsException)},
		{"recurring_event_id", infoValue(event.RecurringEventID)},
		{"create_time", infoValue(event.CreateTime)},
		{"color", infoValueIntPtr(event.Color)},
		{"app_link", infoValue(event.AppLink)},
	}

	if event.VChat != nil {
		rows = append(rows,
			[]string{"vchat.vc_type", infoValue(event.VChat.VCType)},
			[]string{"vchat.icon_type", infoValue(event.VChat.IconType)},
			[]string{"vchat.description", infoValue(event.VChat.Description)},
			[]string{"vchat.meeting_url", infoValue(event.VChat.MeetingURL)},
		)
		if event.VChat.MeetingSettings != nil {
			settings := event.VChat.MeetingSettings
			rows = append(rows,
				[]string{"vchat.meeting_settings.owner_id", infoValue(settings.OwnerID)},
				[]string{"vchat.meeting_settings.join_meeting_permission", infoValue(settings.JoinMeetingPermission)},
				[]string{"vchat.meeting_settings.password", infoValue(settings.Password)},
				[]string{"vchat.meeting_settings.assign_hosts.count", fmt.Sprintf("%d", len(settings.AssignHosts))},
				[]string{"vchat.meeting_settings.auto_record", infoValueBoolPtr(settings.AutoRecord)},
				[]string{"vchat.meeting_settings.open_lobby", infoValueBoolPtr(settings.OpenLobby)},
				[]string{"vchat.meeting_settings.allow_attendees_start", infoValueBoolPtr(settings.AllowAttendeesStart)},
			)
			for i, host := range settings.AssignHosts {
				rows = append(rows, []string{fmt.Sprintf("vchat.meeting_settings.assign_hosts[%d]", i), infoValue(host)})
			}
		} else {
			rows = append(rows,
				[]string{"vchat.meeting_settings.owner_id", infoValue("")},
				[]string{"vchat.meeting_settings.join_meeting_permission", infoValue("")},
				[]string{"vchat.meeting_settings.password", infoValue("")},
				[]string{"vchat.meeting_settings.assign_hosts.count", "0"},
				[]string{"vchat.meeting_settings.auto_record", infoValueBoolPtr(nil)},
				[]string{"vchat.meeting_settings.open_lobby", infoValueBoolPtr(nil)},
				[]string{"vchat.meeting_settings.allow_attendees_start", infoValueBoolPtr(nil)},
			)
		}
	} else {
		rows = append(rows,
			[]string{"vchat.vc_type", infoValue("")},
			[]string{"vchat.icon_type", infoValue("")},
			[]string{"vchat.description", infoValue("")},
			[]string{"vchat.meeting_url", infoValue("")},
			[]string{"vchat.meeting_settings.owner_id", infoValue("")},
			[]string{"vchat.meeting_settings.join_meeting_permission", infoValue("")},
			[]string{"vchat.meeting_settings.password", infoValue("")},
			[]string{"vchat.meeting_settings.assign_hosts.count", "0"},
			[]string{"vchat.meeting_settings.auto_record", infoValueBoolPtr(nil)},
			[]string{"vchat.meeting_settings.open_lobby", infoValueBoolPtr(nil)},
			[]string{"vchat.meeting_settings.allow_attendees_start", infoValueBoolPtr(nil)},
		)
	}

	if event.Location != nil {
		rows = append(rows,
			[]string{"location.name", infoValue(event.Location.Name)},
			[]string{"location.address", infoValue(event.Location.Address)},
			[]string{"location.latitude", infoValueFloatPtr(event.Location.Latitude)},
			[]string{"location.longitude", infoValueFloatPtr(event.Location.Longitude)},
		)
	} else {
		rows = append(rows,
			[]string{"location.name", infoValue("")},
			[]string{"location.address", infoValue("")},
			[]string{"location.latitude", infoValueFloatPtr(nil)},
			[]string{"location.longitude", infoValueFloatPtr(nil)},
		)
	}

	if event.EventOrganizer != nil {
		rows = append(rows,
			[]string{"event_organizer.user_id", infoValue(event.EventOrganizer.UserID)},
			[]string{"event_organizer.display_name", infoValue(event.EventOrganizer.DisplayName)},
		)
	} else {
		rows = append(rows,
			[]string{"event_organizer.user_id", infoValue("")},
			[]string{"event_organizer.display_name", infoValue("")},
		)
	}

	rows = append(rows, []string{"attendees.count", fmt.Sprintf("%d", len(event.Attendees))})
	for i, attendee := range event.Attendees {
		prefix := fmt.Sprintf("attendees[%d]", i)
		rows = append(rows,
			[]string{prefix + ".type", infoValue(attendee.Type)},
			[]string{prefix + ".attendee_id", infoValue(attendee.AttendeeID)},
			[]string{prefix + ".rsvp_status", infoValue(attendee.RsvpStatus)},
			[]string{prefix + ".is_optional", infoValueBoolPtr(attendee.IsOptional)},
			[]string{prefix + ".is_organizer", infoValueBoolPtr(attendee.IsOrganizer)},
			[]string{prefix + ".is_external", infoValueBoolPtr(attendee.IsExternal)},
			[]string{prefix + ".display_name", infoValue(attendee.DisplayName)},
			[]string{prefix + ".user_id", infoValue(attendee.UserID)},
			[]string{prefix + ".chat_id", infoValue(attendee.ChatID)},
			[]string{prefix + ".room_id", infoValue(attendee.RoomID)},
			[]string{prefix + ".third_party_email", infoValue(attendee.ThirdPartyEmail)},
			[]string{prefix + ".operate_id", infoValue(attendee.OperateID)},
			[]string{prefix + ".chat_members.count", fmt.Sprintf("%d", len(attendee.ChatMembers))},
		)
		for j, member := range attendee.ChatMembers {
			memberPrefix := fmt.Sprintf("%s.chat_members[%d]", prefix, j)
			rows = append(rows,
				[]string{memberPrefix + ".rsvp_status", infoValue(member.RsvpStatus)},
				[]string{memberPrefix + ".is_optional", infoValueBoolPtr(member.IsOptional)},
				[]string{memberPrefix + ".display_name", infoValue(member.DisplayName)},
				[]string{memberPrefix + ".is_organizer", infoValueBoolPtr(member.IsOrganizer)},
				[]string{memberPrefix + ".is_external", infoValueBoolPtr(member.IsExternal)},
				[]string{memberPrefix + ".user_id", infoValue(member.UserID)},
			)
		}
	}

	rows = append(rows,
		[]string{"has_more_attendee", infoValueBoolPtr(event.HasMoreAttendee)},
		[]string{"reminders.count", fmt.Sprintf("%d", len(event.Reminders))},
	)
	for i, reminder := range event.Reminders {
		rows = append(rows, []string{fmt.Sprintf("reminders[%d].minutes", i), fmt.Sprintf("%d", reminder.Minutes)})
	}

	rows = append(rows, []string{"attachments.count", fmt.Sprintf("%d", len(event.Attachments))})
	for i, attachment := range event.Attachments {
		prefix := fmt.Sprintf("attachments[%d]", i)
		rows = append(rows,
			[]string{prefix + ".file_token", infoValue(attachment.FileToken)},
			[]string{prefix + ".file_size", infoValue(attachment.FileSize)},
			[]string{prefix + ".name", infoValue(attachment.Name)},
			[]string{prefix + ".is_deleted", infoValueBoolPtr(attachment.IsDeleted)},
		)
	}

	if event.EventCheckIn != nil {
		rows = append(rows,
			[]string{"event_check_in.enable_check_in", infoValueBoolPtr(event.EventCheckIn.EnableCheckIn)},
			[]string{"event_check_in.need_notify_attendees", infoValueBoolPtr(event.EventCheckIn.NeedNotifyAttendees)},
		)
		if event.EventCheckIn.CheckInStartTime != nil {
			rows = append(rows,
				[]string{"event_check_in.check_in_start_time.time_type", infoValue(event.EventCheckIn.CheckInStartTime.TimeType)},
				[]string{"event_check_in.check_in_start_time.duration", infoValueIntPtr(event.EventCheckIn.CheckInStartTime.Duration)},
			)
		} else {
			rows = append(rows,
				[]string{"event_check_in.check_in_start_time.time_type", infoValue("")},
				[]string{"event_check_in.check_in_start_time.duration", infoValueIntPtr(nil)},
			)
		}
		if event.EventCheckIn.CheckInEndTime != nil {
			rows = append(rows,
				[]string{"event_check_in.check_in_end_time.time_type", infoValue(event.EventCheckIn.CheckInEndTime.TimeType)},
				[]string{"event_check_in.check_in_end_time.duration", infoValueIntPtr(event.EventCheckIn.CheckInEndTime.Duration)},
			)
		} else {
			rows = append(rows,
				[]string{"event_check_in.check_in_end_time.time_type", infoValue("")},
				[]string{"event_check_in.check_in_end_time.duration", infoValueIntPtr(nil)},
			)
		}
	} else {
		rows = append(rows,
			[]string{"event_check_in.enable_check_in", infoValueBoolPtr(nil)},
			[]string{"event_check_in.need_notify_attendees", infoValueBoolPtr(nil)},
			[]string{"event_check_in.check_in_start_time.time_type", infoValue("")},
			[]string{"event_check_in.check_in_start_time.duration", infoValueIntPtr(nil)},
			[]string{"event_check_in.check_in_end_time.time_type", infoValue("")},
			[]string{"event_check_in.check_in_end_time.duration", infoValueIntPtr(nil)},
		)
	}

	if extraErr != nil {
		rows = append(rows, []string{"extra_error", infoValue(extraErr.Error())})
	}

	for i := range rows {
		if len(rows[i]) > 1 {
			rows[i][1] = strings.TrimSpace(rows[i][1])
		}
	}
	return formatInfoTable(rows, "no event found")
}

func parseCalendarExtra(raw, path string) (map[string]any, error) {
	if strings.TrimSpace(raw) == "" && strings.TrimSpace(path) == "" {
		return nil, nil
	}
	var data []byte
	if strings.TrimSpace(path) != "" {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read body file: %w", err)
		}
		data = content
	} else {
		data = []byte(raw)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil, errors.New("body is empty")
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("body must be valid JSON object: %w", err)
	}
	if payload == nil {
		return nil, errors.New("body must be a JSON object")
	}
	return payload, nil
}
