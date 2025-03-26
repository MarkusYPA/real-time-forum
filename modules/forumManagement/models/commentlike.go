package models

import (
	"real-time-forum/db"
	userManagementModels "real-time-forum/modules/userManagement/models"
	"time"
)

type CommentLike struct {
	ID        int                       `json:"id"`
	Type      string                    `json:"type"`
	UserId    int                       `json:"user_id"`
	CommentId int                       `json:"comment_id"`
	Status    string                    `json:"status"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt *time.Time                `json:"updated_at"`
	UpdatedBy *int                      `json:"updated_by"`
	Post      Post                      `json:"post"`
	User      userManagementModels.User `json:"user"`
	Comment   Comment                   `json:"comment"`
}

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

func UpdateCommentLike(Type string, commentLike CommentLike) error {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	updateQuery := `UPDATE comment_likes
	SET type = ?,
		updated_at = CURRENT_TIMESTAMP,
		updated_by = ?
	WHERE id = ?;`
	_, insertErr := db.Exec(updateQuery, Type, commentLike.UserId, commentLike.ID)
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

func ReadAllCommentsLikedByUserId(userId int, Type string) ([]Comment, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	selectQuery := `SELECT 
			p.id AS post_id, p.uuid AS post_uuid, p.title AS post_title, p.description AS post_description, 
			p.status AS post_status, p.created_at AS post_created_at, p.updated_at AS post_updated_at, p.updated_by AS post_updated_by,
			c.id AS comment_id, c.user_id AS comment_user_id, c.description AS comment_description, 
			c.status AS comment_status, c.created_at AS comment_created_at, c.updated_at AS comment_updated_at, c.updated_by AS comment_updated_by,
			u.id AS user_id, u.uuid AS user_uuid, u.username AS user_username, u.type AS user_type, u.email AS user_email,  
			u.status AS user_status, u.created_at AS user_created_at, u.updated_at AS user_updated_at, u.updated_by AS user_updated_by,
			cl.id AS comment_likes_id, cl.type AS comment_likes_type, cl.comment_id AS comment_likes_comment_id, cl.user_id AS comment_likes_user_id, cl.status AS comment_likes_status, cl.created_at AS comment_likes_created_at, cl.updated_at AS comment_likes_updated_at, cl.updated_by AS comment_likes_updated_by 
		FROM comment_likes cl
			INNER JOIN comments c
				ON cl.comment_id = c.id AND cl.user_id = ? AND cl.type = ? c.status != 'delete' AND cl.status != 'delete' 
			INNER JOIN posts p 
				ON c.post_id = p.id AND p.status != 'delete' 
			INNER JOIN users u 
				ON cl.user_id = u.id AND u.status != 'delete;'		
	`
	rows, insertErr := db.Query(selectQuery, userId, Type)
	if insertErr != nil {
		// Check if the error is a SQLite constraint violation
		return nil, insertErr
	}

	var comments []Comment

	for rows.Next() {
		var comment Comment
		var user userManagementModels.User
		var post Post
		err := rows.Scan(
			&post.ID,
			&post.UUID,
			&post.Title,
			&post.Description,
			&post.Status,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.UpdatedBy,

			&user.ID,
			&user.UUID,
			&user.Username,
			&user.Type,
			&user.Email,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.UpdatedBy,

			&comment.ID,
			&comment.PostId,
			&comment.Description,
			&comment.UserId,
			&comment.Status,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&comment.UpdatedBy,
		)

		if err != nil {
			return nil, err
		}
		//comment.Post = post
		comment.User = user

		comments = append(comments, comment)
	}

	// Check for any errors during the iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil

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
