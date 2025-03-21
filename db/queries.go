package db

import (
	"database/sql"
	"fmt"
	"net/http"
)

func InsertPost(w http.ResponseWriter, name string, authorId string, title, content string) (int, string) {
	result, err := DB.Exec("INSERT INTO posts (author, authorID, title, content) VALUES (?, ?, ?, ?);", name, authorId, title, content)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Database error", http.StatusInternalServerError)
	}

	postID, err := result.LastInsertId()
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Failed to retrieve post ID", http.StatusInternalServerError)
	}

	date := ""
	selectQueryThread := `SELECT created_at FROM posts WHERE id = ?;`
	err = DB.QueryRow(selectQueryThread, postID).Scan(&date)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Failed to retrieve post ID", http.StatusInternalServerError)
	}

	return int(postID), date
}

func GetPosts(w http.ResponseWriter) *sql.Rows {
	rows, err := DB.Query("SELECT id, title, author, created_at, content FROM posts ORDER BY id ASC")
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Database error", http.StatusInternalServerError)
		return nil
	}
	return rows
}

func InsertCategories(categories []string, threadID int) {
	for _, category := range categories {

		_, err := DB.Exec(`INSERT OR IGNORE INTO categories (name) VALUES (?);`, category)
		if err != nil {
			fmt.Println("Adding category:", err.Error())
			// handle error
			return
		}

		_, err = DB.Exec(`INSERT OR IGNORE INTO posts_categories (post_id, category_id) 
							 VALUES (?, (SELECT id FROM categories WHERE name=?));`, threadID, category)
		if err != nil {
			fmt.Println("Adding posts-category:", err.Error())
			// handle error
			return
		}
	}
}
