package utils

import (
	"encoding/json"
	"strings"
	"time"
)

// DateOnly is a custom type to format date as YYYY-MM-DD
type DateOnly struct {
	time.Time
}

const dateFormat = "2006-01-02"

// UnmarshalJSON converts a JSON string to DateOnly
func (d *DateOnly) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		d.Time = time.Time{}
		return nil
	}
	t, err := time.Parse(dateFormat, s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

// MarshalJSON converts DateOnly to JSON string
func (d DateOnly) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.Time.Format(dateFormat))
}

// String returns DateOnly as YYYY-MM-DD
func (d DateOnly) String() string {
	return d.Time.Format(dateFormat)
}

// ToDateOnly converts time.Time to DateOnly
func ToDateOnly(t time.Time) *DateOnly {
	return &DateOnly{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())}
}

// ToTime converts DateOnly back to time.Time
func ToTime(d DateOnly) time.Time {
	return d.Time
}