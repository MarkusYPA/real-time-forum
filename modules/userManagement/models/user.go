package models

import (
	"database/sql"
	"errors"
	"fmt"

	"log"
	"real-time-forum/db"
	"real-time-forum/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User struct represents the user data model
type User struct {
	ID             int        `json:"id"`
	UUID           string     `json:"uuid"`
	Type           string     `json:"type"`
	Age            string     `json:"age"`
	Gender         string     `json:"gender"`
	FirstName      string     `json:"firstName"`
	LastName       string     `json:"lastName"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	Password       string     `json:"password"`
	Status         string     `json:"status"`
	LastTimeOnline time.Time  `json:"lastTimeOnline"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
	UpdatedBy      *int       `json:"updated_by"`
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

	//insertQuery := `INSERT INTO users (uuid, name, username, email, password) VALUES (?, ?, ?, ?, ?);`
	//result, insertErr := db.Exec(insertQuery, user.UUID, user.Username, user.Username, user.Email, user.Password)

	insertQuery := `INSERT INTO users (uuid, username, email, password, age, gender, firstname, lastname) VALUES (?, ?, ?, ?, ?, ?, ?, ?);`
	result, insertErr := db.Exec(insertQuery, user.UUID, user.Username, user.Email, user.Password, user.Age, user.Gender, user.FirstName, user.LastName)
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

func AuthenticateUser(input, password string) (int, error) {
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
			return -1, errors.New("username or email not found")
		}
		return -1, err
	}

	// Compare the entered password with the stored hashed password using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(password))
	if err != nil {
		// Password is incorrect
		return -1, errors.New("password is incorrect")
	}

	// Successful login if no errors occurred
	return userID, nil
}

func FindUserByUUID(UUID string) (int, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	selectQuery := `
		SELECT
			id
		FROM users
			WHERE uuid = ?;
	`
	idRow, selectError := db.Query(selectQuery, UUID)
	if selectError != nil {
		return -1, selectError
	}

	var id int
	for idRow.Next() {
		if err := idRow.Scan(&id); err != nil {
			fmt.Printf("Failed to scan row: %v\n", err)
		}
	}

	return id, nil
}

func FindUsernameByID(ID int) (string, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	selectQuery := `
		SELECT
			username
		FROM users
			WHERE id = ?;
	`
	idRow, selectError := db.Query(selectQuery, ID)
	if selectError != nil {
		return "", selectError
	}

	var name string
	for idRow.Next() {
		if err := idRow.Scan(&name); err != nil {
			fmt.Printf("Failed to scan row: %v\n", err)
		}
	}

	return name, nil
}

func FindUsername(UUID string) (string, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	selectQuery := `
		SELECT
			username
		FROM users
			WHERE uuid = ?;
	`
	idRow, selectError := db.Query(selectQuery, UUID)
	if selectError != nil {
		return "", selectError
	}

	var username string
	for idRow.Next() {
		if err := idRow.Scan(&username); err != nil {
			fmt.Printf("Failed to scan row: %v\n", err)
		}
	}

	return username, nil
}

func UpdateOnlineTime(UserUUID string) error {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	updateQuery := `UPDATE users
	SET 
		last_time_online = CURRENT_TIMESTAMP
	WHERE uuid = ?;`
	_, updateErr := db.Exec(updateQuery, UserUUID)
	if updateErr != nil {
		fmt.Println(updateErr)
		return updateErr
	}
	return nil
}
