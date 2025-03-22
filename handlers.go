package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/mail"
	"real-time-forum/db"
	"strings"
	"time"

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
	_, _, validSes := ValidateSession(r)
	if !validSes {
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

func SaveSession(userID, usname, sessionToken string, expiresAt time.Time) error {
	query := `INSERT INTO sessions (user_id, username, session_token, expires_at) VALUES (?, ?, ?, ?)`
	_, err := db.DB.Exec(query, userID, usname, sessionToken, expiresAt)
	return err
}

// sessionAndToken creates and puts a new session token into the database and into a user cookie
func sessionAndToken(w *http.ResponseWriter, userID, username string) {
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
	err = SaveSession(userID, username, sessionToken, expiresAt)
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

	if !NameOrEmailExists(creds.UsernameOrEmail) {
		fmt.Println("No username or email found, 1")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Username or E-mail not found",
		})
		return
	}

	// Get user information and check password
	var storedHashedPass, userID, username string
	query := `SELECT password, id, username FROM users WHERE username = ? OR email = ?`
	err := db.DB.QueryRow(query, creds.UsernameOrEmail, creds.UsernameOrEmail).Scan(&storedHashedPass, &userID, &username)
	if err != nil {
		fmt.Println("No username or email found, 2")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Username or E-mail not found",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHashedPass), []byte(creds.Password))
	if err != nil {
		fmt.Println("bad password")
		w.WriteHeader(http.StatusBadRequest) // Is it a bad request?
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Password incorrect",
		})
		return
	}

	// Remove any old sessions
	deleteSession(&w, userID)
	// Create new session and token
	sessionAndToken(&w, userID, username)

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	_, _, validSes := ValidateSession(r)
	if !validSes {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{ // to login screen?
			"success": false,
			"message": "Not logged in",
		})
		return
	}

	cookie, err := r.Cookie("session")
	if err == nil {
		db.DB.Exec(`DELETE FROM sessions WHERE session_token = ?`, cookie.Value)
	}

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour), // Expire immediately
		HttpOnly: true,
	})

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username  string `json:"username"`
		Age       string `json:"age"`
		Gender    string `json:"gender"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Email     string `json:"email"`
		Password  string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&creds)

	// TODO: restrictions for username and password?

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

	userUuid, idErr := uuid.NewV4() // Generate a new UUID user id
	if idErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Server error",
		})
		return
	}

	_, dbErr := db.DB.Exec("INSERT INTO users (uuid, username, age, gender, firstname, lastname, email, password) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", userUuid, creds.Username, creds.Age, creds.Gender, creds.FirstName, creds.LastName, creds.Email, hashPass)
	if dbErr != nil {
		// should return error at duplicate username or email
		fmt.Println(dbErr.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "User registration failed",
		})
		return
	}

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

	post.ID, post.Date = db.InsertPost(w, userName, usrId, post.Title, post.Content)
	db.InsertCategories(post.Categories, post.ID)

	day, time, _ := timeStrings(post.Date)
	post.Date = day + " " + time
	post.Author = userName
	//post.ReplyIds = db.GetReplyIds(w, post.ID)
	//post.RepliesCount = len(post.ReplyIds)
	post.RepliesCount = db.GetHowManyReplies(w, post.ID)

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

	rows := db.GetPosts(w)
	var posts []Post
	for rows.Next() {
		var post Post
		rows.Scan(&post.ID, &post.Title, &post.Author, &post.Date, &post.Content)
		posts = append(posts, post)
	}

	for i := range posts {
		posts[i].Categories = db.GetCategories(posts[i].ID)
		day, time, _ := timeStrings(posts[i].Date)
		posts[i].Date = day + " " + time
		posts[i].Likes, posts[i].Dislikes = countReactions(posts[i].ID)
		//posts[i].ReplyIds = db.GetReplyIds(w, posts[i].ID)
		//posts[i].RepliesCount = len(posts[i].ReplyIds)
		posts[i].RepliesCount = db.GetHowManyReplies(w, posts[i].ID)
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

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Try to delete the exact same row from the table (when already liked/disliked)
	res, _ := db.DB.Exec(`DELETE FROM post_reactions 
						  WHERE user_id = ? AND post_id = ? AND reaction_type = ?;`, userID, req.PostID, opinion)

	// Check if any row was deleted
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		fmt.Println("Affected rows checking failed:", err.Error())
	}

	// Add like/dislike: Update with current value on conflict
	if rowsAffected == 0 {
		_, err2 := db.DB.Exec(`INSERT INTO post_reactions (user_id, post_id, reaction_type) 
							   VALUES (?, ?, ?) 
							   ON CONFLICT (user_id, post_id) 
							   DO UPDATE SET reaction_type = excluded.reaction_type;`, userID, req.PostID, opinion)
		if err2 != nil {
			fmt.Println("Adding like or dislike:", err2.Error())
			http.Error(w, "Error adding reaction", http.StatusInternalServerError)
			return
		}
	}

	// Create post and send to all connections

	var post Post
	post.ID = req.PostID
	selectPostQuery := `SELECT author, title, content, created_at FROM posts WHERE id = ?;`
	err = db.DB.QueryRow(selectPostQuery, post.ID).Scan(&post.Author, &post.Title, &post.Content, &post.Date)
	if err != nil {
		http.Error(w, "Error finding post", http.StatusInternalServerError)
	}
	post.Likes, post.Dislikes = countReactions(req.PostID)
	post.Categories = db.GetCategories(post.ID)
	day, time, _ := timeStrings(post.Date)
	post.Date = day + " " + time
	//post.ReplyIds = db.GetReplyIds(w, post.ID)
	//post.RepliesCount = len(post.ReplyIds)
	post.RepliesCount = db.GetHowManyReplies(w, post.ID)

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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	usrId, userName, valid := ValidateSession(r)

	if valid {

		var post Post
		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		post.Content = html.EscapeString(strings.TrimSpace(post.Content))

		post.ID, post.Date = db.InsertReply(w, userName, usrId, post.Content, post.ParentId)
		db.InsertCategories(post.Categories, post.ID)

		day, time, _ := timeStrings(post.Date)
		post.Date = day + " " + time
		post.Author = userName
		//post.ReplyIds = db.GetReplyIds(w, post.ID)
		//post.RepliesCount = len(post.ReplyIds)
		post.RepliesCount = db.GetHowManyReplies(w, post.ID)

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

	// Get the parent_id (post ID) from the query parameter
	postID := r.URL.Query().Get("postID")
	if postID == "" {
		http.Error(w, "Missing post ID", http.StatusBadRequest)
		return
	}

	// Query the database for replies where parent_id matches postID
	rows, err := db.DB.Query("SELECT id, author, content, created_at, parent_id FROM posts WHERE parent_id = ?", postID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		fmt.Println("Database query error:", err)
		return
	}
	defer rows.Close()

	var replies []Post

	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Author, &p.Content, &p.Date, &p.ParentId); err != nil {
			http.Error(w, "Error scanning database", http.StatusInternalServerError)
			fmt.Println("Row scan error:", err)
			return
		}
		replies = append(replies, p)
	}

	for i := range replies {
		replies[i].Categories = db.GetCategories(replies[i].ID)
		day, time, _ := timeStrings(replies[i].Date)
		replies[i].Date = day + " " + time
		replies[i].Likes, replies[i].Dislikes = countReactions(replies[i].ID)
		replies[i].RepliesCount = db.GetHowManyReplies(w, replies[i].ID)
	}

	// Send response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"replies": replies,
	})
}
