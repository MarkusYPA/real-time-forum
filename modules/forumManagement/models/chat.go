package models

import (
	"database/sql"
	"real-time-forum/db"
	"real-time-forum/utils"
	"sort"
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
	ChatID         int        `json:"chat_id"`
	UserIDFrom     int        `json:"user_id_from"`
	SenderUsername string     `json:"sender_username"`
	Content        string     `json:"content"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
}

func InsertMessage(content string, user_id_from int, chatUUID string) error {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	chatID, updateErr := UpdateChat(chatUUID, user_id_from, tx)
	if updateErr != nil {
		tx.Rollback()
		return updateErr
	}
	insertQuery := `INSERT INTO messages (chat_id, user_id_from, content) VALUES (?, ?, ?);`
	_, insertErr := tx.Exec(insertQuery, chatID, user_id_from, content)
	if insertErr != nil {
		// Check if the error is a SQLite constraint violation
		tx.Rollback()
		if sqliteErr, ok := insertErr.(interface{ ErrorCode() int }); ok {
			if sqliteErr.ErrorCode() == 19 { // SQLite constraint violation error code
				return sql.ErrNoRows // Return custom error to indicate a duplicate
			}
		}
		return insertErr
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

func InsertChat(user_id_1, user_id_2 int) (int, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	UUID, err := utils.GenerateUuid()
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

func UpdateChat(chatUUID string, userID int, tx *sql.Tx) (int, error) {
	var chatID int
	query := `
		UPDATE chats
		SET updated_at = CURRENT_TIMESTAMP,
			updated_by = ?
		WHERE uuid = ?
		RETURNING id;
	`
	err := tx.QueryRow(query, userID, chatUUID).Scan(&chatID)
	if err != nil {
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

// findUserByUUID fetches user ID based on UUID
func findUserByUUID(UUID string) (int, error) {
	db := db.OpenDBConnection()
	defer db.Close()

	var userID int
	selectQuery := `SELECT id FROM users WHERE uuid = ?;`
	err := db.QueryRow(selectQuery, UUID).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

type ChatUser struct {
	Username     string         `json:"username"`
	UserUUID     string         `json:"userUuid"`
	LastActivity sql.NullTime   `json:"lastActivity"`
	ChatUUID     sql.NullString `json:"chatUUID"`
	IsOnline     bool           `json:"isOnline"`
}

// ReadAllUsers retrieves all usernames: those the user has chatted with and those they haven't
func ReadAllUsers(userID int) ([]ChatUser, []ChatUser, error) {

	db := db.OpenDBConnection()
	defer db.Close()

	// Query the records
	rows, selectError := db.Query(`
SELECT u.username,
	   u.uuid,
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
		return nil, nil, selectError
	}
	defer rows.Close()

	var chattedUsers []ChatUser
	var notChattedUsers []ChatUser

	// Iterate over rows and collect usernames
	for rows.Next() {
		var chatID sql.NullInt64
		var chatUser ChatUser
		err := rows.Scan(&chatUser.Username, &chatUser.UserUUID, &chatID, &chatUser.ChatUUID, &chatUser.LastActivity)
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
		return notChattedUsers[i].Username < notChattedUsers[j].Username
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

type PrivateMessage struct {
	Message     Message
	IsCreatedBy bool
}

// ReadAllMessages retrieves the last N messages from a chat
func ReadAllMessages(chatUUID string, numberOfMessages int, userID int) ([]PrivateMessage, error) {
	chatID, findError := findChatByUUID(chatUUID)
	if findError != nil {
		return nil, findError
	}

	db := db.OpenDBConnection()
	defer db.Close()

	// Query messages along with the sender's username
	rows, selectError := db.Query(`
        SELECT 
            m.id AS message_id, 
            m.chat_id, 
            m.user_id_from, 
            u.username AS sender_username, 
            m.content, 
            m.status,
            m.updated_at, 
            m.created_at,  
        FROM messages m
        INNER JOIN chats c 
            ON c.id = m.chat_id
        INNER JOIN users u 
            ON m.user_id_from = u.id
        WHERE m.chat_id = ?
        ORDER BY m.created_at DESC
        LIMIT ?;
    `, chatID, numberOfMessages)

	if selectError != nil {
		return nil, selectError
	}
	defer rows.Close()

	var lastMessages []PrivateMessage

	// Iterate over rows and collect messages
	for rows.Next() {
		var message PrivateMessage

		err := rows.Scan(&message.Message.ID, &message.Message.ChatID, &message.Message.UserIDFrom, &message.Message.SenderUsername, &message.Message.Content, &message.Message.Status, &message.Message.UpdatedAt, &message.Message.CreatedAt)
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
