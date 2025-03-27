package models

import (
	"database/sql"
	"fmt"
	"time"
)

// Post struct represents the user data model
type PostCategory struct {
	ID         int       `json:"id"`
	PostId     int       `json:"post_id"`
	CategoryId int       `json:"category_id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	CreatedBy  int       `json:"created_by"`
	UpdatedAt  time.Time `json:"updated_at"`
	UpdatedBy  int       `json:"updated_by"`
}

func InsertPostCategories(post_id int, categories []int, user_id int, tx *sql.Tx) error {
	// Prepare the bulk insert query for post_categories
	if len(categories) > 0 {
		query := `INSERT INTO post_categories (post_id, category_id, created_by) VALUES `
		values := make([]any, 0, len(categories)*3)

		for i, categoryID := range categories {
			if i > 0 {
				query += ", "
			}
			query += "(?, ?, ?)"
			values = append(values, post_id, categoryID, user_id)
		}
		query += ";"

		// Execute the bulk insert query
		_, err := tx.Exec(query, values...)
		if err != nil {
			fmt.Println("here is error")
			fmt.Println(err)
			tx.Rollback() // Rollback on error
			return err
		}
	}
	return nil
}

func UpdateStatusPostCategories(post_id int, user_id int, status string, tx *sql.Tx) error {
	updateStatusQuery := `UPDATE post_categories
					SET status = ?,
						updated_at = CURRENT_TIMESTAMP,
						updated_by = ?
					WHERE post_id = ?
					AND status != 'delete';`
	_, updateStatusErr := tx.Exec(updateStatusQuery, status, user_id, post_id)
	if updateStatusErr != nil {
		tx.Rollback() // Rollback on error
		// Check if the error is a SQLite constraint violation
		if sqliteErr, ok := updateStatusErr.(interface{ ErrorCode() int }); ok {
			if sqliteErr.ErrorCode() == 19 { // SQLite constraint violation error code
				return sql.ErrNoRows // Return custom error to indicate a duplicate
			}
		}
		return updateStatusErr
	}

	return nil
}

// func DeletePostCategories(post_id int, user_id int, tx *sql.Tx) error {
// 	deleteQuery := `UPDATE post_categories
// 					SET status = 'delete',
// 						updated_at = CURRENT_TIMESTAMP,
// 						updated_by = ?
// 					WHERE post_id = ?
// 					AND status != 'delete';`
// 	_, deleteErr := tx.Exec(deleteQuery, user_id, post_id)
// 	if deleteErr != nil {
// 		tx.Rollback() // Rollback on error
// 		// Check if the error is a SQLite constraint violation
// 		if sqliteErr, ok := deleteErr.(interface{ ErrorCode() int }); ok {
// 			if sqliteErr.ErrorCode() == 19 { // SQLite constraint violation error code
// 				return sql.ErrNoRows // Return custom error to indicate a duplicate
// 			}
// 		}
// 		return deleteErr
// 	}
// 	return nil
// }
