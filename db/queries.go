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

func GetCategories(postId int) []string {
	selectQuery := `SELECT categories.name AS category FROM categories JOIN posts_categories ON posts_categories.category_id = categories.id WHERE post_id = ?;`

	foundCategories := []string{}

	rowsCategories, err := DB.Query(selectQuery, postId)
	if err != nil {
		fmt.Println("fetchCategories selectQuery failed", err.Error())
		return foundCategories
	}
	defer rowsCategories.Close()

	var category string
	for rowsCategories.Next() {
		err = rowsCategories.Scan(&category)
		if err != nil {
			fmt.Println("Error reading category:", err.Error())
			return foundCategories
		}
		foundCategories = append(foundCategories, category)
	}
	if err := rowsCategories.Err(); err != nil {
		fmt.Println("Error iterating through rows:", err.Error())
	}
	return foundCategories
}
