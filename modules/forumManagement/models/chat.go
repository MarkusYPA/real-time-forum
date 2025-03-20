package models

import (
	"database/sql"
	"fmt"
	"forum/db"
	userManagementModels "forum/modules/userManagement/models"
	"forum/utils"
	"log"
	"sort"
	"time"
)

type Chat struct {
	ID               int                       `json:"id"`
	UUID             string                    `json:"uuid"`
	User_id_1        int                       `json:"user_id_1"`
	User_id_2        int                       `json:"user_id_2"`
	Status           string                    `json:"status"`
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        *time.Time                `json:"updated_at"`
	UpdatedBy        *int                  	   `json:"updated_by"`
}

type message struct {
	ID               int                       `json:"id"`
	Chat_id          int                   `json:"chat_id"`
	User_id_from     int                   `json:"user_id_from"`
	Content			string					`json:"content"`
	Status           string                    `json:"status"`
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        *time.Time                `json:"updated_at"`
}

func InsertMessage(content string, user_id_from int, chatID int) error {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	insertQuery := `INSERT INTO messages (chat_id, user_id_from, content) VALUES (?, ?, ?);`
	result, insertErr := db.Exec(insertQuery, chatID, user_id_from, content)
	if insertErr != nil {
		// Check if the error is a SQLite constraint violation
		if sqliteErr, ok := insertErr.(interface{ ErrorCode() int }); ok {
			if sqliteErr.ErrorCode() == 19 { // SQLite constraint violation error code
				return sql.ErrNoRows // Return custom error to indicate a duplicate
			}
		}
		return insertErr
	}
	updateErr := UpdateChat(chatID, user_id_from)
	if updateErr != nil{
		return updateErr
	}

	return nil
}

func UpdateMessageStatus(messageID int, satus string, user_id int) error {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	updateQuery := `UPDATE messages
					SET status = ?,
						updated_at = CURRENT_TIMESTAMP,
					WHERE id = ?;`
	_, updateErr := db.Exec(updateQuery, status, messageID)
	if updateErr != nil {
		return updateErr
	}

	return nil
}


func InsertChat(user_id_1, user_id_2 int) (int, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	UUID, err = utils.GenerateUuid()
	if err != nil {
		return -1, err
	}

	insertQuery := `INSERT INTO chats (uuid, user_id_1, user_id_2) VALUES (?, ?, ?);`
	result, insertErr := db.Exec(insertQuery, UUID, user_id_1, user_id_2)
	if insertErr != nil {
		// Check if the error is a SQLite constraint violation
		if sqliteErr, ok := insertErr.(interface{ ErrorCode() int }); ok {
			if sqliteErr.ErrorCode() == 19 { // SQLite constraint violation error code
				return -1, sql.ErrNoRows // Return custom error to indicate a duplicate
			}
		}
		return -1, insertErr
	}

	// Retrieve the last inserted ID
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}

	return int(lastInsertID), nil
}

func UpdateChat(chatID, user_id int) error {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	// Start a transaction for atomicity
	updateQuery := `UPDATE chats
					SET updated_at = CURRENT_TIMESTAMP,
						updated_by = ?
					WHERE id = ?;`
	_, updateErr := db.Exec(updateQuery, user_id, chatID)
	if updateErr != nil {
		return updateErr
	}

	return nil
}

func UpdateChatStatus(chatID int, satus string, user_id int) error {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	updateQuery := `UPDATE chats
					SET status = ?,
						updated_at = CURRENT_TIMESTAMP,
						updated_by = ?
					WHERE id = ?;`
	_, updateErr := db.Exec(updateQuery, status, user_id, chatID)
	if updateErr != nil {
		return updateErr
	}

	return nil
}


