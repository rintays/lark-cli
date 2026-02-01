package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func parseTimeArg(raw string, now time.Time) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	if isAllDigits(raw) {
		return raw, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return fmt.Sprintf("%d", t.Unix()), nil
	}
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return fmt.Sprintf("%d", t.Unix()), nil
	}
	if strings.HasPrefix(raw, "-") || strings.HasPrefix(raw, "+") {
		dur, err := parseRelativeDuration(raw)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", now.Add(dur).Unix()), nil
	}
	return "", fmt.Errorf("invalid time %q (expected unix seconds, RFC3339, or relative duration like -24h/-7d)", raw)
}

func parseRelativeDuration(raw string) (time.Duration, error) {
	normalized := strings.TrimSpace(raw)
	if strings.HasSuffix(normalized, "d") {
		number := strings.TrimSuffix(normalized, "d")
		if number == "" || number == "-" || number == "+" {
			return 0, fmt.Errorf("invalid relative duration %q", raw)
		}
		return time.ParseDuration(number + "24h")
	}
	return time.ParseDuration(normalized)
}

func isAllDigits(raw string) bool {
	if raw == "" {
		return false
	}
	if raw[0] == '-' || raw[0] == '+' {
		return false
	}
	_, err := strconv.ParseInt(raw, 10, 64)
	return err == nil
}
