package db

import (
	"database/sql"
	"fmt"
	"net/http"
)

func InsertPost(w http.ResponseWriter, name string, id string, title, content string) {
	_, err := DB.Exec("INSERT INTO posts (author, authorID, title, content) VALUES (?, ?, ?, ?);", name, id, title, content)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Database error", http.StatusInternalServerError)
	}
}

func GetPosts(w http.ResponseWriter) *sql.Rows {
	rows, err := DB.Query("SELECT id, title, content FROM posts ORDER BY id ASC")
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Database error", http.StatusInternalServerError)
		return nil
	}
	return rows
}
