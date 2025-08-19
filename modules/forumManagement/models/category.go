package models

import (
	"fmt"
	"real-time-forum/db"
	userManagementModels "real-time-forum/modules/userManagement/models"
	"time"
)

// Category struct represents the user data model
type Category struct {
	ID        int                       `json:"id"`
	Name      string                    `json:"name"`
	Status    string                    `json:"status"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt *time.Time                `json:"updated_at"`
	CreatedBy int                       `json:"created_by"`
	UpdatedBy *int                      `json:"updated_by"`
	User      userManagementModels.User `json:"user"` // Embedded user data
}

func ReadAllCategories() ([]Category, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Close the connection after the function finishes

	// Query the records
	rows, selectError := db.Query(`
        SELECT c.id as category_id, c.name as category_name, c.status as category_status, 
               c.created_at as category_created_at, c.created_by as category_created_by, 
               c.updated_at as category_updated_at, c.updated_by as category_updated_by,
               u.id as user_id, u.username as user_username, u.email as user_email
        FROM categories c
        INNER JOIN users u ON c.created_by = u.id
        WHERE c.status != 'delete';
    `)
	if selectError != nil {
		return nil, selectError
	}
	defer rows.Close()

	var categories []Category

	for rows.Next() {
		var category Category
		var user userManagementModels.User

		// Scan the data into variables
		err := rows.Scan(
			&category.ID, &category.Name, &category.Status, &category.CreatedAt, &category.CreatedBy,
			&category.UpdatedAt, &category.UpdatedBy,
			&user.ID, &user.Username, &user.Email,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		// Assign the user to the category
		category.User = user

		// Append category to the categories slice
		categories = append(categories, category)
	}

	// Check for any errors during row iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	return categories, nil
}

func ReadCategoriesByPostId(postId int) ([]Category, error) {
	db := db.OpenDBConnection()
	defer db.Close() // Ensure the connection is closed

	var categories []Category

	// Query to get all categories linked to a post
	query := `
		SELECT c.id AS category_id, c.name AS category_name, c.status AS category_status, 
		       c.created_at AS category_created_at, c.created_by AS category_created_by, 
		       c.updated_at AS category_updated_at, c.updated_by AS category_updated_by,
		       u.id AS user_id, u.username AS user_username, u.email AS user_email
		FROM post_categories pc
		INNER JOIN categories c ON pc.category_id = c.id AND c.status != 'delete'
		INNER JOIN users u ON c.created_by = u.id
		WHERE pc.post_id = ? AND pc.status != 'delete'
		ORDER BY c.name;
	`

	rows, err := db.Query(query, postId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through results and map to Category struct
	for rows.Next() {
		var category Category
		var user userManagementModels.User

		err := rows.Scan(
			&category.ID, &category.Name, &category.Status, &category.CreatedAt, &category.CreatedBy,
			&category.UpdatedAt, &category.UpdatedBy,
			&user.ID, &user.Username, &user.Email,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		category.User = user
		categories = append(categories, category)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	return categories, nil
}
