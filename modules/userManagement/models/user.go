package models

import (
	"database/sql"
	"errors"
	"forum/db"
	"forum/utils"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User struct represents the user data model
type User struct {
	ID        int        `json:"id"`
	UUID      string     `json:"uuid"`
	Type      string     `json:"type"`
	Age       string     `json:"age"`
	Gender    string     `json:"gender"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Password  string     `json:"password"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	UpdatedBy *int       `json:"updated_by"`
}

func InsertUser(user *User) (int, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	// Generate UUID for the user if not already set
	if user.UUID == "" {
		uuid, err := utils.GenerateUuid()
		if err != nil {
			return -1, err
		}
		user.UUID = uuid
	}

	var existingEmail string
	var existingUsername string
	emailCheckQuery := `SELECT email, username FROM users WHERE email = ? OR username = ? LIMIT 1;`
	err := db.QueryRow(emailCheckQuery, user.Email, user.Username).Scan(&existingEmail, &existingUsername)
	if err == nil {
		if existingEmail == user.Email {
			return -1, errors.New("duplicateEmail")
		}
		if existingUsername == user.Username {
			return -1, errors.New("duplicateUsername")
		}
	}

	insertQuery := `INSERT INTO users (uuid, name, username, email, password) VALUES (?, ?, ?, ?, ?);`
	result, insertErr := db.Exec(insertQuery, user.UUID, user.Username, user.Username, user.Email, user.Password)
	if insertErr != nil {
		// Check if the error is a SQLite constraint violation (duplicate entry)
		if sqliteErr, ok := insertErr.(interface{ ErrorCode() int }); ok {
			if sqliteErr.ErrorCode() == 19 { // 19 = UNIQUE constraint failed (SQLite error code)
				return -1, errors.New("user with this email or username already exists")
			}
		}
		return -1, insertErr // Other DB errors
	}

	// Retrieve the last inserted ID
	userId, err := result.LastInsertId()
	if err != nil {
		log.Fatal(err)
		return -1, err
	}

	return int(userId), nil
}

func AuthenticateUser(input, password string) (bool, int, error) {
	// Open SQLite database
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	// Query to retrieve the hashed password stored in the database for the given username
	var userID int
	var storedHashedPassword string
	err := db.QueryRow("SELECT id, password FROM users WHERE username = ? OR email = ?", input, input).Scan(&userID, &storedHashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			// Username not found
			return false, -1, errors.New("username or email not found")
		}
		// Handle other database errors
		log.Fatal(err)
	}

	// Compare the entered password with the stored hashed password using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(password))
	if err != nil {
		// Password is incorrect
		return false, -1, errors.New("password is incorrect")
	}

	// Successful login if no errors occurred
	return true, userID, nil
}

func findUserByUUID(UUID string) (int, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	selectQuery := `
		SELECT
			id
		FROM users
			WHERE uuid = ?;
	`
	id, selectError := db.Query(selectQuery, UUID)
	if selectError != nil {
		return -1, selectError
	}
	return id, nil
}
