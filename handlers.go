package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/mail"
	"real-time-forum/db"
	userControllers "real-time-forum/modules/userManagement/controllers"
	userModels "real-time-forum/modules/userManagement/models"
	"strings"
	"time"

	// "github.com/gofrs/uuid"
	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 Not Found", http.StatusNotFound) // error 404
		return
	}
	if r.Method == http.MethodGet {
		err := homeTmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
	}
}

// handleSessionCheck is used to check if the user hasa valid session at first loading the page
func handleSessionCheck(w http.ResponseWriter, r *http.Request) {
	loginStatus, _, _, _ := userControllers.ValidateSession(w, r)
	// if checkLoginError != nil {
	// 	errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
	// 	return
	// }

	if !loginStatus {
		json.NewEncoder(w).Encode(map[string]bool{"loggedIn": false})
		return
	}
	json.NewEncoder(w).Encode(map[string]bool{"loggedIn": true})
}

func NameOrEmailExists(input string) bool {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = ? OR email = ?)`
	err := db.DB.QueryRow(query, input, input).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

// deleteSession removes all sessions from the db by user Id
func deleteSession(w *http.ResponseWriter, usrId string) {
	query := `DELETE FROM sessions WHERE user_id = ?`
	_, err := db.DB.Exec(query, usrId)
	if err != nil {
		(*w).WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(*w).Encode(map[string]any{
			"success": false,
			"message": "Server error",
		})
	}
}

func CreateSession() (string, error) {
	sessionUUID, err := uuid.NewV4() // Generate a new UUID
	if err != nil {
		return "", err
	}
	return sessionUUID.String(), nil
}

func SaveSession(userID, sessionToken string, expiresAt time.Time) error {
	query := `INSERT INTO sessions (user_id, session_token, expires_at) VALUES (?, ?, ?)`
	_, err := db.DB.Exec(query, userID, sessionToken, expiresAt)
	return err
}

// sessionAndToken creates and puts a new session token into the database and into a user cookie
func sessionAndToken(w *http.ResponseWriter, userID string) {
	// New session token
	sessionToken, err := CreateSession()
	if err != nil {
		(*w).WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(*w).Encode(map[string]any{
			"success": false,
			"message": "Server error",
		})
		return
	}
	expiresAt := time.Now().Add(24 * time.Hour)

	// Token into database
	err = SaveSession(userID, sessionToken, expiresAt)
	if err != nil {
		(*w).WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(*w).Encode(map[string]any{
			"success": false,
			"message": "Server error",
		})
		return
	}

	// Token into cookie
	http.SetCookie(*w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  expiresAt,
		HttpOnly: true,
	})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		UsernameOrEmail string `json:"usernameOrEmail"`
		Password        string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&creds)
	hasFound, userID, err := userModels.AuthenticateUser(creds.UsernameOrEmail, creds.Password)
	if !hasFound {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	userControllers.SessionGenerator(w, r, userID)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	loginStatus, _, sessionToken, _ := userControllers.ValidateSession(w, r)
	if !loginStatus {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Not logged in",
		})
		return
	}
	userModels.DeleteSession(sessionToken)
	// if err != nil {
	// 	errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
	// 	return
	// }

	userControllers.DeleteCookie(w, "session_token")

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	// loginStatus, _, sessionToken, _ := userControllers.ValidateSession(w, r)
	// if !loginStatus {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	json.NewEncoder(w).Encode(map[string]any{
	// 		"success": false,
	// 		"message": "Not logged in",
	// 	})
	// 	return
	// }
	var creds userModels.User
	json.NewDecoder(r.Body).Decode(&creds)

	// if len(creds.Username) == 0 || len(creds.Email) == 0 || len(creds.Password) == 0 {
	// 	return
	// }
	_, emailErr := mail.ParseAddress(creds.Email)
	if emailErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Invalid e-mail",
		})
		return
	}
	hashPass, cryptErr := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if cryptErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Server error",
		})
		return
	}
	creds.Password = string(hashPass)
	// Insert a record while checking duplicates
	userId, insertError := userModels.InsertUser(&creds)
	if insertError != nil {
		fmt.Println(insertError.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "User registration failed",
		})
		return
	} else {
		fmt.Println("User added successfully!")
	}

	userControllers.SessionGenerator(w, r, userId)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// Handle WebSocket connections
func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket error:", err)
		return
	}
	defer conn.Close()

	mu.Lock()
	clients[conn] = true
	mu.Unlock()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			mu.Lock()
			delete(clients, conn)
			mu.Unlock()
			break
		}
	}
}

// Broadcast new posts
func handleBroadcasts() {
	for {
		post := <-broadcast
		mu.Lock()
		for client := range clients {
			err := client.WriteJSON(post)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
	}
}

// Handle new post submissions
func handleNewPost(w http.ResponseWriter, r *http.Request) {
	usrId, userName, validSes := ValidateSession(r)
	if !validSes {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Not logged in",
		})
		return
	}

	var post Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil { // title, content and categories from request
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	post.Title = html.EscapeString(strings.TrimSpace(post.Title))
	post.Content = html.EscapeString(strings.TrimSpace(post.Content))
	for i, cat := range post.Categories {
		post.Categories[i] = html.EscapeString(strings.TrimSpace(cat))
	}

	post.ID, post.Date = db.InsertPost(w, usrId, post.Title, post.Content)
	db.InsertCategories(post.Categories, post.ID, usrId)

	day, time, _ := timeStrings(post.Date)
	post.Title = html.UnescapeString(post.Title)
	post.Content = html.UnescapeString(post.Content)
	post.Date = day + " " + time
	post.Author = userName
	post.RepliesCount = db.GetHowManyCommentsPost(w, post.ID)
	post.PostType = "post"

	// Broadcast the new post
	broadcast <- post

	w.WriteHeader(http.StatusCreated)

	//json.NewEncoder(w).Encode(post)
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"post":    post,
	})
}

// Get all posts
func handleGetPosts(w http.ResponseWriter, r *http.Request) {

	_, _, validSes := ValidateSession(r)
	if !validSes {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Not logged in",
		})
		return
	}

	// Get category from query
	category := r.URL.Query().Get("category")
	if category == "" {
		http.Error(w, "Missing category", http.StatusBadRequest)
		return
	}

	//posts, _ := db.ReadAllPosts()

	rows := db.GetPosts(w, category)
	var posts []Post
	for rows.Next() {
		var post Post
		rows.Scan(&post.ID, &post.Title, &post.Author, &post.Date, &post.Content)
		posts = append(posts, post)
	}

	for i := range posts {
		posts[i].Title = html.UnescapeString(posts[i].Title)
		posts[i].Content = html.UnescapeString(posts[i].Content)
		posts[i].Categories = db.GetCategories(posts[i].ID)
		day, time, _ := timeStrings(posts[i].Date)
		posts[i].Date = day + " " + time
		//posts[i].Likes, posts[i].Dislikes = countReactions(posts[i].ID)
		posts[i].Likes, posts[i].Dislikes = db.GetPostLikes(posts[i].ID)
		posts[i].RepliesCount = db.GetHowManyCommentsPost(w, posts[i].ID)
	}

	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"posts":   posts,
	})
}

func handlePosts(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		handleNewPost(w, r) // Call function to process new post
		return
	}
	if r.Method == http.MethodGet {
		handleGetPosts(w, r) // Call function to fetch all posts
		return
	}
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func likeOrDislike(w http.ResponseWriter, r *http.Request, opinion string) {
	if r.URL.Path != "/api/like" && r.URL.Path != "/api/dislike" {
		http.Error(w, "Page does not exist", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _, valid := ValidateSession(r)
	if !valid {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Not logged in",
		})
		return
	}

	var req struct {
		PostID int `json:"postID"`
	}

	// Get the post type from the query parameter
	postType := r.URL.Query().Get("postType")
	if postType == "" {
		http.Error(w, "Missing post type", http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Try to delete the exact same row from the table (when already liked/disliked)
	res, _ := db.DB.Exec(`DELETE FROM `+postType+`_likes 
						  WHERE user_id = ? AND `+postType+`_id = ? AND type = ?;`, userID, req.PostID, opinion)

	// Check if any row was deleted
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		fmt.Println("Affected rows checking failed:", err.Error())
	}

	// Add like/dislike: Update with current value on conflict
	if rowsAffected == 0 {

		query := `INSERT INTO ` + postType + `_likes (user_id, ` + postType + `_id, type) 
							   VALUES (?, ?, ?) 
							   ON CONFLICT (user_id, ` + postType + `_id) 
							   DO UPDATE SET type = excluded.type;`
		_, err2 := db.DB.Exec(query, userID, req.PostID, opinion)
		if err2 != nil {
			fmt.Println("Adding like or dislike:", err2.Error())
			http.Error(w, "Error adding reaction", http.StatusInternalServerError)
			return
		}
	}

	// Create post and send to all connections

	var post Post
	post.ID = req.PostID
	//selectPostQuery := `SELECT author, title, description, created_at FROM ` + postType + `s WHERE id = ?;`
	selectPostQuery := ""

	if postType == "post" {
		selectPostQuery = `
			SELECT u.username, p.title, p.description, p.created_at
			FROM posts p
			JOIN users u ON p.user_id = u.id
			WHERE p.id = ? AND p.status = 'enable'
		`

		err = db.DB.QueryRow(selectPostQuery, post.ID).Scan(&post.Author, &post.Title, &post.Content, &post.Date)
	}

	if postType == "comment" {
		selectPostQuery = `
			SELECT u.username, c.description, c.created_at
			FROM comments c
			JOIN users u ON c.user_id = u.id
			WHERE c.id = ? AND c.status = 'enable'
		`

		err = db.DB.QueryRow(selectPostQuery, post.ID).Scan(&post.Author, &post.Content, &post.Date)
	}

	if err != nil {
		fmt.Println("Some error:", err.Error())
		http.Error(w, "Error finding post", http.StatusInternalServerError)

	}
	post.Title = html.UnescapeString(post.Title) // Ok if title is empty?
	post.Content = html.UnescapeString(post.Content)
	post.Likes, post.Dislikes = db.GetPostLikes(req.PostID)
	if postType == "post" {
		post.Categories = db.GetCategories(post.ID)
		post.RepliesCount = db.GetHowManyCommentsPost(w, post.ID)
	}
	if postType == "comment" {
		post.RepliesCount = db.GetHowManyCommentsComment(w, post.ID)
	}
	day, time, _ := timeStrings(post.Date)
	post.Date = day + " " + time
	post.PostType = postType

	broadcast <- post // Send to all WebSocket clients

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Post liked"})
}

func likeHandler(w http.ResponseWriter, r *http.Request) {
	likeOrDislike(w, r, "like")
}

func dislikeHandler(w http.ResponseWriter, r *http.Request) {
	likeOrDislike(w, r, "dislike")
}

func replyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fmt.Println("Bad method at replying")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the post type from the query parameter
	postType := r.URL.Query().Get("postType")
	if postType == "" {
		fmt.Println("No post type at replying")
		http.Error(w, "Missing post type", http.StatusBadRequest)
		return
	}

	usrId, userName, valid := ValidateSession(r)

	if valid {
		fmt.Println("reply is valid")

		var post Post
		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		post.Content = html.EscapeString(strings.TrimSpace(post.Content))
		post.ID, post.Date = db.InsertReply(w, usrId, post.Content, post.ParentId, postType)
		day, time, _ := timeStrings(post.Date)
		post.Date = day + " " + time
		post.Author = userName
		post.RepliesCount = db.GetHowManyCommentsComment(w, post.ID)
		post.PostType = "comment"

		fmt.Println(post)

		// Broadcast the new reply
		broadcast <- post

		// Also broadcast parent to update number of replies?

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"post":    post,
		})
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"message": "Not logged in",
	})
}

func getRepliesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the parent_id (parentID) from the query parameter
	parentID := r.URL.Query().Get("parentID")
	if parentID == "" {

		fmt.Println("Missing post id")
		http.Error(w, "Missing post ID", http.StatusBadRequest)
		return
	}

	// Get the post type from the query parameter
	postType := r.URL.Query().Get("postType")
	if postType == "" {
		http.Error(w, "Missing post type", http.StatusBadRequest)
		return
	}

	// Query the database for replies where parent_id matches parentID
	//rows, err := db.DB.Query("SELECT id, author, content, created_at, parent_id FROM comments WHERE parent_id = ?", parentID)
	rows, err := db.DB.Query(`
    	SELECT c.id, u.username, c.description, c.created_at, c.`+postType+`_id
    	FROM comments c
    	JOIN users u ON c.user_id = u.id
    	WHERE c.`+postType+`_id = ?`, parentID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Println("Database query error:", err)
		return
	}
	defer rows.Close()

	var replies []Post

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Author, &post.Content, &post.Date, &post.ParentId); err != nil {
			http.Error(w, "Error scanning database", http.StatusInternalServerError)
			log.Println("Row scan error:", err)
			return
		}
		replies = append(replies, post)
	}

	for i := range replies {
		replies[i].Content = html.UnescapeString(replies[i].Content)
		//replies[i].Categories = db.GetCategories(replies[i].ID) // No categories on replies
		day, time, _ := timeStrings(replies[i].Date)
		replies[i].Date = day + " " + time
		replies[i].Likes, replies[i].Dislikes = db.GetCommentLikes(replies[i].ID)
		replies[i].RepliesCount = db.GetHowManyCommentsComment(w, replies[i].ID)
	}

	// Send response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"replies": replies,
	})
}
