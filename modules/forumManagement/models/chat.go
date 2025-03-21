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


func findChatID(UUID string) (int, error){
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	selectQuery := `
		SELECT 
			id
		FROM chats
			WHERE uuid = ?;
	`
	id, selectError := db.Query(selectQuery, UUID)
	if selectError != nil {
		return nil, selectError
	}
	return id, nil
}

func ReadAllUsers(UUID string) (string,error){
	userID, findError = findUserByUUID(UUID)
	if findError != nil {
		return "", findError
	}

	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	// Query the records
	rows, selectError := db.Query(`
SELECT u.username
FROM users u
JOIN chats c 
  ON u.id = c.user_id_1 OR u.id = c.user_id_2
WHERE (c.user_id_1 = (SELECT id FROM users WHERE uuid = 'specific_uuid')
       OR c.user_id_2 = (SELECT id FROM users WHERE uuid = 'specific_uuid'))
  AND u.uuid != 'specific_uuid';
    `)
	if selectError != nil {
		return nil, selectError
	}
	defer rows.Close()

	var posts []Post
	// Map to track posts by their ID to avoid duplicates
	postMap := make(map[int]*Post)

	for rows.Next() {
		var post Post
		var user userManagementModels.User
		var category Category

		// Scan the post and user data
		err := rows.Scan(
			&post.ID, &post.UUID, &post.Title, &post.Description, &post.Status,
			&post.CreatedAt, &post.UpdatedAt, &post.UpdatedBy, &post.UserId,
			&user.Name, &user.Username, &user.Email,
			&category.ID, &category.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		// Check if the post already exists in the postMap
		if existingPost, found := postMap[post.ID]; found {
			// If the post exists, append the category to the existing post's Categories
			existingPost.Categories = append(existingPost.Categories, category)
		} else {
			// If the post doesn't exist in the map, add it and initialize the Categories field
			post.User = user
			post.Categories = []Category{category}
			postMap[post.ID] = &post
		}
	}

	// Check for any errors during row iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	// Convert the map of posts into a slice
	for _, post := range postMap {
		posts = append(posts, *post)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].ID > posts[j].ID
	})

	return posts, nil
}

function ReadAllChattedUsers(userID int) ([]int, error){
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	var comments []Comment
	commentMap := make(map[int]User)
	// Updated query to join comments with posts
	selectQuery := `
SELECT DISTINCT 
    CASE 
        WHEN c.user_id_1 = ? THEN c.user_id_2
        ELSE c.user_id_1
    END AS chatted_user_id
FROM chats c
WHERE c.user_id_1 = ? OR c.user_id_2 = ?;

	`
	rows, selectError := db.Query(selectQuery, userID, userID, userID) // Query the database
	if selectError != nil {
		return nil, selectError
	}
	defer rows.Close() // Ensure rows are closed after processing
	// Iterate over rows and populate the slice
	for rows.Next() {
		var comment Comment
		var user userManagementModels.User

		err := rows.Scan(
			// Map post fields
			&user.ID,
			&user.UUID,
			&user.Username,
			&user.Name,
			&user.Type,
			&user.Email,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.UpdatedBy,

			// Map comment fields
			&comment.ID,
			&comment.PostId,
			&comment.UserId,
			&comment.Description,
			&comment.Status,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&comment.UpdatedBy,

			&comment.NumberOfLikes, &comment.NumberOfDislikes,
			&comment.IsLikedByUser, &comment.IsDislikedByUser,
		)
		comment.User = user
		if err != nil {
			return nil, err
		}

		_, found := commentMap[comment.ID]
		if !found {
			commentMap[comment.ID] = &comment
		}

	}

	// Check for any errors during the iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Convert the map of comments into a slice
	for _, comment := range commentMap {
		comments = append(comments, *comment)
	}

	return comments, nil
}