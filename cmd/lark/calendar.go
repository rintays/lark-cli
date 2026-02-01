package main

import (
	"context"
	"errors"
	"fmt"
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
	var userIDType string
	var idempotencyKey string
	var needNotification bool
	var visibility string
	var attendeeAbility string
	var freeBusyStatus string
	var locationName string
	var locationAddress string
	var locationLatitude float64
	var locationLongitude float64
	var color int
	var reminders []int
	var recurrence string
	var schemaEntries []string
	var attachments []string
	var vcType string
	var vcIconType string
	var vcDescription string
	var vcMeetingURL string
	var vcOwnerID string
	var vcJoinPermission string
	var vcPassword string
	var vcAssignHosts []string
	var vcAutoRecord bool
	var vcOpenLobby bool
	var vcAllowAttendeesStart bool
	var checkInEnabled bool
	var checkInStartType string
	var checkInStartDuration int
	var checkInEndType string
	var checkInEndDuration int
	var checkInNotifyAttendees bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a calendar event",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if summary == "" {
				return errors.New("summary is required")
			}
			if start == "" || end == "" {
				return errors.New("start and end times are required")
			}
			var startTime time.Time
			var endTime time.Time
			startTime, err := time.Parse(time.RFC3339, start)
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
			schemas, err := parseCalendarSchemas(schemaEntries)
			if err != nil {
				return err
			}
			vchat, err := buildCalendarVChat(cmd, vcType, vcIconType, vcDescription, vcMeetingURL, vcOwnerID, vcJoinPermission, vcPassword, vcAssignHosts, vcAutoRecord, vcOpenLobby, vcAllowAttendeesStart)
			if err != nil {
				return err
			}
			location := buildCalendarLocation(cmd, locationName, locationAddress, locationLatitude, locationLongitude)
			checkIn, err := buildCalendarCheckIn(cmd, checkInEnabled, checkInStartType, checkInStartDuration, checkInEndType, checkInEndDuration, checkInNotifyAttendees)
			if err != nil {
				return err
			}
			reminderItems := buildCalendarReminders(reminders)
			attachmentItems := buildCalendarAttachments(attachments, nil)
			req := larksdk.CreateCalendarEventRequest{
				CalendarID:      resolvedCalendarID,
				Summary:         summary,
				Description:     description,
				IdempotencyKey:  idempotencyKey,
				UserIDType:      userIDType,
				Visibility:      visibility,
				AttendeeAbility: attendeeAbility,
				FreeBusyStatus:  freeBusyStatus,
				Location:        location,
				Color:           flagIntPtr(cmd, "color", color),
				Reminders:       reminderItems,
				Recurrence:      recurrence,
				VChat:           vchat,
				Schemas:         schemas,
				Attachments:     attachmentItems,
				EventCheckIn:    checkIn,
			}
			if needNotificationPtr := flagBoolPtr(cmd, "need-notification", needNotification); needNotificationPtr != nil {
				req.NeedNotification = needNotificationPtr
			}
			req.StartTime = startTime.Unix()
			req.EndTime = endTime.Unix()
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
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user id type (open_id|union_id|user_id)")
	cmd.Flags().StringVar(&idempotencyKey, "idempotency-key", "", "idempotency key for event creation")
	cmd.Flags().BoolVar(&needNotification, "need-notification", true, "notify attendees about event creation")
	cmd.Flags().StringVar(&visibility, "visibility", "", "visibility (default|public|private)")
	cmd.Flags().StringVar(&attendeeAbility, "attendee-ability", "", "attendee ability (none|can_see_others|can_invite_others|can_modify_event)")
	cmd.Flags().StringVar(&freeBusyStatus, "free-busy-status", "", "free/busy status (busy|free)")
	cmd.Flags().StringVar(&locationName, "location-name", "", "location name")
	cmd.Flags().StringVar(&locationAddress, "location-address", "", "location address")
	cmd.Flags().Float64Var(&locationLatitude, "location-latitude", 0, "location latitude (GCJ-02 in CN, WGS84 elsewhere)")
	cmd.Flags().Float64Var(&locationLongitude, "location-longitude", 0, "location longitude (GCJ-02 in CN, WGS84 elsewhere)")
	cmd.Flags().IntVar(&color, "color", 0, "event color (RGB int32; 0/-1 for default)")
	cmd.Flags().IntSliceVar(&reminders, "reminder", nil, "reminder minutes offset (repeatable)")
	cmd.Flags().StringVar(&recurrence, "recurrence", "", "recurrence rule (RFC5545)")
	cmd.Flags().StringArrayVar(&schemaEntries, "schema", nil, "schema entry (repeatable, format: ui_name=<name>,ui_status=<status>)")
	cmd.Flags().StringArrayVar(&attachments, "attachment", nil, "attachment file_token (repeatable)")
	cmd.Flags().StringVar(&vcType, "vc-type", "", "video conference type (vc|third_party|no_meeting)")
	cmd.Flags().StringVar(&vcIconType, "vc-icon-type", "", "third-party VC icon type (vc|live|default)")
	cmd.Flags().StringVar(&vcDescription, "vc-description", "", "third-party VC description")
	cmd.Flags().StringVar(&vcMeetingURL, "vc-meeting-url", "", "third-party VC meeting URL")
	cmd.Flags().StringVar(&vcOwnerID, "vc-owner-id", "", "VC owner user id")
	cmd.Flags().StringVar(&vcJoinPermission, "vc-join-permission", "", "VC join permission (anyone_can_join|only_organization_employees|only_event_attendees)")
	cmd.Flags().StringVar(&vcPassword, "vc-password", "", "VC meeting password")
	cmd.Flags().StringArrayVar(&vcAssignHosts, "vc-assign-host", nil, "VC assign host user id (repeatable)")
	cmd.Flags().BoolVar(&vcAutoRecord, "vc-auto-record", false, "VC auto record")
	cmd.Flags().BoolVar(&vcOpenLobby, "vc-open-lobby", false, "VC open lobby")
	cmd.Flags().BoolVar(&vcAllowAttendeesStart, "vc-allow-attendees-start", false, "allow attendees to start VC")
	cmd.Flags().BoolVar(&checkInEnabled, "check-in-enabled", false, "enable event check-in")
	cmd.Flags().StringVar(&checkInStartType, "check-in-start-type", "", "check-in start time type (before_event_start|after_event_start|after_event_end)")
	cmd.Flags().IntVar(&checkInStartDuration, "check-in-start-duration", 0, "check-in start offset minutes (0,5,15,30,60)")
	cmd.Flags().StringVar(&checkInEndType, "check-in-end-type", "", "check-in end time type (before_event_start|after_event_start|after_event_end)")
	cmd.Flags().IntVar(&checkInEndDuration, "check-in-end-duration", 0, "check-in end offset minutes (0,5,15,30,60)")
	cmd.Flags().BoolVar(&checkInNotifyAttendees, "check-in-notify-attendees", false, "notify attendees when check-in starts")

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
			rows := calendarEventDetailRows(event)
			if extraErr != nil {
				payload["extra_error"] = extraErr.Error()
				rows = append(rows, []string{"extra_error", extraErr.Error()})
			}
			text := tableTextFromRows([]string{"name", "value"}, rows, "no event found")
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
	var userIDType string
	var needNotification bool
	var visibility string
	var attendeeAbility string
	var freeBusyStatus string
	var locationName string
	var locationAddress string
	var locationLatitude float64
	var locationLongitude float64
	var color int
	var reminders []int
	var recurrence string
	var schemaEntries []string
	var attachments []string
	var attachmentDeletes []string
	var vcType string
	var vcIconType string
	var vcDescription string
	var vcMeetingURL string
	var vcOwnerID string
	var vcJoinPermission string
	var vcPassword string
	var vcAssignHosts []string
	var vcAutoRecord bool
	var vcOpenLobby bool
	var vcAllowAttendeesStart bool
	var checkInEnabled bool
	var checkInStartType string
	var checkInStartDuration int
	var checkInEndType string
	var checkInEndDuration int
	var checkInNotifyAttendees bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a calendar event",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			hasUpdate := summary != "" || description != "" || start != "" || end != "" || visibility != "" || attendeeAbility != "" || freeBusyStatus != "" || recurrence != "" || len(reminders) > 0 || len(schemaEntries) > 0 || len(attachments) > 0 || len(attachmentDeletes) > 0
			hasUpdate = hasUpdate ||
				cmd.Flags().Changed("need-notification") ||
				cmd.Flags().Changed("color") ||
				cmd.Flags().Changed("location-name") ||
				cmd.Flags().Changed("location-address") ||
				cmd.Flags().Changed("location-latitude") ||
				cmd.Flags().Changed("location-longitude") ||
				cmd.Flags().Changed("vc-type") ||
				cmd.Flags().Changed("vc-icon-type") ||
				cmd.Flags().Changed("vc-description") ||
				cmd.Flags().Changed("vc-meeting-url") ||
				cmd.Flags().Changed("vc-owner-id") ||
				cmd.Flags().Changed("vc-join-permission") ||
				cmd.Flags().Changed("vc-password") ||
				cmd.Flags().Changed("vc-assign-host") ||
				cmd.Flags().Changed("vc-auto-record") ||
				cmd.Flags().Changed("vc-open-lobby") ||
				cmd.Flags().Changed("vc-allow-attendees-start") ||
				cmd.Flags().Changed("check-in-enabled") ||
				cmd.Flags().Changed("check-in-start-type") ||
				cmd.Flags().Changed("check-in-start-duration") ||
				cmd.Flags().Changed("check-in-end-type") ||
				cmd.Flags().Changed("check-in-end-duration") ||
				cmd.Flags().Changed("check-in-notify-attendees")
			if !hasUpdate {
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
			schemas, err := parseCalendarSchemas(schemaEntries)
			if err != nil {
				return err
			}
			vchat, err := buildCalendarVChat(cmd, vcType, vcIconType, vcDescription, vcMeetingURL, vcOwnerID, vcJoinPermission, vcPassword, vcAssignHosts, vcAutoRecord, vcOpenLobby, vcAllowAttendeesStart)
			if err != nil {
				return err
			}
			location := buildCalendarLocation(cmd, locationName, locationAddress, locationLatitude, locationLongitude)
			checkIn, err := buildCalendarCheckIn(cmd, checkInEnabled, checkInStartType, checkInStartDuration, checkInEndType, checkInEndDuration, checkInNotifyAttendees)
			if err != nil {
				return err
			}
			reminderItems := buildCalendarReminders(reminders)
			attachmentItems := buildCalendarAttachments(attachments, attachmentDeletes)
			req := larksdk.UpdateCalendarEventRequest{
				CalendarID:      resolvedCalendarID,
				EventID:         eventID,
				Summary:         summary,
				Description:     description,
				UserIDType:      userIDType,
				Visibility:      visibility,
				AttendeeAbility: attendeeAbility,
				FreeBusyStatus:  freeBusyStatus,
				Location:        location,
				Color:           flagIntPtr(cmd, "color", color),
				Reminders:       reminderItems,
				Recurrence:      recurrence,
				VChat:           vchat,
				Schemas:         schemas,
				Attachments:     attachmentItems,
				EventCheckIn:    checkIn,
			}
			if needNotificationPtr := flagBoolPtr(cmd, "need-notification", needNotification); needNotificationPtr != nil {
				req.NeedNotification = needNotificationPtr
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
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user id type (open_id|union_id|user_id)")
	cmd.Flags().BoolVar(&needNotification, "need-notification", true, "notify attendees about event update")
	cmd.Flags().StringVar(&visibility, "visibility", "", "visibility (default|public|private)")
	cmd.Flags().StringVar(&attendeeAbility, "attendee-ability", "", "attendee ability (none|can_see_others|can_invite_others|can_modify_event)")
	cmd.Flags().StringVar(&freeBusyStatus, "free-busy-status", "", "free/busy status (busy|free)")
	cmd.Flags().StringVar(&locationName, "location-name", "", "location name")
	cmd.Flags().StringVar(&locationAddress, "location-address", "", "location address")
	cmd.Flags().Float64Var(&locationLatitude, "location-latitude", 0, "location latitude (GCJ-02 in CN, WGS84 elsewhere)")
	cmd.Flags().Float64Var(&locationLongitude, "location-longitude", 0, "location longitude (GCJ-02 in CN, WGS84 elsewhere)")
	cmd.Flags().IntVar(&color, "color", 0, "event color (RGB int32; 0/-1 for default)")
	cmd.Flags().IntSliceVar(&reminders, "reminder", nil, "reminder minutes offset (repeatable)")
	cmd.Flags().StringVar(&recurrence, "recurrence", "", "recurrence rule (RFC5545)")
	cmd.Flags().StringArrayVar(&schemaEntries, "schema", nil, "schema entry (repeatable, format: ui_name=<name>,ui_status=<status>)")
	cmd.Flags().StringArrayVar(&attachments, "attachment", nil, "attachment file_token to add (repeatable)")
	cmd.Flags().StringArrayVar(&attachmentDeletes, "attachment-delete", nil, "attachment file_token to delete (repeatable)")
	cmd.Flags().StringVar(&vcType, "vc-type", "", "video conference type (vc|third_party|no_meeting)")
	cmd.Flags().StringVar(&vcIconType, "vc-icon-type", "", "third-party VC icon type (vc|live|default)")
	cmd.Flags().StringVar(&vcDescription, "vc-description", "", "third-party VC description")
	cmd.Flags().StringVar(&vcMeetingURL, "vc-meeting-url", "", "third-party VC meeting URL")
	cmd.Flags().StringVar(&vcOwnerID, "vc-owner-id", "", "VC owner user id")
	cmd.Flags().StringVar(&vcJoinPermission, "vc-join-permission", "", "VC join permission (anyone_can_join|only_organization_employees|only_event_attendees)")
	cmd.Flags().StringVar(&vcPassword, "vc-password", "", "VC meeting password")
	cmd.Flags().StringArrayVar(&vcAssignHosts, "vc-assign-host", nil, "VC assign host user id (repeatable)")
	cmd.Flags().BoolVar(&vcAutoRecord, "vc-auto-record", false, "VC auto record")
	cmd.Flags().BoolVar(&vcOpenLobby, "vc-open-lobby", false, "VC open lobby")
	cmd.Flags().BoolVar(&vcAllowAttendeesStart, "vc-allow-attendees-start", false, "allow attendees to start VC")
	cmd.Flags().BoolVar(&checkInEnabled, "check-in-enabled", false, "enable event check-in")
	cmd.Flags().StringVar(&checkInStartType, "check-in-start-type", "", "check-in start time type (before_event_start|after_event_start|after_event_end)")
	cmd.Flags().IntVar(&checkInStartDuration, "check-in-start-duration", 0, "check-in start offset minutes (0,5,15,30,60)")
	cmd.Flags().StringVar(&checkInEndType, "check-in-end-type", "", "check-in end time type (before_event_start|after_event_start|after_event_end)")
	cmd.Flags().IntVar(&checkInEndDuration, "check-in-end-duration", 0, "check-in end offset minutes (0,5,15,30,60)")
	cmd.Flags().BoolVar(&checkInNotifyAttendees, "check-in-notify-attendees", false, "notify attendees when check-in starts")
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

func flagBoolPtr(cmd *cobra.Command, name string, value bool) *bool {
	if !cmd.Flags().Changed(name) {
		return nil
	}
	v := value
	return &v
}

func flagIntPtr(cmd *cobra.Command, name string, value int) *int {
	if !cmd.Flags().Changed(name) {
		return nil
	}
	v := value
	return &v
}

func buildCalendarReminders(values []int) []larksdk.CalendarEventReminder {
	if len(values) == 0 {
		return nil
	}
	reminders := make([]larksdk.CalendarEventReminder, 0, len(values))
	for _, minutes := range values {
		reminders = append(reminders, larksdk.CalendarEventReminder{Minutes: minutes})
	}
	return reminders
}

func parseCalendarSchemas(entries []string) ([]larksdk.CalendarEventSchema, error) {
	if len(entries) == 0 {
		return nil, nil
	}
	schemas := make([]larksdk.CalendarEventSchema, 0, len(entries))
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		var schema larksdk.CalendarEventSchema
		parts := strings.Split(entry, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			key, value, ok := strings.Cut(part, "=")
			if !ok {
				return nil, fmt.Errorf("schema %q must use key=value format", entry)
			}
			key = strings.TrimSpace(strings.ToLower(key))
			value = strings.TrimSpace(value)
			switch key {
			case "ui_name", "ui-name":
				schema.UIName = value
			case "ui_status", "ui-status":
				schema.UIStatus = value
			case "app_link", "app-link":
				schema.AppLink = value
			default:
				return nil, fmt.Errorf("schema %q has unsupported key %q", entry, key)
			}
		}
		if schema.UIName == "" {
			return nil, fmt.Errorf("schema %q must include ui_name", entry)
		}
		schemas = append(schemas, schema)
	}
	return schemas, nil
}

func buildCalendarAttachments(additions []string, deletes []string) []larksdk.CalendarEventAttachment {
	if len(additions) == 0 && len(deletes) == 0 {
		return nil
	}
	attachments := make([]larksdk.CalendarEventAttachment, 0, len(additions)+len(deletes))
	for _, token := range additions {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		attachments = append(attachments, larksdk.CalendarEventAttachment{FileToken: token})
	}
	for _, token := range deletes {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		deleted := true
		attachments = append(attachments, larksdk.CalendarEventAttachment{FileToken: token, IsDeleted: &deleted})
	}
	if len(attachments) == 0 {
		return nil
	}
	return attachments
}

func buildCalendarLocation(cmd *cobra.Command, name, address string, latitude, longitude float64) *larksdk.CalendarEventLocation {
	if !cmd.Flags().Changed("location-name") && !cmd.Flags().Changed("location-address") && !cmd.Flags().Changed("location-latitude") && !cmd.Flags().Changed("location-longitude") {
		if name == "" && address == "" && latitude == 0 && longitude == 0 {
			return nil
		}
	}
	return &larksdk.CalendarEventLocation{
		Name:      name,
		Address:   address,
		Latitude:  latitude,
		Longitude: longitude,
	}
}

func buildCalendarVChat(cmd *cobra.Command, vcType, iconType, description, meetingURL, ownerID, joinPermission, password string, assignHosts []string, autoRecord, openLobby, allowAttendeesStart bool) (*larksdk.CalendarEventVChat, error) {
	has := vcType != "" || iconType != "" || description != "" || meetingURL != "" || ownerID != "" || joinPermission != "" || password != "" || len(assignHosts) > 0 ||
		cmd.Flags().Changed("vc-auto-record") || cmd.Flags().Changed("vc-open-lobby") || cmd.Flags().Changed("vc-allow-attendees-start")
	if !has {
		return nil, nil
	}
	vchat := &larksdk.CalendarEventVChat{
		VCType:      vcType,
		IconType:    iconType,
		Description: description,
		MeetingURL:  meetingURL,
	}
	autoRecordPtr := flagBoolPtr(cmd, "vc-auto-record", autoRecord)
	openLobbyPtr := flagBoolPtr(cmd, "vc-open-lobby", openLobby)
	allowStartPtr := flagBoolPtr(cmd, "vc-allow-attendees-start", allowAttendeesStart)
	if ownerID != "" || joinPermission != "" || password != "" || len(assignHosts) > 0 || autoRecordPtr != nil || openLobbyPtr != nil || allowStartPtr != nil {
		vchat.MeetingSettings = &larksdk.CalendarEventMeetingSettings{
			OwnerID:               ownerID,
			JoinMeetingPermission: joinPermission,
			Password:              password,
			AssignHosts:           assignHosts,
			AutoRecord:            autoRecordPtr,
			OpenLobby:             openLobbyPtr,
			AllowAttendeesStart:   allowStartPtr,
		}
	}
	return vchat, nil
}

func buildCalendarCheckIn(cmd *cobra.Command, enabled bool, startType string, startDuration int, endType string, endDuration int, notify bool) (*larksdk.CalendarEventCheckIn, error) {
	changed := cmd.Flags().Changed("check-in-enabled") ||
		cmd.Flags().Changed("check-in-start-type") ||
		cmd.Flags().Changed("check-in-start-duration") ||
		cmd.Flags().Changed("check-in-end-type") ||
		cmd.Flags().Changed("check-in-end-duration") ||
		cmd.Flags().Changed("check-in-notify-attendees")
	if !changed {
		return nil, nil
	}
	enablePtr := flagBoolPtr(cmd, "check-in-enabled", enabled)
	if enablePtr == nil {
		return nil, errors.New("check-in-enabled is required when configuring check-in")
	}
	checkIn := &larksdk.CalendarEventCheckIn{
		EnableCheckIn:       enablePtr,
		NeedNotifyAttendees: flagBoolPtr(cmd, "check-in-notify-attendees", notify),
	}
	startChanged := cmd.Flags().Changed("check-in-start-type") || cmd.Flags().Changed("check-in-start-duration")
	if startChanged {
		if startType == "" {
			return nil, errors.New("check-in-start-type is required when setting check-in start time")
		}
		checkIn.CheckInStartTime = &larksdk.CalendarEventCheckInTime{
			TimeType: startType,
			Duration: flagIntPtr(cmd, "check-in-start-duration", startDuration),
		}
	}
	endChanged := cmd.Flags().Changed("check-in-end-type") || cmd.Flags().Changed("check-in-end-duration")
	if endChanged {
		if endType == "" {
			return nil, errors.New("check-in-end-type is required when setting check-in end time")
		}
		checkIn.CheckInEndTime = &larksdk.CalendarEventCheckInTime{
			TimeType: endType,
			Duration: flagIntPtr(cmd, "check-in-end-duration", endDuration),
		}
	}
	return checkIn, nil
}

func calendarEventDetailRows(event larksdk.CalendarEvent) [][]string {
	rows := make([][]string, 0, 24)
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
	add("start_time", formatEventTimeDetail(event.StartTime))
	add("end_time", formatEventTimeDetail(event.EndTime))
	add("vchat", formatVChat(event.VChat))
	add("visibility", event.Visibility)
	add("attendee_ability", event.AttendeeAbility)
	add("free_busy_status", event.FreeBusyStatus)
	add("location", formatLocation(event.Location))
	add("color", formatIntPtr(event.Color))
	add("reminders", formatReminders(event.Reminders))
	add("recurrence", event.Recurrence)
	add("is_exception", formatBoolPtr(event.IsException))
	add("recurring_event_id", event.RecurringEventID)
	add("create_time", formatEpoch(event.CreateTime))
	add("app_link", event.AppLink)
	add("event_organizer", formatOrganizer(event.EventOrganizer))
	add("schemas", formatSchemas(event.Schemas))
	add("attendees.count", strconv.Itoa(len(event.Attendees)))
	add("attendees", formatAttendeesSummary(event.Attendees))
	add("has_more_attendee", formatBoolPtr(event.HasMoreAttendee))
	add("attachments.count", strconv.Itoa(len(event.Attachments)))
	add("attachments", formatAttachments(event.Attachments))
	add("event_check_in", formatEventCheckIn(event.EventCheckIn))

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

func formatFloat(value *float64) string {
	if value == nil {
		return ""
	}
	return strconv.FormatFloat(*value, 'f', -1, 64)
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

func formatAttendeesSummary(attendees []larksdk.CalendarEventAttendee) string {
	if len(attendees) == 0 {
		return "0"
	}
	values := make([]string, 0, len(attendees))
	for _, attendee := range attendees {
		parts := make([]string, 0, 10)
		if attendee.DisplayName != "" {
			parts = append(parts, attendee.DisplayName)
		}
		if attendee.UserID != "" {
			parts = append(parts, "user_id="+attendee.UserID)
		} else if attendee.ThirdPartyEmail != "" {
			parts = append(parts, "email="+attendee.ThirdPartyEmail)
		} else if attendee.RoomID != "" {
			parts = append(parts, "room_id="+attendee.RoomID)
		} else if attendee.ChatID != "" {
			parts = append(parts, "chat_id="+attendee.ChatID)
		}
		if attendee.RsvpStatus != "" {
			parts = append(parts, "rsvp="+attendee.RsvpStatus)
		}
		if attendee.IsOptional != nil {
			parts = append(parts, "optional="+strconv.FormatBool(*attendee.IsOptional))
		}
		if attendee.IsOrganizer != nil {
			parts = append(parts, "organizer="+strconv.FormatBool(*attendee.IsOrganizer))
		}
		if attendee.IsExternal != nil {
			parts = append(parts, "external="+strconv.FormatBool(*attendee.IsExternal))
		}
		if len(parts) == 0 {
			parts = append(parts, "type="+attendee.Type)
		}
		values = append(values, strings.Join(parts, " "))
	}
	return fmt.Sprintf("count=%d; %s", len(attendees), strings.Join(values, " | "))
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

func formatEventTimeDetail(eventTime larksdk.CalendarEventTime) string {
	parts := make([]string, 0, 3)
	display := formatEventTime(eventTime)
	if display != "" {
		parts = append(parts, display)
	}
	if eventTime.Timestamp != "" {
		parts = append(parts, "timestamp="+eventTime.Timestamp)
	}
	if eventTime.Timezone != "" {
		parts = append(parts, "timezone="+eventTime.Timezone)
	}
	if eventTime.Date != "" && display == "" {
		parts = append(parts, "date="+eventTime.Date)
	}
	return strings.Join(parts, " | ")
}

func formatLocation(location *larksdk.CalendarEventLocation) string {
	if location == nil {
		return ""
	}
	parts := make([]string, 0, 4)
	if location.Name != "" {
		parts = append(parts, "name="+location.Name)
	}
	if location.Address != "" {
		parts = append(parts, "address="+location.Address)
	}
	if location.Latitude != nil || location.Longitude != nil {
		if location.Latitude != nil {
			parts = append(parts, "latitude="+formatFloat(location.Latitude))
		}
		if location.Longitude != nil {
			parts = append(parts, "longitude="+formatFloat(location.Longitude))
		}
	}
	return strings.Join(parts, " | ")
}

func formatVChat(vchat *larksdk.CalendarEventVChat) string {
	if vchat == nil {
		return ""
	}
	parts := make([]string, 0, 6)
	if vchat.VCType != "" {
		parts = append(parts, "vc_type="+vchat.VCType)
	}
	if vchat.IconType != "" {
		parts = append(parts, "icon_type="+vchat.IconType)
	}
	if vchat.Description != "" {
		parts = append(parts, "description="+vchat.Description)
	}
	if vchat.MeetingURL != "" {
		parts = append(parts, "meeting_url="+vchat.MeetingURL)
	}
	settings := vchat.MeetingSettings
	if settings != nil {
		settingsParts := make([]string, 0, 7)
		if settings.OwnerID != "" {
			settingsParts = append(settingsParts, "owner_id="+settings.OwnerID)
		}
		if settings.JoinMeetingPermission != "" {
			settingsParts = append(settingsParts, "join_permission="+settings.JoinMeetingPermission)
		}
		if settings.Password != "" {
			settingsParts = append(settingsParts, "password="+settings.Password)
		}
		if len(settings.AssignHosts) > 0 {
			settingsParts = append(settingsParts, "assign_hosts="+formatStringSlice(settings.AssignHosts))
		}
		if settings.AutoRecord != nil {
			settingsParts = append(settingsParts, "auto_record="+strconv.FormatBool(*settings.AutoRecord))
		}
		if settings.OpenLobby != nil {
			settingsParts = append(settingsParts, "open_lobby="+strconv.FormatBool(*settings.OpenLobby))
		}
		if settings.AllowAttendeesStart != nil {
			settingsParts = append(settingsParts, "allow_attendees_start="+strconv.FormatBool(*settings.AllowAttendeesStart))
		}
		if len(settingsParts) > 0 {
			parts = append(parts, "meeting_settings={"+strings.Join(settingsParts, ", ")+"}")
		}
	}
	return strings.Join(parts, " | ")
}

func formatOrganizer(organizer *larksdk.CalendarEventOrganizer) string {
	if organizer == nil {
		return ""
	}
	parts := make([]string, 0, 2)
	if organizer.DisplayName != "" {
		parts = append(parts, organizer.DisplayName)
	}
	if organizer.UserID != "" {
		parts = append(parts, "user_id="+organizer.UserID)
	}
	return strings.Join(parts, " | ")
}

func formatEventCheckIn(checkIn *larksdk.CalendarEventCheckIn) string {
	if checkIn == nil {
		return ""
	}
	parts := make([]string, 0, 4)
	if checkIn.EnableCheckIn != nil {
		parts = append(parts, "enable="+strconv.FormatBool(*checkIn.EnableCheckIn))
	}
	if checkIn.CheckInStartTime != nil {
		startParts := make([]string, 0, 2)
		if checkIn.CheckInStartTime.TimeType != "" {
			startParts = append(startParts, checkIn.CheckInStartTime.TimeType)
		}
		if checkIn.CheckInStartTime.Duration != nil {
			startParts = append(startParts, "duration="+strconv.Itoa(*checkIn.CheckInStartTime.Duration))
		}
		if len(startParts) > 0 {
			parts = append(parts, "start={"+strings.Join(startParts, ", ")+"}")
		}
	}
	if checkIn.CheckInEndTime != nil {
		endParts := make([]string, 0, 2)
		if checkIn.CheckInEndTime.TimeType != "" {
			endParts = append(endParts, checkIn.CheckInEndTime.TimeType)
		}
		if checkIn.CheckInEndTime.Duration != nil {
			endParts = append(endParts, "duration="+strconv.Itoa(*checkIn.CheckInEndTime.Duration))
		}
		if len(endParts) > 0 {
			parts = append(parts, "end={"+strings.Join(endParts, ", ")+"}")
		}
	}
	if checkIn.NeedNotifyAttendees != nil {
		parts = append(parts, "notify="+strconv.FormatBool(*checkIn.NeedNotifyAttendees))
	}
	return strings.Join(parts, " | ")
}
