package models

import (
	"real-time-forum/db"
)

func InsertCommentLike(Type string, commentId int, userId int) error {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	insertQuery := `INSERT INTO comment_likes (type, user_id, comment_id) VALUES (?, ?, ?);`
	_, insertErr := db.Exec(insertQuery, Type, userId, commentId)
	if insertErr != nil {
		// Check if the error is a SQLite constraint violation
		return insertErr
	}
	return nil
}

func UpdateCommentLikesStatus(commentLikeId int, status string, user_id int) error {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	updateQuery := `UPDATE comment_likes
	SET status = ?,
		updated_at = CURRENT_TIMESTAMP,
		updated_by = ?
	WHERE id = ?;`
	_, insertErr := db.Exec(updateQuery, status, user_id, commentLikeId)
	if insertErr != nil {
		// Check if the error is a SQLite constraint violation
		return insertErr
	}
	return nil
}

func CommentHasLiked(userId int, commentID int) (int, string) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes
	var existingLikeId int
	var existingLikeType string
	likeCheckQuery := `SELECT id, type
		FROM comment_likes cl
		WHERE cl.user_id = ? AND cl.comment_id = ?
		AND status = 'enable'
	`
	err := db.QueryRow(likeCheckQuery, userId, commentID).Scan(&existingLikeId, &existingLikeType)

	if err == nil { //it means that post has like or dislike
		return existingLikeId, existingLikeType
	} else {
		return -1, ""
	}
}
