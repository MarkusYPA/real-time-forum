package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/mail"
	"real-time-forum/db"
	forumModels "real-time-forum/modules/forumManagement/models"
	userControllers "real-time-forum/modules/userManagement/controllers"
	userModels "real-time-forum/modules/userManagement/models"
	"strconv"
	"strings"
	"time"

	// "github.com/gofrs/uuid"

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

// sessionAndToken creates and puts a new session token into the database and into a user cookie
/* func sessionAndToken(w *http.ResponseWriter, userID string) {
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
} */

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
	err := userModels.DeleteSession(sessionToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Server Error",
		})
		return
	}

	userControllers.DeleteCookie(w, "session_token")

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	// allow registering new user while logged in

	var creds userModels.User
	json.NewDecoder(r.Body).Decode(&creds)

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
	_, insertError := userModels.InsertUser(&creds)
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
		msg := <-broadcast
		mu.Lock()
		for client := range clients {
			err := client.WriteJSON(msg)
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
	loginStatus, user, _, _ := userControllers.ValidateSession(w, r)
	if !loginStatus {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Not logged in",
		})
		return
	}

	var msg Message
	var requestData struct {
		Title      string `json:"title"`
		Content    string `json:"content"`
		Categories []int  `json:"categories"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Invalid request",
		})
		return
	}

	// Sanitize input
	title := html.EscapeString(strings.TrimSpace(requestData.Title))
	description := html.EscapeString(strings.TrimSpace(requestData.Content))

	// Convert category strings to Category structs
	var categories []forumModels.Category
	for _, cat := range requestData.Categories {
		categories = append(categories, forumModels.Category{ID: cat})
	}

	// Create a Post struct
	msg.MsgType = "post"
	msg.Updated = false
	msg.Post = forumModels.Post{
		Title:       title,
		Description: description,
		//Categories:  categories, // Only ids?
		CreatedAt: time.Now(),
		User:      user,
	}

	// Store post in DB
	var err error
	msg.Post.ID, err = forumModels.InsertPost(&msg.Post, requestData.Categories)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
		})
		return
	}

	msg.Post.Categories, err = forumModels.ReadCategoriesByPostId(msg.Post.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
		})
		return
	}

	// Broadcast the post
	broadcast <- msg

	// Send response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
	})

}
func categoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		//do something
		return
	}
	categories, err := forumModels.ReadAllCategories()
	if err != nil {
		// do something
		return
	}

	type dataToSend struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}
	var data []dataToSend
	for i := 0; i < len(categories); i++ {
		data = append(data, dataToSend{Id: categories[i].ID, Name: categories[i].Name})
	}

	json.NewEncoder(w).Encode(map[string]any{
		"success":    true,
		"categories": data,
	})

}

// Get all posts
func handleGetPosts(w http.ResponseWriter, r *http.Request) {
	loginStatus, user, _, _ := userControllers.ValidateSession(w, r)
	if !loginStatus {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Not logged in",
		})
		return
	}

	// Get category from query
	categoryIdString := r.URL.Query().Get("categoryid")
	if categoryIdString == "" {
		http.Error(w, "Missing category", http.StatusBadRequest)
		return
	}
	catId, err := strconv.Atoi(categoryIdString)
	if err != nil {
		//do something
		return
	}

	var posts []forumModels.Post
	//posts, _ := db.ReadAllPosts()
	if catId == 0 {
		posts, err = forumModels.ReadAllPosts(user.ID)
	} else if catId > 0 {
		posts, err = forumModels.ReadPostsByCategoryId(user.ID, catId)
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

	loginStatus, user, _, _ := userControllers.ValidateSession(w, r)
	if !loginStatus {
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
		//http.Error(w, "Missing post type", http.StatusBadRequest)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Missing post type",
		})
		return
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	/* 	// Try to delete the exact same row from the table (when already liked/disliked)
	   	res, _ := db.DB.Exec(`DELETE FROM `+postType+`_likes
	   						  WHERE user_id = ? AND `+postType+`_id = ? AND type = ?;`, user.ID, req.PostID, opinion)

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
	   	} */

	err = forumModels.UpdateStatusPostLike(req.PostID, "enable", user.ID)
	if err != nil {
		fmt.Println("DB error:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Error updating post",
		})
		return
	}

	// Create post and send to all connections

	var msg Message

	//selectPostQuery := `SELECT author, title, description, created_at FROM ` + postType + `s WHERE id = ?;`
	selectPostQuery := ""

	if postType == "post" {
		msg.Post.ID = req.PostID
		selectPostQuery = `
			SELECT u.username, p.title, p.description, p.created_at
			FROM posts p
			JOIN users u ON p.user_id = u.id
			WHERE p.id = ? AND p.status = 'enable'
		`

		err = db.DB.QueryRow(selectPostQuery, msg.Post.ID).Scan(&msg.Post.User, &msg.Post.Title, &msg.Post.Description, &msg.Post.CreatedAt)
	}

	if postType == "comment" {
		msg.Comment.ID = req.PostID
		selectPostQuery = `
			SELECT u.username, c.description, c.created_at
			FROM comments c
			JOIN users u ON c.user_id = u.id
			WHERE c.id = ? AND c.status = 'enable'
		`

		err = db.DB.QueryRow(selectPostQuery, msg.Comment.ID).Scan(&msg.Comment.User, &msg.Comment.Description, &msg.Comment.CreatedAt)
	}

	if err != nil {
		fmt.Println("DB error:", err.Error())
		//http.Error(w, "Error finding post", http.StatusInternalServerError)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Error finding post",
		})
		return
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

	broadcast <- msg // Send to all WebSocket clients

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

	// Get the parent type from the query parameter
	parentType := r.URL.Query().Get("postType")
	if parentType == "" {
		fmt.Println("No post type at replying")
		http.Error(w, "Missing post type", http.StatusBadRequest)
		return
	}

	loginStatus, user, _, _ := userControllers.ValidateSession(w, r)
	if !loginStatus {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Not logged in",
		})
		return
	}

	if loginStatus {
		fmt.Println("reply is valid")

		var msg Message

		var requestData struct {
			Content  string `json:"content"`
			ParentId int    `json:"parentid"`
		}

		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		msg.MsgType = "comment"
		msg.Updated = false
		msg.Comment.Description = html.EscapeString(strings.TrimSpace(requestData.Content))

		parentPost, parentComment := requestData.ParentId, requestData.ParentId
		if parentType == "post" {
			parentComment = 0
		} else if parentType == "comment" {
			parentPost = 0
		}
		var err error
		msg.Comment.ID, err = forumModels.InsertComment(parentPost, parentComment, user.ID, msg.Comment.Description)
		if err != nil {
			// do something here
			return
		}
		msg.Comment.User = user
		msg.Comment.CreatedAt = time.Now()

		// Broadcast the new reply
		broadcast <- msg

		// Also broadcast parent to update number of replies?

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
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
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Method not allowed",
		})
		return
	}

	// Get the parent_id (parentID) from the query parameter
	parentIDString := r.URL.Query().Get("parentID")
	if parentIDString == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Missing parent ID",
		})
		return
	}

	parentID, err := strconv.Atoi(parentIDString)

	// Get the parent type from the query parameter
	parentType := r.URL.Query().Get("parentType")
	if parentType == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Missing parent type",
		})
		return
	}

	// Query the database for replies where post_id or comment_id matches parentID
	var comments []forumModels.Comment
	if parentType == "post" {
		comments, err = forumModels.ReadAllCommentsForPost(parentID)
	} else if parentType == "comment" {
		comments, err = forumModels.ReadAllCommentsForComment(parentID)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Getting comments failed",
		})
		return
	}

	/* 	rows, err := db.DB.Query(`
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
		} */

	// Send response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success":  true,
		"comments": comments,
	})
}
