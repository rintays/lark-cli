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
			req := larksdk.GetCalendarEventRequest{
				CalendarID: resolvedCalendarID,
				EventID:    eventID,
				UserIDType: userIDType,
			}
			needMeeting := needMeetingSettings
			needAttend := needAttendee
			req.NeedMeetingSettings = &needMeeting
			req.NeedAttendee = &needAttend
			if maxAttendeeNum > 0 {
				value := maxAttendeeNum
				req.MaxAttendeeNum = &value
			}
			event, err := state.SDK.GetCalendarEvent(context.Background(), token, req)
			if err != nil {
				return err
			}
			payload := map[string]any{
				"calendar_id": resolvedCalendarID,
				"event":       event,
			}
			rows := calendarEventDetailRows(event)
			text := tableTextFromRows([]string{"name", "value"}, rows, "no event found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar-id", "", "calendar ID (default: primary)")
	cmd.Flags().StringVar(&eventID, "event-id", "", "event ID")
	cmd.Flags().BoolVar(&needMeetingSettings, "need-meeting-settings", true, "include VC meeting settings")
	cmd.Flags().BoolVar(&needAttendee, "need-attendee", true, "include attendee details")
	cmd.Flags().IntVar(&maxAttendeeNum, "max-attendee-num", 10, "max attendee entries to return")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user id type (open_id|union_id|user_id)")
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

func calendarEventDetailRows(event larksdk.CalendarEvent) [][]string {
	rows := make([][]string, 0, 40)
	add := func(name, value string) {
		if strings.TrimSpace(value) == "" {
			value = "-"
		}
		rows = append(rows, []string{name, value})
	}

	add("event_id", event.EventID)
	add("organizer_calendar_id", event.OrganizerCalendarID)
	add("summary", event.Summary)
	add("description", event.Description)
	add("status", event.Status)

	add("start_time", formatEventTime(event.StartTime))
	add("start_time.date", event.StartTime.Date)
	add("start_time.timestamp", event.StartTime.Timestamp)
	add("start_time.timezone", event.StartTime.Timezone)

	add("end_time", formatEventTime(event.EndTime))
	add("end_time.date", event.EndTime.Date)
	add("end_time.timestamp", event.EndTime.Timestamp)
	add("end_time.timezone", event.EndTime.Timezone)

	if event.VChat == nil {
		add("vchat.vc_type", "")
		add("vchat.icon_type", "")
		add("vchat.description", "")
		add("vchat.meeting_url", "")
		add("vchat.meeting_settings.owner_id", "")
		add("vchat.meeting_settings.join_meeting_permission", "")
		add("vchat.meeting_settings.password", "")
		add("vchat.meeting_settings.assign_hosts", "")
		add("vchat.meeting_settings.auto_record", "")
		add("vchat.meeting_settings.open_lobby", "")
		add("vchat.meeting_settings.allow_attendees_start", "")
	} else {
		add("vchat.vc_type", event.VChat.VCType)
		add("vchat.icon_type", event.VChat.IconType)
		add("vchat.description", event.VChat.Description)
		add("vchat.meeting_url", event.VChat.MeetingURL)
		settings := event.VChat.MeetingSettings
		if settings == nil {
			add("vchat.meeting_settings.owner_id", "")
			add("vchat.meeting_settings.join_meeting_permission", "")
			add("vchat.meeting_settings.password", "")
			add("vchat.meeting_settings.assign_hosts", "")
			add("vchat.meeting_settings.auto_record", "")
			add("vchat.meeting_settings.open_lobby", "")
			add("vchat.meeting_settings.allow_attendees_start", "")
		} else {
			add("vchat.meeting_settings.owner_id", settings.OwnerID)
			add("vchat.meeting_settings.join_meeting_permission", settings.JoinMeetingPermission)
			add("vchat.meeting_settings.password", settings.Password)
			add("vchat.meeting_settings.assign_hosts", formatStringSlice(settings.AssignHosts))
			add("vchat.meeting_settings.auto_record", formatBoolPtr(settings.AutoRecord))
			add("vchat.meeting_settings.open_lobby", formatBoolPtr(settings.OpenLobby))
			add("vchat.meeting_settings.allow_attendees_start", formatBoolPtr(settings.AllowAttendeesStart))
		}
	}

	add("visibility", event.Visibility)
	add("attendee_ability", event.AttendeeAbility)
	add("free_busy_status", event.FreeBusyStatus)

	if event.Location == nil {
		add("location.name", "")
		add("location.address", "")
		add("location.latitude", "")
		add("location.longitude", "")
	} else {
		add("location.name", event.Location.Name)
		add("location.address", event.Location.Address)
		add("location.latitude", formatFloat(event.Location.Latitude))
		add("location.longitude", formatFloat(event.Location.Longitude))
	}

	add("color", formatIntPtr(event.Color))
	add("reminders", formatReminders(event.Reminders))
	add("recurrence", event.Recurrence)
	add("is_exception", formatBoolPtr(event.IsException))
	add("recurring_event_id", event.RecurringEventID)
	add("create_time", formatEpoch(event.CreateTime))
	add("app_link", event.AppLink)

	if event.EventOrganizer == nil {
		add("event_organizer.user_id", "")
		add("event_organizer.display_name", "")
	} else {
		add("event_organizer.user_id", event.EventOrganizer.UserID)
		add("event_organizer.display_name", event.EventOrganizer.DisplayName)
	}

	add("schemas", formatSchemas(event.Schemas))
	add("attendees", formatAttendees(event.Attendees))
	add("has_more_attendee", formatBoolPtr(event.HasMoreAttendee))
	add("attachments", formatAttachments(event.Attachments))

	if event.EventCheckIn == nil {
		add("event_check_in.enable_check_in", "")
		add("event_check_in.check_in_start_time.time_type", "")
		add("event_check_in.check_in_start_time.duration", "")
		add("event_check_in.check_in_end_time.time_type", "")
		add("event_check_in.check_in_end_time.duration", "")
		add("event_check_in.need_notify_attendees", "")
	} else {
		add("event_check_in.enable_check_in", formatBoolPtr(event.EventCheckIn.EnableCheckIn))
		if event.EventCheckIn.CheckInStartTime == nil {
			add("event_check_in.check_in_start_time.time_type", "")
			add("event_check_in.check_in_start_time.duration", "")
		} else {
			add("event_check_in.check_in_start_time.time_type", event.EventCheckIn.CheckInStartTime.TimeType)
			add("event_check_in.check_in_start_time.duration", formatIntPtr(event.EventCheckIn.CheckInStartTime.Duration))
		}
		if event.EventCheckIn.CheckInEndTime == nil {
			add("event_check_in.check_in_end_time.time_type", "")
			add("event_check_in.check_in_end_time.duration", "")
		} else {
			add("event_check_in.check_in_end_time.time_type", event.EventCheckIn.CheckInEndTime.TimeType)
			add("event_check_in.check_in_end_time.duration", formatIntPtr(event.EventCheckIn.CheckInEndTime.Duration))
		}
		add("event_check_in.need_notify_attendees", formatBoolPtr(event.EventCheckIn.NeedNotifyAttendees))
	}

	return rows
}

func formatBoolPtr(value *bool) string {
	if value == nil {
		return ""
	}
	return strconv.FormatBool(*value)
}

func formatIntPtr(value *int) string {
	if value == nil {
		return ""
	}
	return strconv.Itoa(*value)
}

func formatStringSlice(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.Join(values, ",")
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func formatEpoch(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return value
	}
	return fmt.Sprintf("%s (%s)", value, time.Unix(seconds, 0).UTC().Format(time.RFC3339))
}

func formatReminders(reminders []larksdk.CalendarEventReminder) string {
	if len(reminders) == 0 {
		return ""
	}
	values := make([]string, 0, len(reminders))
	for _, reminder := range reminders {
		values = append(values, strconv.Itoa(reminder.Minutes))
	}
	return strings.Join(values, ",")
}

func formatSchemas(schemas []larksdk.CalendarEventSchema) string {
	if len(schemas) == 0 {
		return ""
	}
	values := make([]string, 0, len(schemas))
	for _, schema := range schemas {
		parts := []string{
			"ui_name=" + schema.UIName,
			"ui_status=" + schema.UIStatus,
			"app_link=" + schema.AppLink,
		}
		values = append(values, strings.Join(parts, " "))
	}
	return strings.Join(values, " | ")
}

func formatAttachments(attachments []larksdk.CalendarEventAttachment) string {
	if len(attachments) == 0 {
		return ""
	}
	values := make([]string, 0, len(attachments))
	for _, attachment := range attachments {
		parts := []string{
			"file_token=" + attachment.FileToken,
			"name=" + attachment.Name,
			"file_size=" + attachment.FileSize,
		}
		if attachment.IsDeleted != nil {
			parts = append(parts, "is_deleted="+strconv.FormatBool(*attachment.IsDeleted))
		}
		values = append(values, strings.Join(parts, " "))
	}
	return strings.Join(values, " | ")
}

func formatAttendees(attendees []larksdk.CalendarEventAttendee) string {
	if len(attendees) == 0 {
		return ""
	}
	values := make([]string, 0, len(attendees))
	for _, attendee := range attendees {
		parts := make([]string, 0, 10)
		parts = append(parts, "type="+attendee.Type)
		if attendee.AttendeeID != "" {
			parts = append(parts, "attendee_id="+attendee.AttendeeID)
		}
		if attendee.UserID != "" {
			parts = append(parts, "user_id="+attendee.UserID)
		}
		if attendee.ChatID != "" {
			parts = append(parts, "chat_id="+attendee.ChatID)
		}
		if attendee.RoomID != "" {
			parts = append(parts, "room_id="+attendee.RoomID)
		}
		if attendee.ThirdPartyEmail != "" {
			parts = append(parts, "third_party_email="+attendee.ThirdPartyEmail)
		}
		if attendee.OperateID != "" {
			parts = append(parts, "operate_id="+attendee.OperateID)
		}
		if attendee.RsvpStatus != "" {
			parts = append(parts, "rsvp_status="+attendee.RsvpStatus)
		}
		if attendee.IsOptional != nil {
			parts = append(parts, "is_optional="+strconv.FormatBool(*attendee.IsOptional))
		}
		if attendee.IsOrganizer != nil {
			parts = append(parts, "is_organizer="+strconv.FormatBool(*attendee.IsOrganizer))
		}
		if attendee.IsExternal != nil {
			parts = append(parts, "is_external="+strconv.FormatBool(*attendee.IsExternal))
		}
		if attendee.DisplayName != "" {
			parts = append(parts, "display_name="+attendee.DisplayName)
		}
		if len(attendee.ChatMembers) > 0 {
			parts = append(parts, "chat_members="+formatAttendeeChatMembers(attendee.ChatMembers))
		}
		values = append(values, strings.Join(parts, " "))
	}
	return strings.Join(values, " | ")
}

func formatAttendeeChatMembers(members []larksdk.CalendarEventAttendeeChatMember) string {
	values := make([]string, 0, len(members))
	for _, member := range members {
		parts := make([]string, 0, 6)
		if member.UserID != "" {
			parts = append(parts, "user_id="+member.UserID)
		}
		if member.DisplayName != "" {
			parts = append(parts, "display_name="+member.DisplayName)
		}
		if member.RsvpStatus != "" {
			parts = append(parts, "rsvp_status="+member.RsvpStatus)
		}
		if member.IsOptional != nil {
			parts = append(parts, "is_optional="+strconv.FormatBool(*member.IsOptional))
		}
		if member.IsOrganizer != nil {
			parts = append(parts, "is_organizer="+strconv.FormatBool(*member.IsOrganizer))
		}
		if member.IsExternal != nil {
			parts = append(parts, "is_external="+strconv.FormatBool(*member.IsExternal))
		}
		values = append(values, "{"+strings.Join(parts, " ")+"}")
	}
	return "[" + strings.Join(values, ", ") + "]"
}
