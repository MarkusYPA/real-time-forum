package utils

import "time"

// Custom function to format time globally
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02 15:04:05") // YYYY-MM-DD HH:MM:SS
}
