package models

import (
	"database/sql"
	"errors"
	"log"
	"real-time-forum/db"
	"real-time-forum/utils"
	"time"
)

// User struct represents the user data model
type Session struct {
	ID           int       `json:"id"`
	SessionToken string    `json:"session_token"`
	UserId       int       `json:"user_id"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func InsertSession(session *Session) (*Session, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	// Generate UUID for the user if not already set
	if session.SessionToken == "" {
		uuidSessionTokenid, err := utils.GenerateUuid()
		if err != nil {
			return nil, err
		}
		session.SessionToken = uuidSessionTokenid
	}

	// Set session expiration time
	session.ExpiresAt = time.Now().Add(12 * time.Hour)

	// Start a transaction for atomicity
	tx, err := db.Begin()
	if err != nil {
		return &Session{}, err
	}

	updateQuery := `UPDATE sessions SET expires_at = CURRENT_TIMESTAMP WHERE user_id = ? AND expires_at > CURRENT_TIMESTAMP;`
	_, updateErr := tx.Exec(updateQuery, session.UserId)
	if updateErr != nil {
		tx.Rollback()
		return nil, updateErr
	}

	insertQuery := `INSERT INTO sessions (session_token, user_id, expires_at) VALUES (?, ?, ?);`
	_, insertErr := tx.Exec(insertQuery, session.SessionToken, session.UserId, session.ExpiresAt)
	if insertErr != nil {
		tx.Rollback()
		// Check if the error is a SQLite constraint violation
		if sqliteErr, ok := insertErr.(interface{ ErrorCode() int }); ok {
			if sqliteErr.ErrorCode() == 19 { // SQLite constraint violation error code
				return nil, sql.ErrNoRows // Return custom error to indicate a duplicate
			}
		}
		return nil, insertErr
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		tx.Rollback() // Rollback on error
		return nil, err
	}

	return session, nil
}

func SelectSession(sessionToken string) (User, time.Time, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	var user User
	var expirationTime time.Time
	err := db.QueryRow(`SELECT 
							u.id as user_id,u.uuid ,u.type as user_type, u.username as username, u.email as user_email, u.gender, u.firstname, u.lastname, u.age,
							expires_at 
						FROM sessions s
							INNER JOIN users u
								ON s.user_id = u.id
						WHERE session_token = ?`, sessionToken).Scan(&user.ID, &user.UUID, &user.Type, &user.Username, &user.Email, &user.Gender, &user.FirstName, &user.LastName, &user.Age, &expirationTime)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			// Handle other database errors
			return User{}, time.Time{}, errors.New("sql: no rows in result set")
		} else {
			// Handle other database errors
			return User{}, time.Time{}, errors.New("database error")
		}
	}
	return user, expirationTime, nil
}
func DeleteSession(sessionToken string) error {

	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes
	_, err := db.Exec(`UPDATE sessions
					SET expires_at = CURRENT_TIMESTAMP
					WHERE session_token = ?;`, sessionToken)
	if err != nil {
		// Handle other database errors
		log.Fatal(err)
		return errors.New("database error")
	}

	return nil

}
