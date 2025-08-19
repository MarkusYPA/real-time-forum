package models

import (
	"errors"
	"log"
	"real-time-forum/db"
	userManagementModels "real-time-forum/modules/userManagement/models"
	"time"
)

type PostLike struct {
	ID        int                       `json:"id"`
	Type      string                    `json:"Type"`
	PostId    int                       `json:"post_id"`
	UserId    int                       `json:"user_id"`
	Status    string                    `json:"status"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt *time.Time                `json:"updated_at"`
	UpdatedBy *int                      `json:"updated_by"`
	User      userManagementModels.User `json:"user"`
	Post      Post                      `json:"post"`
}

func InsertPostLike(postLike *PostLike) (int, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	insertQuery := `INSERT INTO post_likes (type, post_id, user_id) VALUES (?, ?, ?);`
	result, insertErr := db.Exec(insertQuery, postLike.Type, postLike.PostId, postLike.UserId)
	if insertErr != nil {
		// Check if the error is a SQLite constraint violation
		var ErrDuplicatePostLike = errors.New("duplicate post like")
		if sqliteErr, ok := insertErr.(interface{ ErrorCode() int }); ok {
			// if sqliteErr.ErrorCode() == 19 { // SQLite constraint violation error code
			// 	return -1, sql.ErrNoRows // Return custom error to indicate a duplicate
			// }
			if sqliteErr.ErrorCode() == 19 {
				return -1, ErrDuplicatePostLike
			}
		}

		return -1, insertErr
	}

	// Retrieve the last inserted ID
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		log.Fatal(err)
		return -1, err
	}

	return int(lastInsertID), nil
}

func UpdateStatusPostLike(post_like_id int, status string, user_id int) error {

	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	updateQuery := `UPDATE post_likes
		               SET status = ?,
			           updated_at = CURRENT_TIMESTAMP,
			           updated_by = ?
		               WHERE id = ?;`
	_, updateErr := db.Exec(updateQuery, status, user_id, post_like_id)
	if updateErr != nil {
		return updateErr
	}
	return nil
}

func PostHasLike(userId int, postID int) (int, string) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes
	var existingLikeId int
	var existingLikeType string
	likeCheckQuery := `SELECT id, type
		FROM post_likes pl
		WHERE pl.user_id = ? AND pl.post_id = ?
		AND status = 'enable'
	`
	err := db.QueryRow(likeCheckQuery, userId, postID).Scan(&existingLikeId, &existingLikeType)

	if err == nil { //  post has like or dislike
		return existingLikeId, existingLikeType
	} else {
		return -1, ""
	}
}
