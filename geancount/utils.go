package geancount

import "time"

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
