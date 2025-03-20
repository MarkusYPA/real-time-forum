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
