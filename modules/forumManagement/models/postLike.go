package models

import (
	"errors"
	"fmt"
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

func ReadAllPostsLikes() ([]PostLike, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	// Query the records
	rows, selectError := db.Query(`
        SELECT 
			pl.id as post_like_id, pl.type, pl.status as post_like_status, pl.created_at as post_like_created_at, pl.updated_at as post_like_updated_at, pl.updated_by as post_like_updated_by,
			p.id as post_id, p.status as post_status, p.created_at as post_created_at, p.updated_at as post_updated_at, p.updated_by as post_updated_by,
			u.id as user_id, u.username as user_username, u.email as user_email,
			c.id as category_id, c.name as category_name
		FROM post_likes pl
			INNER JOIN posts p
				ON pl.post_id = p.id	
				AND p.status != 'delete'
			INNER JOIN users u
				ON pl.user_id = u.id
				AND u.status != 'delete'
			LEFT JOIN post_categories pc
				ON p.id = pc.post_id
				AND pc.status = 'enable'
			LEFT JOIN categories c
				ON pc.category_id = c.id
				AND c.status = 'enable'
		WHERE pl.status != 'delete'
			;
    `)
	if selectError != nil {
		return nil, selectError
	}
	defer rows.Close()

	var postLikes []PostLike
	// Map to track postLikes by their ID to avoid duplicates
	postLikeMap := make(map[int]*PostLike)

	for rows.Next() {
		var postLike PostLike
		var post Post
		var user userManagementModels.User
		var category Category

		// Scan the post_like, post, user, and category data
		err := rows.Scan(
			&postLike.ID, &postLike.Type, &postLike.Status, &postLike.CreatedAt, &postLike.UpdatedAt, &postLike.UpdatedBy,
			&post.ID, &post.Status, &post.CreatedAt, &post.UpdatedAt, &post.UpdatedBy,
			&user.ID, &user.Username, &user.Email,
			&category.ID, &category.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		// Check if the post_like already exists in the postLikeMap
		if existingPostLike, found := postLikeMap[postLike.ID]; found {
			// If the post_like exists, append the category to the existing post's Categories
			existingPostLike.Post.Categories = append(existingPostLike.Post.Categories, category)
		} else {
			// If the post_like doesn't exist in the map, add it and initialize the Categories field
			postLike.Post = post
			postLike.User = user
			postLike.Post.Categories = []Category{category}
			postLikeMap[postLike.ID] = &postLike
		}
	}

	// Check for any errors during row iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	// Convert the map of postLikes into a slice
	for _, postLike := range postLikeMap {
		postLikes = append(postLikes, *postLike)
	}

	return postLikes, nil
}

func ReadPostsLikeByUserId(userId int) ([]PostLike, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	// Query the records
	rows, selectError := db.Query(`
        SELECT 
			pl.id as post_like_id, pl.type, pl.status as post_like_status, pl.created_at as post_like_created_at, pl.updated_at as post_like_updated_at, pl.updated_by as post_like_updated_by,
			p.id as post_id, p.status as post_status, p.created_at as post_created_at, p.updated_at as post_updated_at, p.updated_by as post_updated_by,
			u.id as user_id, u.username as user_username, u.email as user_email,
			c.id as category_id, c.name as category_name
		FROM post_likes pl
			INNER JOIN posts p
				ON pl.post_id = p.id	
				AND p.status != 'delete'
			INNER JOIN users u
				ON pl.user_id = u.id
				AND u.status != 'delete'
				AND u.id = ?
			LEFT JOIN post_categories pc
				ON p.id = pc.post_id
				AND pc.status = 'enable'
			LEFT JOIN categories c
				ON pc.category_id = c.id
				AND c.status = 'enable'
		WHERE pl.status != 'delete'
			;
    `, userId)
	if selectError != nil {
		return nil, selectError
	}
	defer rows.Close()

	var postLikes []PostLike
	// Map to track postLikes by their ID to avoid duplicates
	postLikeMap := make(map[int]*PostLike)

	for rows.Next() {
		var postLike PostLike
		var post Post
		var user userManagementModels.User
		var category Category

		// Scan the post_like, post, user, and category data
		err := rows.Scan(
			&postLike.ID, &postLike.Type, &postLike.Status, &postLike.CreatedAt, &postLike.UpdatedAt, &postLike.UpdatedBy,
			&post.ID, &post.Status, &post.CreatedAt, &post.UpdatedAt, &post.UpdatedBy,
			&user.ID, &user.Username, &user.Email,
			&category.ID, &category.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		// Check if the post_like already exists in the postLikeMap
		if existingPostLike, found := postLikeMap[postLike.ID]; found {
			// If the post_like exists, append the category to the existing post's Categories
			existingPostLike.Post.Categories = append(existingPostLike.Post.Categories, category)
		} else {
			// If the post_like doesn't exist in the map, add it and initialize the Categories field
			postLike.Post = post
			postLike.User = user
			postLike.Post.Categories = []Category{category}
			postLikeMap[postLike.ID] = &postLike
		}
	}

	// Check for any errors during row iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	// Convert the map of postLikes into a slice
	for _, postLike := range postLikeMap {
		postLikes = append(postLikes, *postLike)
	}

	return postLikes, nil
}

func ReadPostsLikeByPostId(postId int) ([]PostLike, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	// Query the records
	rows, selectError := db.Query(`
        SELECT 
			pl.id as post_like_id, pl.type, pl.status as post_like_status, pl.created_at as post_like_created_at, pl.updated_at as post_like_updated_at, pl.updated_by as post_like_updated_by,
			p.id as post_id, p.status as post_status, p.created_at as post_created_at, p.updated_at as post_updated_at, p.updated_by as post_updated_by,
			u.id as user_id, u.username as user_username, u.email as user_email,
			c.id as category_id, c.name as category_name
		FROM post_likes pl
			INNER JOIN posts p
				ON pl.post_id = p.id	
				AND p.status != 'delete'
				AND p.id = ?
			INNER JOIN users u
				ON pl.user_id = u.id
				AND u.status != 'delete'
			LEFT JOIN post_categories pc
				ON p.id = pc.post_id
				AND pc.status = 'enable'
			LEFT JOIN categories c
				ON pc.category_id = c.id
				AND c.status = 'enable'
		WHERE pl.status != 'delete'
			;
    `, postId)
	if selectError != nil {
		return nil, selectError
	}
	defer rows.Close()

	var postLikes []PostLike
	// Map to track postLikes by their ID to avoid duplicates
	postLikeMap := make(map[int]*PostLike)

	for rows.Next() {
		var postLike PostLike
		var post Post
		var user userManagementModels.User
		var category Category

		// Scan the post_like, post, user, and category data
		err := rows.Scan(
			&postLike.ID, &postLike.Type, &postLike.Status, &postLike.CreatedAt, &postLike.UpdatedAt, &postLike.UpdatedBy,
			&post.ID, &post.Status, &post.CreatedAt, &post.UpdatedAt, &post.UpdatedBy,
			&user.ID, &user.Username, &user.Email,
			&category.ID, &category.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		// Check if the post_like already exists in the postLikeMap
		if existingPostLike, found := postLikeMap[postLike.ID]; found {
			// If the post_like exists, append the category to the existing post's Categories
			existingPostLike.Post.Categories = append(existingPostLike.Post.Categories, category)
		} else {
			// If the post_like doesn't exist in the map, add it and initialize the Categories field
			postLike.Post = post
			postLike.User = user
			postLike.Post.Categories = []Category{category}
			postLikeMap[postLike.ID] = &postLike
		}
	}

	// Check for any errors during row iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	// Convert the map of postLikes into a slice
	for _, postLike := range postLikeMap {
		postLikes = append(postLikes, *postLike)
	}

	return postLikes, nil
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
