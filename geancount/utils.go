package geancount

import "time"

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}
