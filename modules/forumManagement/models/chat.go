package models

import (
	"database/sql"
	"fmt"
	"real-time-forum/db"
	userModels "real-time-forum/modules/userManagement/models"
	"real-time-forum/utils"
	"sort"
	"strings"
	"time"
)

type Chat struct {
	ID        int        `json:"id"`
	UUID      string     `json:"uuid"`
	User_id_1 int        `json:"user_id_1"`
	User_id_2 int        `json:"user_id_2"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	UpdatedBy *int       `json:"updated_by"`
}

type Message struct {
	ID             int        `json:"id"`
	ChatUUID       string     `json:"chat_uuid"`
	ChatID         int        `json:"chat_id"`
	UserIDFrom     int        `json:"user_id_from"`
	SenderUsername string     `json:"sender_username"`
	Content        string     `json:"content"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
}

type PrivateMessage struct {
	Message     Message `json:"message"`
	IsCreatedBy bool    `json:"isCreatedBy"`
}

func InsertMessage(content string, user_id_from int, chatUUID string) error {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes
	tx, err := db.Begin()
	if err != nil {
		fmt.Println("db error in InsertMessage", err)
		return err
	}
	chatID, updateErr := UpdateChat(chatUUID, user_id_from, tx)

	if updateErr != nil {
		fmt.Println("update error in InsertMessage", updateErr)
		tx.Rollback()
		return updateErr
	}
	insertQuery := `INSERT INTO messages (chat_id, user_id_from, content) VALUES (?, ?, ?);`
	_, insertErr := tx.Exec(insertQuery, chatID, user_id_from, content)
	if insertErr != nil {
		fmt.Println("Insert error in InsertMessage", insertErr)
		// Check if the error is a SQLite constraint violation
		tx.Rollback()
		if sqliteErr, ok := insertErr.(interface{ ErrorCode() int }); ok {
			if sqliteErr.ErrorCode() == 19 { // SQLite constraint violation error code
				return sql.ErrNoRows // Return custom error to indicate a duplicate
			}
		}
		return insertErr
	}
	err = tx.Commit()
	if err != nil {
		fmt.Println("Error commiting query at InsertMessage", err)
		return err
	}

	return nil
}

func UpdateMessageStatus(messageID int, status string, user_id int) error {
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

func InsertChat(user_id_1, user_id_2 int) (string, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	UUID, err := utils.GenerateUuid()
	if err != nil {
		return "", err
	}

	insertQuery := `INSERT INTO chats (uuid, user_id_1, user_id_2) VALUES (?, ?, ?);`
	_, insertErr := db.Exec(insertQuery, UUID, user_id_1, user_id_2)
	if insertErr != nil {
		// Check if the error is a SQLite constraint violation
		if sqliteErr, ok := insertErr.(interface{ ErrorCode() int }); ok {
			if sqliteErr.ErrorCode() == 19 { // SQLite constraint violation error code
				return "", sql.ErrNoRows // Return custom error to indicate a duplicate
			}
		}
		return "", insertErr
	}

	return UUID, nil
}

func UpdateChat(chatUUID string, userID int, tx *sql.Tx) (int, error) {
	// Perform the update
	query := `
		UPDATE chats
		SET updated_at = CURRENT_TIMESTAMP,
			updated_by = ?
		WHERE uuid = ?;
	`
	result, err := tx.Exec(query, userID, chatUUID)
	if err != nil {
		fmt.Println("Error, arguments:", chatUUID, userID)
		return 0, err
	}

	// Ensure at least one row was affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("rowsAffected error, arguments:", chatUUID, userID)
		return 0, err
	}
	if rowsAffected == 0 {
		fmt.Println("No rows afffected, arguments:", chatUUID, userID)
		return 0, sql.ErrNoRows // No rows were updated, meaning the UUID wasn't found
	}

	// Retrieve the chat ID separately
	var chatID int
	err = tx.QueryRow("SELECT id FROM chats WHERE uuid = ?", chatUUID).Scan(&chatID)
	if err != nil {
		fmt.Println("Error getting chat id:", chatUUID, userID, chatID)
		return 0, err
	}

	return chatID, nil
}

func UpdateChatStatus(chatID int, status string, user_id int) error {
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

type ChatUser struct {
	User         userModels.User `json:"user"`
	Username     string          `json:"username"`
	UserUUID     string          `json:"userUuid"`
	LastActivity sql.NullString  `json:"lastActivity"` // Changed to NullString
	ChatUUID     sql.NullString  `json:"chatUUID"`
	IsOnline     bool            `json:"isOnline"`
}

// ReadAllUsers retrieves all usernames: those the user has chatted with and those they haven't
func ReadAllUsers(userID int) ([]ChatUser, []ChatUser, error) {

	db := db.OpenDBConnection()
	defer db.Close()

	// Query the records
	rows, selectError := db.Query(`
SELECT u.username,
	   u.uuid,
	   u.age,
	   u.gender,
	   u.firstname,
	   u.lastname,
	   u.last_time_online,
       c.id AS chat_id,
	   c.uuid,
       COALESCE(c.updated_at, c.created_at) AS last_activity
FROM users u
LEFT JOIN chats c 
  ON (u.id = c.user_id_1 OR u.id = c.user_id_2)
  AND (c.user_id_1 = ? OR c.user_id_2 = ?)
WHERE u.id != ?
ORDER BY last_activity DESC;
    `, userID, userID, userID)

	if selectError != nil {
		fmt.Println("Select error in ReadAllUsers:", selectError)
		return nil, nil, selectError
	}
	defer rows.Close()

	var chattedUsers []ChatUser
	var notChattedUsers []ChatUser

	// Iterate over rows and collect usernames
	for rows.Next() {
		var chatID sql.NullInt64
		var chatUser ChatUser
		err := rows.Scan(&chatUser.Username, &chatUser.UserUUID, &chatUser.User.Age, &chatUser.User.Gender, &chatUser.User.FirstName, &chatUser.User.LastName, &chatUser.User.LastTimeOnline, &chatID, &chatUser.ChatUUID, &chatUser.LastActivity)
		if err != nil {
			return nil, nil, err
		}

		if chatID.Valid {
			chattedUsers = append(chattedUsers, chatUser)
		} else {
			notChattedUsers = append(notChattedUsers, chatUser)
		}
	}

	// Check for errors after iteration
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// Sort non-chatted users alphabetically
	sort.Slice(notChattedUsers, func(i, j int) bool {
		return strings.ToLower(notChattedUsers[i].Username) < strings.ToLower(notChattedUsers[j].Username)
	})

	return chattedUsers, notChattedUsers, nil
}

// findChatByUUID fetches chat ID based on UUID
func findChatByUUID(UUID string) (int, error) {
	db := db.OpenDBConnection()
	defer db.Close()

	var chatID int
	selectQuery := `SELECT id FROM chats WHERE uuid = ?;`
	err := db.QueryRow(selectQuery, UUID).Scan(&chatID)
	if err != nil {
		return 0, err
	}
	return chatID, nil
}

// ReadAllMessages retrieves the last N messages from a chat
func ReadAllMessages(chatUUID string, numberOfMessages int, userID int) ([]PrivateMessage, error) {
	var lastMessages []PrivateMessage

	chatID, findError := findChatByUUID(chatUUID)
	if findError != nil {
		if findError.Error() == "sql: no rows in result set" {
			return lastMessages, nil
		}
		return nil, findError
	}

	db := db.OpenDBConnection()
	defer db.Close()

	// Query messages along with the sender's username
	rows, selectError := db.Query(`
        SELECT 
            m.id AS message_id, 
            m.chat_id,
			c.uuid, 
            m.user_id_from, 
            u.username AS sender_username, 
            m.content, 
            m.status,
            m.updated_at, 
            m.created_at
        FROM messages m
        INNER JOIN chats c 
            ON c.id = m.chat_id
        INNER JOIN users u 
            ON m.user_id_from = u.id
        WHERE m.chat_id = ?
        ORDER BY m.id DESC
        LIMIT ?;
    `, chatID, numberOfMessages)

	if selectError != nil {
		fmt.Println("Select error at ReadAllMessages:", selectError)
		return nil, selectError
	}
	defer rows.Close()

	// Iterate over rows and collect messages
	for rows.Next() {
		var message PrivateMessage

		err := rows.Scan(&message.Message.ID, &message.Message.ChatID, &message.Message.ChatUUID, &message.Message.UserIDFrom, &message.Message.SenderUsername, &message.Message.Content, &message.Message.Status, &message.Message.UpdatedAt, &message.Message.CreatedAt)
		if err != nil {
			return nil, err
		}
		if message.Message.UserIDFrom == userID {
			message.IsCreatedBy = true
		}
		lastMessages = append(lastMessages, message)
	}

	// Check for errors after iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lastMessages, nil
}

func FindChatUUIDbyUserIDS(userID1, userID2 int) (string, error) {
	db := db.OpenDBConnection()
	defer db.Close()

	var chatUUID string

	err := db.QueryRow(`
		SELECT uuid 
		FROM chats 
		WHERE 
		    (user_id_1 = ? AND user_id_2 = ?) 
		    OR 
		    (user_id_2 = ? AND user_id_1 = ?);
	`, userID1, userID2, userID1, userID2).Scan(&chatUUID)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return chatUUID, nil
}
