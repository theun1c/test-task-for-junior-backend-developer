package handlers

import (
	"fmt"
	"strings"
	"time"
)

const calendarDateTimeLayout = "02.01.2006 15:04"

func parseCalendarDateTime(raw string) (time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("datetime is required in format %s", calendarDateTimeLayout)
	}

	parsed, err := time.ParseInLocation(calendarDateTimeLayout, trimmed, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid datetime format, expected %s", calendarDateTimeLayout)
	}

	return parsed, nil
}

func parseOptionalCalendarDateTime(raw *string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}

	if strings.TrimSpace(*raw) == "" {
		return nil, nil
	}

	parsed, err := parseCalendarDateTime(*raw)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func formatCalendarDateTime(value time.Time) string {
	return value.In(time.Local).Format(calendarDateTimeLayout)
}

func formatOptionalCalendarDateTime(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := formatCalendarDateTime(*value)

	return &formatted
}
