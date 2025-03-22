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

func InsertReply(w http.ResponseWriter, name string, authorId string, content string, parentId int) (int, string) {
	result, err := DB.Exec("INSERT INTO posts (author, authorID, content, parent_id ) VALUES (?, ?, ?, ?);", name, authorId, content, parentId)
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
	rows, err := DB.Query("SELECT id, title, author, created_at, content FROM posts WHERE parent_id = 0 ORDER BY id ASC")
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

func GetReactions(postId int) (int, int) {

	reactionsQuery := `SELECT reaction_type, COUNT(*) AS count FROM post_reactions WHERE post_id = ? GROUP BY reaction_type;`
	rows, err := DB.Query(reactionsQuery, postId)
	if err != nil {
		fmt.Println("Fetching reactions query failed", err.Error())
		return 0, 0
	}
	defer rows.Close()

	var likes, dislikes int
	for rows.Next() {
		var reactionType string
		var count int
		if err := rows.Scan(&reactionType, &count); err != nil {
			fmt.Printf("Failed to scan row: %v\n", err)
		}

		// Assign counts based on reaction type
		switch reactionType {
		case "like":
			likes = count
		case "dislike":
			dislikes = count
		}
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		fmt.Printf("Error iterating rows: %v\n", err)
	}

	return likes, dislikes
}

func GetReplyIds(w http.ResponseWriter, postId int) []int {
	replyIds := []int{}

	// get post IDs with the given parent_id
	rows, err := DB.Query("SELECT id FROM posts WHERE parent_id = ?", postId)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return replyIds
	}
	defer rows.Close()

	// reply IDs into a slice
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			http.Error(w, "Error scanning database", http.StatusInternalServerError)
			return replyIds
		}
		replyIds = append(replyIds, id)
	}

	return replyIds
}

func GetHowManyReplies(w http.ResponseWriter, postId int) int {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM posts WHERE parent_id = ?", postId).Scan(&count)
	if err != nil {
		http.Error(w, "Error scanning database", http.StatusInternalServerError)
		return count
	}
	return count
}
