package models

import (
	"database/sql"
	"fmt"
)

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
