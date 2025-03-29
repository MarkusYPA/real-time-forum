package models

import (
	"database/sql"
	"fmt"
	"log"
	"real-time-forum/db"
	userManagementModels "real-time-forum/modules/userManagement/models"
	"real-time-forum/utils"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Post struct represents the user data model
type Post struct {
	ID               int                       `json:"id"`
	UUID             string                    `json:"uuid"`
	Title            string                    `json:"title"`
	Description      string                    `json:"description"`
	UserId           int                       `json:"user_id"`
	Status           string                    `json:"status"`
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        *time.Time                `json:"updated_at"`
	UpdatedBy        *int                      `json:"updated_by"`
	IsLikedByUser    bool                      `json:"liked"`
	IsDislikedByUser bool                      `json:"disliked"`
	NumberOfLikes    int                       `json:"number_of_likes"`
	NumberOfDislikes int                       `json:"number_of_dislikes"`
	User             userManagementModels.User `json:"user"`       // Embedded user data
	Categories       []Category                `json:"categories"` // List of categories related to the post
	RepliesCount     int                       `json:"repliesCount"`
}

func InsertPost(post *Post, categoryIds []int) (int, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	// Start a transaction for atomicity
	tx, err := db.Begin()
	if err != nil {
		return -1, err
	}

	post.UUID, err = utils.GenerateUuid()
	if err != nil {
		tx.Rollback() // Rollback if UUID generation fails
		return -1, err
	}

	insertQuery := `INSERT INTO posts (uuid, title, description, user_id) VALUES (?, ?, ?, ?);`
	result, insertErr := tx.Exec(insertQuery, post.UUID, post.Title, post.Description, post.User.ID)
	if insertErr != nil {
		tx.Rollback() // Rollback on error
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
		tx.Rollback() // Rollback on error
		log.Println(err)
		return -1, err
	}

	insertPostCategoriesErr := InsertPostCategories(int(lastInsertID), categoryIds, post.User.ID, tx)
	if insertPostCategoriesErr != nil {
		fmt.Println("err insert cats:", insertPostCategoriesErr)
		tx.Rollback() // Rollback on error
		return -1, insertPostCategoriesErr
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		tx.Rollback() // Rollback on error
		return -1, err
	}

	return int(lastInsertID), nil
}

func UpdatePost(post *Post, categories []int, user_id int) error {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	// Start a transaction for atomicity
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	updateQuery := `UPDATE posts
					SET title = ?,
						description = ?,
						updated_at = CURRENT_TIMESTAMP,
						updated_by = ?
					WHERE id = ?;`
	_, updateErr := tx.Exec(updateQuery, post.Title, post.Description, user_id, post.ID)
	if updateErr != nil {
		tx.Rollback() // Rollback on error
		// Check if the error is a SQLite constraint violation
		if sqliteErr, ok := updateErr.(interface{ ErrorCode() int }); ok {
			if sqliteErr.ErrorCode() == 19 { // SQLite constraint violation error code
				return sql.ErrNoRows // Return custom error to indicate a duplicate
			}
		}
		return updateErr
	}

	deletePostCategoriesErr := UpdateStatusPostCategories(post.ID, user_id, "delete", tx)
	if deletePostCategoriesErr != nil {
		tx.Rollback() // Rollback on error
		return deletePostCategoriesErr
	}

	insertPostCategoriesErr := InsertPostCategories(post.ID, categories, user_id, tx)
	if insertPostCategoriesErr != nil {
		tx.Rollback() // Rollback on error
		return insertPostCategoriesErr
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		tx.Rollback() // Rollback on error
		return err
	}

	return nil
}

func UpdateStatusPost(post_id int, status string, user_id int) error {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	// Start a transaction for atomicity
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	updateQuery := `UPDATE posts
					SET status = ?,
						updated_at = CURRENT_TIMESTAMP,
						updated_by = ?
					WHERE id = ?;`
	_, updateErr := tx.Exec(updateQuery, status, user_id, post_id)
	if updateErr != nil {
		tx.Rollback() // Rollback on error
		// Check if the error is a SQLite constraint violation
		if sqliteErr, ok := updateErr.(interface{ ErrorCode() int }); ok {
			if sqliteErr.ErrorCode() == 19 { // SQLite constraint violation error code
				return sql.ErrNoRows // Return custom error to indicate a duplicate
			}
		}
		return updateErr
	}

	updateStatusPostCategories := UpdateStatusPostCategories(post_id, user_id, status, tx)
	if updateStatusPostCategories != nil {
		tx.Rollback() // Rollback on error
		return updateStatusPostCategories
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		tx.Rollback() // Rollback on error
		return err
	}

	return nil
}

func ReadAllPosts(userId int) ([]Post, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	// Query the records
	/* 	rows, selectError := db.Query(`
	        SELECT p.id as post_id, p.uuid as post_uuid, p.title as post_title, p.description as post_description, p.status as post_status, p.created_at as post_created_at, p.updated_at as post_updated_at, p.updated_by as post_updated_by,
				u.id as user_id, u.username as user_username, u.email as user_email,
				c.id as category_id, c.name as category_name
			FROM posts p
				INNER JOIN users u
					ON p.user_id = u.id
				LEFT JOIN post_categories pc
					ON p.id = pc.post_id
					AND pc.status = 'enable'
				LEFT JOIN categories c
					ON pc.category_id = c.id
					AND c.status = 'enable'
			WHERE p.status != 'delete'
				AND u.status != 'delete'
			ORDER BY p.id desc;
	    `) */

	// Query the records
	rows, selectError := db.Query(`
       SELECT 
    p.id as post_id, 
    p.uuid as post_uuid, 
    p.title as post_title, 
    p.description as post_description, 
    p.status as post_status, 
    p.created_at as post_created_at, 
    p.updated_at as post_updated_at, 
    p.updated_by as post_updated_by,
    (SELECT COUNT(DISTINCT id) 
     FROM post_likes 
     WHERE post_id = p.id AND status != 'delete' AND type = 'like') AS number_of_likes,
    (SELECT COUNT(DISTINCT id) 
     FROM post_likes 
     WHERE post_id = p.id AND status != 'delete' AND type = 'dislike') AS number_of_dislikes,
    (SELECT COUNT(*) 
     FROM comments 
     WHERE post_id = p.id
    ) AS number_of_comments,
    u.id as user_id, 
    u.username as user_username, 
    u.email as user_email,
    c.id as category_id, 
    c.name as category_name,
    CASE 
        WHEN EXISTS (SELECT 1 FROM post_likes WHERE post_id = p.id AND status != 'delete' AND type = 'like' AND user_id = ?) THEN 1
        ELSE 0
    END AS is_liked_by_user,
    CASE 
        WHEN EXISTS (SELECT 1 FROM post_likes WHERE post_id = p.id AND status != 'delete' AND type = 'dislike' AND user_id = ?) THEN 1
        ELSE 0
    END AS is_disliked_by_user
FROM posts p
INNER JOIN users u ON p.user_id = u.id
LEFT JOIN post_categories pc ON p.id = pc.post_id AND pc.status = 'enable'
LEFT JOIN categories c ON pc.category_id = c.id AND c.status = 'enable'
WHERE p.status != 'delete' AND u.status != 'delete';

    `, userId, userId)

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
			&post.CreatedAt, &post.UpdatedAt, &post.UpdatedBy,
			&post.NumberOfLikes, &post.NumberOfDislikes, &post.RepliesCount,
			&post.UserId, &user.Username, &user.Email,
			&category.ID, &category.Name,
			&post.IsLikedByUser, &post.IsDislikedByUser,
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
		return posts[i].ID < posts[j].ID
	})

	return posts, nil
}

func ReadPostsByCategoryId(userID int, categoryID int) ([]Post, error) {
	db := db.OpenDBConnection()
	defer db.Close()

	rows, selectError := db.Query(`
	SELECT 
		p.id AS post_id, p.uuid AS post_uuid, p.title AS post_title, 
		p.description AS post_description, p.status AS post_status, 
		p.created_at AS post_created_at, p.updated_at AS post_updated_at, 
		p.updated_by AS post_updated_by,
		
		(SELECT COUNT(*) FROM post_likes WHERE post_id = p.id AND status != 'delete' AND type = 'like') AS number_of_likes,
		(SELECT COUNT(*) FROM post_likes WHERE post_id = p.id AND status != 'delete' AND type = 'dislike') AS number_of_dislikes,
	
		u.id AS user_id, u.username AS user_username, u.email AS user_email,
	
		GROUP_CONCAT(DISTINCT c.id) AS category_ids,
		GROUP_CONCAT(DISTINCT c.name) AS category_names,
	
		CASE 
			WHEN EXISTS (SELECT 1 FROM post_likes WHERE post_id = p.id AND status != 'delete' AND type = 'like' AND user_id = ?) THEN 1
			ELSE 0
		END AS is_liked_by_user,
		CASE 
			WHEN EXISTS (SELECT 1 FROM post_likes WHERE post_id = p.id AND status != 'delete' AND type = 'dislike' AND user_id = ?) THEN 1
			ELSE 0
		END AS is_disliked_by_user,
	
		(SELECT COUNT(*) FROM comments WHERE post_id = p.id) AS number_of_comments
	
	FROM posts p
	INNER JOIN users u ON p.user_id = u.id
	LEFT JOIN post_categories pc 
		ON p.id = pc.post_id 
		AND pc.status = 'enable'
	LEFT JOIN categories c 
		ON pc.category_id = c.id 
		AND c.status = 'enable'
	WHERE p.status != 'delete' 
		AND u.status != 'delete'
		AND p.id IN (
			SELECT post_id FROM post_categories WHERE category_id = ? AND status = 'enable'
		)
	GROUP BY p.id, u.id;
	`, userID, userID, categoryID)

	if selectError != nil {
		return nil, selectError
	}
	defer rows.Close()

	var posts []Post
	postMap := make(map[int]*Post)

	for rows.Next() {
		var post Post
		var user userManagementModels.User
		var categoryIDs string
		var categoryNames string

		err := rows.Scan(
			&post.ID, &post.UUID, &post.Title, &post.Description, &post.Status,
			&post.CreatedAt, &post.UpdatedAt, &post.UpdatedBy,
			&post.NumberOfLikes, &post.NumberOfDislikes,
			&post.UserId, &user.Username, &user.Email,
			&categoryIDs, &categoryNames,
			&post.IsLikedByUser, &post.IsDislikedByUser, &post.RepliesCount,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		// Convert category strings into slices
		categoryIDList := strings.Split(categoryIDs, ",")
		categoryNameList := strings.Split(categoryNames, ",")

		var categories []Category
		for i := range categoryIDList {
			id, _ := strconv.Atoi(categoryIDList[i])
			categories = append(categories, Category{ID: id, Name: categoryNameList[i]})
		}

		post.User = user
		post.Categories = categories
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	for _, post := range postMap {
		posts = append(posts, *post)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].ID < posts[j].ID
	})

	return posts, nil
}

func FilterPosts(searchTerm string) ([]Post, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	searchPattern := "%" + searchTerm + "%" // Add wildcards for LIKE comparison

	// Query the records
	rows, selectError := db.Query(`
        SELECT p.id as post_id, p.uuid as post_uuid, p.title as post_title, p.description as post_description, p.status as post_status, p.created_at as post_created_at, p.updated_at as post_updated_at, p.updated_by as post_updated_by,
			u.id as user_id, u.username as user_username, u.email as user_email,
			c.id as category_id, c.name as category_name
		FROM posts p
			INNER JOIN users u
				ON p.user_id = u.id
			LEFT JOIN post_categories pc
				ON p.id = pc.post_id
				AND pc.status = 'enable'
			LEFT JOIN categories c
				ON pc.category_id = c.id
				AND c.status = 'enable'
		WHERE p.status != 'delete'
			AND u.status != 'delete'
      		AND (p.title LIKE ? OR p.description LIKE ?)
		ORDER BY p.id asc;
    `, searchPattern, searchPattern)
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
			&user.Username, &user.Email,
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

func ReadPostsByUserId(userId int) ([]Post, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	// Query the records
	rows, selectError := db.Query(`
        SELECT p.id as post_id, p.uuid as post_uuid, p.title as post_title, p.description as post_description, p.status as post_status, p.created_at as post_created_at, p.updated_at as post_updated_at, p.updated_by as post_updated_by,
			u.id as user_id, u.username as user_username, u.email as user_email,
			c.id as category_id, c.name as category_name
		FROM posts p
			INNER JOIN users u
				ON p.user_id = u.id
				AND u.id = ?
			LEFT JOIN post_categories pc
				ON p.id = pc.post_id
				AND pc.status = 'enable'
			LEFT JOIN categories c
				ON pc.category_id = c.id
				AND c.status = 'enable'
		WHERE p.status != 'delete'
			AND u.status != 'delete'
		ORDER BY p.id asc;
    `, userId)
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
			&user.Username, &user.Email,
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

func ReadPostsLikedByUserId(userId int) ([]Post, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	// Query the records
	rows, selectError := db.Query(`
        SELECT p.id as post_id, p.uuid as post_uuid, p.title as post_title, p.description as post_description, p.status as post_status, p.created_at as post_created_at, p.updated_at as post_updated_at, p.updated_by as post_updated_by,
			u.id as user_id, u.username as user_username, u.email as user_email,
			c.id as category_id, c.name as category_name
		FROM posts p
			INNER JOIN post_likes pl
				ON pl.post_id = p.id
				AND pl.status = 'enable'
			INNER JOIN users u
				ON p.user_id = u.id
			INNER JOIN users liked_user
				ON pl.user_id = liked_user.id
				AND liked_user.id = ?
			LEFT JOIN post_categories pc
				ON p.id = pc.post_id
				AND pc.status = 'enable'
			LEFT JOIN categories c
				ON pc.category_id = c.id
				AND c.status = 'enable'
		WHERE p.status != 'delete'
			AND u.status != 'delete'
		ORDER BY p.id asc;
    `, userId)
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
			&user.Username, &user.Email,
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

func ReadPostById(postId int, checkLikeForUser int) (Post, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	// Query the records
	rows, selectError := db.Query(`
        SELECT p.id as post_id, p.uuid as post_uuid, p.title as post_title, p.description as post_description, p.status as post_status, p.created_at as post_created_at, p.updated_at as post_updated_at, p.updated_by as post_updated_by,
			(SELECT COUNT(DISTINCT id) from post_likes WHERE post_id = p.id AND status != 'delete' AND type = 'like') AS number_of_likes,
			(SELECT COUNT(DISTINCT id) from post_likes WHERE post_id = p.id AND status != 'delete' AND type = 'dislike') AS number_of_dislikes,
			u.id as user_id, u.username as user_username, u.email as user_email,
			c.id as category_id, c.name as category_name,
			CASE 
                WHEN EXISTS (SELECT 1 FROM post_likes WHERE post_id = p.id AND status != 'delete' AND type = 'like' AND user_id = ?) THEN 1
                ELSE 0
            END AS is_liked_by_user,
            CASE 
                WHEN EXISTS (SELECT 1 FROM post_likes WHERE post_id = p.id AND status != 'delete' AND type = 'dislike' AND user_id = ?) THEN 1
                ELSE 0
            END AS is_disliked_by_user
		FROM posts p
			INNER JOIN users u
				ON p.user_id = u.id
				AND p.id = ?
			LEFT JOIN post_categories pc
				ON p.id = pc.post_id
				AND pc.status = 'enable'
			LEFT JOIN categories c
				ON pc.category_id = c.id
				AND c.status = 'enable'
		WHERE p.status != 'delete'
			AND u.status != 'delete';
    `, checkLikeForUser, checkLikeForUser, postId)
	if selectError != nil {
		return Post{}, selectError
	}
	defer rows.Close()

	var post Post
	var user userManagementModels.User
	var categories []Category

	// Scan the records
	for rows.Next() {
		var category Category

		err := rows.Scan(
			&post.ID, &post.UUID, &post.Title, &post.Description, &post.Status,
			&post.CreatedAt, &post.UpdatedAt, &post.UpdatedBy,
			&post.NumberOfLikes, &post.NumberOfDislikes,
			&post.UserId, &user.Username, &user.Email,
			&category.ID, &category.Name,
			&post.IsLikedByUser, &post.IsDislikedByUser,
		)
		if err != nil {
			return Post{}, fmt.Errorf("error scanning row: %v", err)
		}

		// Assign user to post
		if post.UserId == 0 { // If this is the first time we're encountering the post
			post.User = user
		}

		// Append category to post categories list
		categories = append(categories, category)
	}

	// If no rows were returned, the post doesn't exist
	if post.ID == 0 {
		return Post{}, fmt.Errorf("post with ID %d not found", postId)
	}

	// Assign categories to the post
	post.Categories = categories

	// Check for any errors during row iteration
	if err := rows.Err(); err != nil {
		return Post{}, fmt.Errorf("row iteration error: %v", err)
	}

	return post, nil
}

func ReadPostByUUID(postUUID string, checkLikeForUser int) (Post, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	// Query the records
	rows, selectError := db.Query(`
        SELECT p.id as post_id, p.uuid as post_uuid, p.title as post_title, p.description as post_description, p.status as post_status, p.created_at as post_created_at, p.updated_at as post_updated_at, p.updated_by as post_updated_by,
			(SELECT COUNT(DISTINCT id) from post_likes WHERE post_id = p.id AND status != 'delete' AND type = 'like') AS number_of_likes,
			(SELECT COUNT(DISTINCT id) from post_likes WHERE post_id = p.id AND status != 'delete' AND type = 'dislike') AS number_of_dislikes,
			u.id as user_id, u.username as user_username, u.email as user_email,
			c.id as category_id, c.name as category_name,
			CASE 
                WHEN EXISTS (SELECT 1 FROM post_likes WHERE post_id = p.id AND status != 'delete' AND type = 'like' AND user_id = ?) THEN 1
                ELSE 0
            END AS is_liked_by_user,
            CASE 
                WHEN EXISTS (SELECT 1 FROM post_likes WHERE post_id = p.id AND status != 'delete' AND type = 'dislike' AND user_id = ?) THEN 1
                ELSE 0
            END AS is_disliked_by_user
		FROM posts p
			INNER JOIN users u
				ON p.user_id = u.id
				AND p.uuid = ?
			LEFT JOIN post_categories pc
				ON p.id = pc.post_id
				AND pc.status = 'enable'
			LEFT JOIN categories c
				ON pc.category_id = c.id
				AND c.status = 'enable'
		WHERE p.status != 'delete'
			AND u.status != 'delete';
    `, checkLikeForUser, checkLikeForUser, postUUID)
	if selectError != nil {
		return Post{}, selectError
	}
	defer rows.Close()

	var post Post
	var user userManagementModels.User
	var categories []Category

	// Scan the records
	for rows.Next() {
		var category Category

		err := rows.Scan(
			&post.ID, &post.UUID, &post.Title, &post.Description, &post.Status,
			&post.CreatedAt, &post.UpdatedAt, &post.UpdatedBy,
			&post.NumberOfLikes, &post.NumberOfDislikes,
			&post.UserId, &user.Username, &user.Email,
			&category.ID, &category.Name,
			&post.IsLikedByUser, &post.IsDislikedByUser,
		)
		if err != nil {
			return Post{}, fmt.Errorf("error scanning row: %v", err)
		}

		// Append category to post categories list
		categories = append(categories, category)
	}

	// If no rows were returned, the post doesn't exist
	if post.ID == 0 {
		return Post{}, fmt.Errorf("post with UUID %s not found", postUUID)
	}

	// Assign categories to the post
	post.Categories = categories
	post.User = user

	// Check for any errors during row iteration
	if err := rows.Err(); err != nil {
		return Post{}, fmt.Errorf("row iteration error: %v", err)
	}

	return post, nil
}

func ReadPostByUserID(postId int, userID int) (Post, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes
	// Updated query to join comments with posts
	rows, selectError := db.Query(`
        SELECT p.id as post_id, p.uuid as post_uuid, p.title as post_title, p.description as post_description, p.status as post_status, p.created_at as post_created_at, p.updated_at as post_updated_at, p.updated_by as post_updated_by,
			p.user_id as post_user_id, u.id as user_id, u.username as user_username, u.email as user_email,
			c.id as category_id, c.name as category_name,
			COALESCE(pl.type, '')
		FROM posts p
			INNER JOIN users u
				ON p.user_id = u.id
				AND p.id = ?
			LEFT JOIN post_categories pc
				ON p.id = pc.post_id
				AND pc.status = 'enable'
			LEFT JOIN categories c
				ON pc.category_id = c.id
				AND c.status = 'enable'
			LEFT JOIN post_likes pl
				ON p.id = pl.post_id AND pl.status != 'delete'	
		WHERE p.status != 'delete'
			AND u.status != 'delete';
    `, postId)
	if selectError != nil {
		return Post{}, selectError
	}
	defer rows.Close()

	var post Post
	var user userManagementModels.User
	var categories []Category

	// Scan the records
	for rows.Next() {
		var category Category
		var Type string
		err := rows.Scan(
			&post.ID, &post.UUID, &post.Title, &post.Description, &post.Status,
			&post.CreatedAt, &post.UpdatedAt, &post.UpdatedBy, &post.UserId,
			&user.ID, &user.Username, &user.Email,
			&category.ID, &category.Name, &Type,
		)
		if err != nil {
			return Post{}, fmt.Errorf("error scanning row: %v", err)
		}
		if user.ID == userID {
			if Type == "like" {
				post.IsLikedByUser = true
			} else if Type == "dislike" {
				post.IsDislikedByUser = true
			}
		}
		if Type == "like" {
			post.NumberOfLikes++
		} else if Type == "dislike" {
			post.NumberOfDislikes++
		}
		// Append category to post categories list
		categories = append(categories, category)
	}

	// Assign categories to the post
	post.Categories = categories
	post.User = user

	// Check for any errors during row iteration
	if err := rows.Err(); err != nil {
		return Post{}, fmt.Errorf("row iteration error: %v", err)
	}

	return post, nil
}
