package db

import (
	"database/sql"
	"fmt"
	"net/http"
)

func InsertPost(w http.ResponseWriter, content string) {
	_, err := DB.Exec("INSERT INTO posts (author, authorID, title, content) VALUES (?, ?, ?, ?);", "testAuthor", 1, "testTitle", content)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
}

func GetPosts(w http.ResponseWriter) *sql.Rows {
	rows, err := DB.Query("SELECT id, content FROM posts ORDER BY id DESC")
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Database error", http.StatusInternalServerError)
		return nil
	}

	return rows
}
