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

func newMeetingsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meetings",
		Short: "Manage meetings",
	}
	cmd.AddCommand(newMeetingInfoCmd(state))
	cmd.AddCommand(newMeetingListCmd(state))
	cmd.AddCommand(newMeetingCreateCmd(state))
	cmd.AddCommand(newMeetingUpdateCmd(state))
	cmd.AddCommand(newMeetingDeleteCmd(state))
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
	var start string
	var end string
	var limit int
	var meetingStatus int
	var meetingNo string
	var userID string
	var roomID string
	var meetingType int
	var includeExternal bool
	var includeWebinar bool
	var userIDType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List meetings",
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				return errors.New("limit must be greater than 0")
			}
			if start == "" || end == "" {
				return errors.New("start and end times are required")
			}
			startUnix, err := parseMeetingTime(start)
			if err != nil {
				return fmt.Errorf("invalid start time: %w", err)
			}
			endUnix, err := parseMeetingTime(end)
			if err != nil {
				return fmt.Errorf("invalid end time: %w", err)
			}
			if endUnix <= startUnix {
				return errors.New("end time must be after start time")
			}
			filterCount := 0
			if meetingNo != "" {
				filterCount++
			}
			if userID != "" {
				filterCount++
			}
			if roomID != "" {
				filterCount++
			}
			if cmd.Flags().Changed("meeting-type") {
				filterCount++
			}
			if filterCount > 1 {
				return errors.New("meeting-no, user-id, room-id, and meeting-type are mutually exclusive")
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
			var meetingStatusPtr *int
			if cmd.Flags().Changed("status") {
				meetingStatusPtr = &meetingStatus
			}
			var meetingTypePtr *int
			if cmd.Flags().Changed("meeting-type") {
				meetingTypePtr = &meetingType
			}
			var includeExternalPtr *bool
			if cmd.Flags().Changed("include-external") {
				includeExternalPtr = &includeExternal
			}
			var includeWebinarPtr *bool
			if cmd.Flags().Changed("include-webinar") {
				includeWebinarPtr = &includeWebinar
			}
			for {
				pageSize := remaining
				result, err := state.SDK.ListMeetings(context.Background(), token, larksdk.ListMeetingsRequest{
					StartTime:               strconv.FormatInt(startUnix, 10),
					EndTime:                 strconv.FormatInt(endUnix, 10),
					MeetingStatus:           meetingStatusPtr,
					MeetingNo:               meetingNo,
					UserID:                  userID,
					RoomID:                  roomID,
					MeetingType:             meetingTypePtr,
					PageSize:                pageSize,
					PageToken:               pageToken,
					IncludeExternalMeetings: includeExternalPtr,
					IncludeWebinar:          includeWebinarPtr,
					UserIDType:              userIDType,
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

	cmd.Flags().StringVar(&start, "start", "", "start time (RFC3339 or unix seconds)")
	cmd.Flags().StringVar(&end, "end", "", "end time (RFC3339 or unix seconds)")
	cmd.Flags().IntVar(&limit, "limit", 20, "max number of meetings to return")
	cmd.Flags().IntVar(&meetingStatus, "status", 0, "meeting status")
	cmd.Flags().StringVar(&meetingNo, "meeting-no", "", "meeting number (9 digits)")
	cmd.Flags().StringVar(&userID, "user-id", "", "participant user ID")
	cmd.Flags().StringVar(&roomID, "room-id", "", "room ID")
	cmd.Flags().IntVar(&meetingType, "meeting-type", 0, "meeting type")
	cmd.Flags().BoolVar(&includeExternal, "include-external", false, "include external meetings")
	cmd.Flags().BoolVar(&includeWebinar, "include-webinar", false, "include webinars")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user ID type (user_id, union_id, open_id)")
	_ = cmd.MarkFlagRequired("start")
	_ = cmd.MarkFlagRequired("end")
	return cmd
}

func newMeetingCreateCmd(state *appState) *cobra.Command {
	var endTime string
	var ownerID string
	var userIDType string
	var topic string
	var meetingInitialType int
	var autoRecord bool
	var password string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a meeting reservation",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, tokenType, err := resolveAccessToken(context.Background(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if tokenType == tokenTypeTenant && ownerID == "" {
				return errors.New("owner-id is required when using tenant access token")
			}
			endUnix, err := parseMeetingTime(endTime)
			if err != nil {
				return fmt.Errorf("invalid end time: %w", err)
			}
			settings := buildReserveMeetingSettings(cmd, topic, meetingInitialType, autoRecord, password)
			reserve, correction, err := state.SDK.ApplyReserve(context.Background(), token, larksdk.ApplyReserveRequest{
				EndTime:         strconv.FormatInt(endUnix, 10),
				OwnerID:         ownerID,
				UserIDType:      userIDType,
				MeetingSettings: settings,
			})
			if err != nil {
				return err
			}
			payload := map[string]any{"reserve": reserve}
			if correction != nil {
				payload["reserve_correction_check_info"] = correction
			}
			text := tableTextRow(
				[]string{"reserve_id", "meeting_no", "topic", "end_time", "url"},
				[]string{reserve.ID, reserve.MeetingNo, reserveTopic(reserve), reserve.EndTime, reserve.URL},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&endTime, "end-time", "", "end time (RFC3339 or unix seconds)")
	cmd.Flags().StringVar(&ownerID, "owner-id", "", "owner user ID (required for tenant token)")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user ID type (user_id, union_id, open_id)")
	cmd.Flags().StringVar(&topic, "topic", "", "meeting topic")
	cmd.Flags().IntVar(&meetingInitialType, "meeting-initial-type", 0, "meeting initial type")
	cmd.Flags().BoolVar(&autoRecord, "auto-record", false, "enable auto recording")
	cmd.Flags().StringVar(&password, "password", "", "meeting password (4-9 digits)")
	_ = cmd.MarkFlagRequired("end-time")
	return cmd
}

func newMeetingUpdateCmd(state *appState) *cobra.Command {
	var reserveID string
	var endTime string
	var userIDType string
	var topic string
	var meetingInitialType int
	var autoRecord bool
	var password string

	cmd := &cobra.Command{
		Use:   "update <reserve-id>",
		Short: "Update a meeting reservation",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) > 0 {
				if reserveID != "" && reserveID != args[0] {
					return errors.New("reserve-id provided twice")
				}
				if err := cmd.Flags().Set("reserve-id", args[0]); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			endUnix := ""
			if cmd.Flags().Changed("end-time") {
				parsed, err := parseMeetingTime(endTime)
				if err != nil {
					return fmt.Errorf("invalid end time: %w", err)
				}
				endUnix = strconv.FormatInt(parsed, 10)
			}
			settings := buildReserveMeetingSettings(cmd, topic, meetingInitialType, autoRecord, password)
			if endUnix == "" && settings == nil {
				return errors.New("no fields to update")
			}
			reserve, correction, err := state.SDK.UpdateReserve(context.Background(), token, larksdk.UpdateReserveRequest{
				ReserveID:       reserveID,
				EndTime:         endUnix,
				UserIDType:      userIDType,
				MeetingSettings: settings,
			})
			if err != nil {
				return err
			}
			payload := map[string]any{"reserve": reserve}
			if correction != nil {
				payload["reserve_correction_check_info"] = correction
			}
			text := tableTextRow(
				[]string{"reserve_id", "meeting_no", "topic", "end_time", "url"},
				[]string{reserve.ID, reserve.MeetingNo, reserveTopic(reserve), reserve.EndTime, reserve.URL},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&reserveID, "reserve-id", "", "reserve ID (or provide as positional argument)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "end time (RFC3339 or unix seconds)")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "", "user ID type (user_id, union_id, open_id)")
	cmd.Flags().StringVar(&topic, "topic", "", "meeting topic")
	cmd.Flags().IntVar(&meetingInitialType, "meeting-initial-type", 0, "meeting initial type")
	cmd.Flags().BoolVar(&autoRecord, "auto-record", false, "enable auto recording")
	cmd.Flags().StringVar(&password, "password", "", "meeting password (4-9 digits)")
	_ = cmd.MarkFlagRequired("reserve-id")
	return cmd
}

func newMeetingDeleteCmd(state *appState) *cobra.Command {
	var reserveID string

	cmd := &cobra.Command{
		Use:   "delete <reserve-id>",
		Short: "Delete a meeting reservation",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) > 0 {
				if reserveID != "" && reserveID != args[0] {
					return errors.New("reserve-id provided twice")
				}
				if err := cmd.Flags().Set("reserve-id", args[0]); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			if err := state.SDK.DeleteReserve(context.Background(), token, larksdk.DeleteReserveRequest{ReserveID: reserveID}); err != nil {
				return err
			}
			payload := map[string]any{"deleted": true, "reserve_id": reserveID}
			text := tableTextRow([]string{"deleted", "reserve_id"}, []string{strconv.FormatBool(true), reserveID})
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&reserveID, "reserve-id", "", "reserve ID (or provide as positional argument)")
	_ = cmd.MarkFlagRequired("reserve-id")
	return cmd
}

func buildReserveMeetingSettings(cmd *cobra.Command, topic string, meetingInitialType int, autoRecord bool, password string) *larksdk.ReserveMeetingSetting {
	settings := larksdk.ReserveMeetingSetting{}
	changed := false
	if cmd.Flags().Changed("topic") {
		settings.Topic = &topic
		changed = true
	}
	if cmd.Flags().Changed("meeting-initial-type") {
		settings.MeetingInitialType = &meetingInitialType
		changed = true
	}
	if cmd.Flags().Changed("auto-record") {
		settings.AutoRecord = &autoRecord
		changed = true
	}
	if cmd.Flags().Changed("password") {
		settings.Password = &password
		changed = true
	}
	if !changed {
		return nil
	}
	return &settings
}

func reserveTopic(reserve larksdk.Reserve) string {
	if reserve.MeetingSettings == nil || reserve.MeetingSettings.Topic == nil {
		return ""
	}
	return *reserve.MeetingSettings.Topic
}

func parseMeetingTime(raw string) (int64, error) {
	if raw == "" {
		return 0, errors.New("time is required")
	}
	if unix, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return unix, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return 0, err
	}
	return parsed.Unix(), nil
}
