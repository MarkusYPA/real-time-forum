package models

import (
	"database/sql"
	"fmt"
	"log"
	"real-time-forum/db"
	userManagementModels "real-time-forum/modules/userManagement/models"
	"sort"
	"time"
)

type Comment struct {
	ID               int        `json:"id"`
	PostId           int        `json:"post_id"`
	CommentId        int        `json:"comment_id"`
	Description      string     `json:"description"`
	UserId           int        `json:"user_id"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at"`
	UpdatedBy        *int       `json:"updated_by"`
	IsLikedByUser    bool       `json:"liked"`
	IsDislikedByUser bool       `json:"disliked"`
	NumberOfLikes    int        `json:"number_of_likes"`
	NumberOfDislikes int        `json:"number_of_dislikes"`
	//Post             Post                      `json:"post"`
	User         userManagementModels.User `json:"user"`
	RepliesCount int                       `json:"repliesCount"`
}

func InsertComment(postId int, commentID int, userId int, description string) (int, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	parentId := commentID
	var insertQuery string
	if postId == 0 {
		insertQuery = `INSERT INTO comments  (comment_id ,description, user_id) VALUES ( ?, ?, ?);`
	} else {
		insertQuery = `INSERT INTO comments (post_id,description, user_id) VALUES (?, ?, ?);`
		parentId = postId
	}
	result, insertErr := db.Exec(insertQuery, parentId, description, userId)

	if insertErr != nil {
		return -1, insertErr
	}

	// Retrieve the last inserted ID
	lastInsertID, errFind := result.LastInsertId()
	if errFind != nil {
		log.Println(errFind)
		return -1, errFind
	}

	return int(lastInsertID), nil
}

func ReadAllCommentsForComment(commentId int, userID int) ([]Comment, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Ensure the connection is closed after the function finishes

	var comments []Comment
	commentMap := make(map[int]*Comment)

	// Updated query to retrieve replies to a specific comment
	selectQuery := `
		SELECT 
			u.id AS user_id, u.uuid AS user_uuid, u.username AS user_username, u.type AS user_type, u.email AS user_email,  
			u.status AS user_status, u.created_at AS user_created_at, u.updated_at AS user_updated_at, u.updated_by AS user_updated_by,
			c.id,c.comment_id AS comment_id, c.post_id as comment_post_id ,c.user_id AS comment_user_id, c.description AS comment_description, 
			c.status AS comment_status, c.created_at AS comment_created_at, c.updated_at AS comment_updated_at, c.updated_by AS comment_updated_by,
			(SELECT COUNT(DISTINCT id) from comment_likes WHERE comment_id = c.id AND status != 'delete' AND type = 'like') AS number_of_likes,
			(SELECT COUNT(DISTINCT id) from comment_likes WHERE comment_id = c.id AND status != 'delete' AND type = 'dislike') AS number_of_dislikes,
			CASE 
                WHEN EXISTS (SELECT 1 FROM comment_likes WHERE comment_id = c.id AND status != 'delete' AND type = 'like' AND user_id = ?) THEN 1
                ELSE 0
            END AS is_liked_by_user,
            CASE 
                WHEN EXISTS (SELECT 1 FROM comment_likes WHERE comment_id = c.id AND status != 'delete' AND type = 'dislike' AND user_id = ?) THEN 1
                ELSE 0
            END AS is_disliked_by_user
		FROM comments c
			INNER JOIN users u
				ON c.user_id = u.id AND c.status != 'delete' AND u.status != 'delete' AND c.comment_id = ?;
	`
	rows, selectError := db.Query(selectQuery, userID, userID, commentId) // Query the database
	if selectError != nil {
		return nil, selectError
	}
	defer rows.Close() // Ensure rows are closed after processing

	// Iterate over rows and populate the slice
	for rows.Next() {
		var comment Comment
		var user userManagementModels.User
		var postId sql.NullInt64
		var commentID sql.NullInt64
		err := rows.Scan(
			// Map user fields
			&user.ID,
			&user.UUID,
			&user.Username,
			&user.Type,
			&user.Email,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.UpdatedBy,

			// Map comment fields
			&comment.ID,
			&commentID,
			&postId, // Parent comment ID
			&comment.UserId,
			&comment.Description,
			&comment.Status,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&comment.UpdatedBy,
			&comment.NumberOfLikes,
			&comment.NumberOfDislikes,
			&comment.IsLikedByUser,
			&comment.IsDislikedByUser,
		)
		comment.User = user
		if err != nil {
			return nil, err
		}
		comment.CommentId = 0
		comment.PostId = 0
		if postId.Valid {
			comment.PostId = int(postId.Int64)
		} else {
			comment.CommentId = int(commentID.Int64)
		}
		comment.RepliesCount, err = CountCommentsForComment(comment.ID)
		if err != nil {
			return nil, err
		}
		// Handle likes/dislikes aggregation
		if _, found := commentMap[comment.ID]; !found {
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
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].ID < comments[j].ID
	})

	return comments, nil
}

func CountCommentsForComment(commentID int) (int, error) {
	db := db.OpenDBConnection()
	defer db.Close()

	var numberOfComments int
	err := db.QueryRow("SELECT COUNT(*) FROM comments WHERE comment_id = ?", commentID).Scan(&numberOfComments)
	if err != nil {
		return 0, err
	}

	return numberOfComments, nil
}

func CountCommentsForPost(postID int) (int, error) {
	db := db.OpenDBConnection()
	defer db.Close()

	var numberOfComments int
	err := db.QueryRow("SELECT COUNT(*) FROM comments WHERE post_id = ?", postID).Scan(&numberOfComments)
	if err != nil {
		return 0, err
	}

	return numberOfComments, nil
}

func ReadAllCommentsForPostByUserID(postId int, userID int) ([]Comment, error) {
	db := db.OpenDBConnection()
	defer db.Close()

	var comments []Comment
	commentMap := make(map[int]*Comment)

	// Updated query to join comments with posts
	selectQuery := `
		SELECT 
			u.id AS user_id, u.uuid AS user_uuid, u.username AS user_username, u.type AS user_type, u.email AS user_email,  
			u.status AS user_status, u.created_at AS user_created_at, u.updated_at AS user_updated_at, u.updated_by AS user_updated_by,
			c.id,c.comment_id AS comment_id, c.post_id as comment_post_id ,c.user_id AS comment_user_id, c.description AS comment_description, 
			c.status AS comment_status, c.created_at AS comment_created_at, c.updated_at AS comment_updated_at, c.updated_by AS comment_updated_by,
			(SELECT COUNT(DISTINCT id) from comment_likes WHERE comment_id = c.id AND status != 'delete' AND type = 'like') AS number_of_likes,
			(SELECT COUNT(DISTINCT id) from comment_likes WHERE comment_id = c.id AND status != 'delete' AND type = 'dislike') AS number_of_dislikes,
			CASE 
                WHEN EXISTS (SELECT 1 FROM comment_likes WHERE comment_id = c.id AND status != 'delete' AND type = 'like' AND user_id = ?) THEN 1
                ELSE 0
            END AS is_liked_by_user,
            CASE 
                WHEN EXISTS (SELECT 1 FROM comment_likes WHERE comment_id = c.id AND status != 'delete' AND type = 'dislike' AND user_id = ?) THEN 1
                ELSE 0
            END AS is_disliked_by_user
		FROM comments c
			INNER JOIN users u
				ON c.user_id = u.id AND c.status != 'delete' AND u.status != 'delete' AND c.post_id = ?;
	`
	rows, selectError := db.Query(selectQuery, userID, userID, postId) // Query the database
	if selectError != nil {
		return nil, selectError
	}

	defer rows.Close() // Ensure rows are closed after processing

	// Iterate over rows and populate the slice
	for rows.Next() {
		var comment Comment
		var user userManagementModels.User
		var postID sql.NullInt64
		var commentID sql.NullInt64
		err := rows.Scan(
			// Map post fields
			&user.ID,
			&user.UUID,
			&user.Username,
			&user.Type,
			&user.Email,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.UpdatedBy,

			// Map comment fields
			&comment.ID,
			&commentID,
			&postID,
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
		comment.CommentId = 0
		comment.PostId = 0
		if postID.Valid {
			comment.PostId = int(postID.Int64)
		} else {
			comment.CommentId = int(commentID.Int64)
		}
		comment.RepliesCount, err = CountCommentsForComment(comment.ID)
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
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].ID < comments[j].ID
	})

	return comments, nil
}

func ReadCommentById(commentId int, checkLikeForUser int) (Comment, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	// Query the records
	rows, selectError := db.Query(`
        SELECT c.id, c.post_id, c.comment_id, c.description, c.user_id, c.status, 
               c.created_at, c.updated_at, c.updated_by,
               (SELECT COUNT(DISTINCT id) FROM comment_likes WHERE comment_id = c.id AND status != 'delete' AND type = 'like') AS number_of_likes,
               (SELECT COUNT(DISTINCT id) FROM comment_likes WHERE comment_id = c.id AND status != 'delete' AND type = 'dislike') AS number_of_dislikes,
               u.id as user_id, u.username, u.email,
               (SELECT COUNT(id) FROM comments WHERE comment_id = c.id AND status != 'delete') AS replies_count,
               CASE 
                   WHEN EXISTS (SELECT 1 FROM comment_likes WHERE comment_id = c.id AND status != 'delete' AND type = 'like' AND user_id = ?) THEN 1
                   ELSE 0
               END AS is_liked_by_user,
               CASE 
                   WHEN EXISTS (SELECT 1 FROM comment_likes WHERE comment_id = c.id AND status != 'delete' AND type = 'dislike' AND user_id = ?) THEN 1
                   ELSE 0
               END AS is_disliked_by_user
        FROM comments c
        INNER JOIN users u ON c.user_id = u.id
        WHERE c.id = ? AND c.status != 'delete' AND u.status != 'delete';
    `, checkLikeForUser, checkLikeForUser, commentId)
	if selectError != nil {
		return Comment{}, selectError
	}
	defer rows.Close()

	var comment Comment
	var user userManagementModels.User
	var postId sql.NullInt64
	var commentID sql.NullInt64
	// Scan the records
	if rows.Next() {
		err := rows.Scan(
			&comment.ID, &postId, &commentID, &comment.Description, &comment.UserId, &comment.Status,
			&comment.CreatedAt, &comment.UpdatedAt, &comment.UpdatedBy,
			&comment.NumberOfLikes, &comment.NumberOfDislikes,
			&user.ID, &user.Username, &user.Email,
			&comment.RepliesCount,
			&comment.IsLikedByUser, &comment.IsDislikedByUser,
		)
		if err != nil {
			return Comment{}, fmt.Errorf("error scanning row: %v", err)
		}
		comment.CommentId = 0
		comment.PostId = 0
		if postId.Valid {
			comment.PostId = int(postId.Int64)
		} else {
			comment.CommentId = int(commentID.Int64)
		}
		// Assign user to comment
		comment.User = user
	} else {
		// No rows returned, meaning the comment doesn't exist
		return Comment{}, fmt.Errorf("comment with ID %d not found", commentId)
	}

	// Check for any errors during row iteration
	if err := rows.Err(); err != nil {
		return Comment{}, fmt.Errorf("row iteration error: %v", err)
	}

	return comment, nil
}
