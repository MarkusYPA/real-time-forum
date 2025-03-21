package main

import (
	"net/http"
	"real-time-forum/db"
	"time"
)

// ValidateSession returns user id, name and if session is (still) valid
func ValidateSession(r *http.Request) (string, string, bool) {
	validSes := true
	var userID string
	var userName string

	cookie, _ := r.Cookie("session_token")
	if cookie != nil {
		query := `SELECT user_id, username FROM sessions WHERE session_token = ? AND expires_at > ?`
		err := db.DB.QueryRow(query, cookie.Value, time.Now()).Scan(&userID, &userName)
		if err != nil { // invalid session
			validSes = false
		}
	} else {
		validSes = false
	}

	return userID, userName, validSes
}

func timeStrings(created string) (string, string, error) {
	createdGoTime, err := time.Parse(time.RFC3339, created) // "created" looks something like this: 2024-12-02T15:44:52Z
	if err != nil {
		return "", "", err
	}

	// Convert to Finnish timezone (UTC+2)
	location, err := time.LoadLocation("Europe/Helsinki")
	if err != nil {
		return "", "", err
	}
	createdGoTime = createdGoTime.In(location)

	day := createdGoTime.Format("2.1.2006")
	time := createdGoTime.Format("15:04") //"15.04.05"

	return day, time, nil
}
