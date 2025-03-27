package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/mail"
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
	loginStatus, _, sessionToken, _ := userControllers.ValidateSession(w, r)
	// if checkLoginError != nil {
	// 	errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
	// 	return
	// }

	if !loginStatus {
		json.NewEncoder(w).Encode(map[string]any{"loggedIn": false})
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"loggedIn": true, "token": sessionToken})
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
	sessionToken := userControllers.SessionGenerator(w, r, userID)
	json.NewEncoder(w).Encode(map[string]any{"success": true, "token": sessionToken})
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
	}

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// Handle WebSocket connections
func handleConnections(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.URL.Query().Get("session")
	if sessionToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Missing uuid ",
		})
		return
	}
	user, _, _ := userModels.SelectSession(sessionToken)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket error:", err)
		return
	}
	defer conn.Close()
	mu.Lock()
	clients[user.UUID] = conn
	mu.Unlock()
	//fmt.Println(clients)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}
	}
	mu.Lock()
	delete(clients, user.UUID)
	mu.Unlock()
}

// Broadcast new posts
func handleBroadcasts() {
	for {
		msg := <-broadcast
		//fmt.Println(clients)
		mu.Lock()
		specificClient, exists := clients[msg.UserUUID]
		mu.Unlock()

		if exists {
			err := specificClient.WriteJSON(msg)
			if err != nil {
				specificClient.Close()
				mu.Lock()
				delete(clients, msg.UserUUID)
				mu.Unlock()
			}
		}

		// Now broadcast to other clients
		mu.Lock()
		for uuid, client := range clients {
			if uuid == msg.UserUUID {
				continue
			}
			msg.Comment.IsLikedByUser = false
			msg.Comment.IsDislikedByUser = false
			msg.Post.IsDislikedByUser = false
			msg.Post.IsLikedByUser = false
			msg.IsLikAction = false

			mu.Unlock()
			err := client.WriteJSON(msg)
			mu.Lock()

			if err != nil {
				client.Close()
				delete(clients, uuid)
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
		Categories []int  `json:"categoryIds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		fmt.Println("json parse error:", err.Error())
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
	/* 	var categories []forumModels.Category
	   	for _, cat := range requestData.Categories {
	   		categories = append(categories, forumModels.Category{ID: cat})
	   	} */

	// Create a Post struct
	msg.MsgType = "post"
	msg.Updated = false
	msg.Post = forumModels.Post{
		Title:       title,
		Description: description,
		CreatedAt:   time.Now(),
		User:        user,
	}
	msg.UserUUID = user.UUID
	// Store post in DB
	var err error
	msg.Post.ID, err = forumModels.InsertPost(&msg.Post, requestData.Categories)
	if err != nil {
		fmt.Println("error inserting post:", err.Error())

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
		})
		return
	}

	msg.Post.Categories, err = forumModels.ReadCategoriesByPostId(msg.Post.ID)
	if err != nil {
		fmt.Println("error reading categories:", err.Error())

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
		fmt.Println("Wrong method on getting categories")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
		})
		return
	}
	categories, err := forumModels.ReadAllCategories()
	if err != nil {
		fmt.Println("Error reading categories:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
		})
		return
	}

	type dataToSend struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}

	var data []dataToSend
	for i := range categories {
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
		fmt.Println("faulty category id:", categoryIdString)

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Missing category",
		})
		return
	}

	catId, err := strconv.Atoi(categoryIdString)
	if err != nil {
		fmt.Println("faulty category id:", categoryIdString, catId, err.Error())

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Missing category",
		})
		return
	}

	var posts []forumModels.Post
	//posts, _ := db.ReadAllPosts()
	if catId == 0 {
		posts, err = forumModels.ReadAllPosts(user.ID)
	} else if catId > 0 {
		posts, err = forumModels.ReadPostsByCategoryId(user.ID, catId)
	}

	if err != nil {
		fmt.Println("error getting posts:", err.Error())

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Server error",
		})
		return
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
	var msg Message
	msg.IsLikAction = true
	if r.URL.Path != "/api/like" && r.URL.Path != "/api/dislike" {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Page does not exist",
		})
		return
	}
	if r.Method != http.MethodPost {
		fmt.Println("method:", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Method Not Allowed",
		})
		return
	}

	loginStatus, user, _, _ := userControllers.ValidateSession(w, r)
	if !loginStatus {
		fmt.Println("not a valid session")
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

	if postType == "post" {
		existingLikeId, existingLikeType := forumModels.PostHasLiked(user.ID, req.PostID)

		if existingLikeId == -1 {
			post := &forumModels.PostLike{
				Type:   opinion,
				PostId: req.PostID,
				UserId: user.ID,
			}
			_, insertError := forumModels.InsertPostLike(post)
			if insertError != nil {
				fmt.Println("Insert like error:", insertError.Error())
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{
					"success": false,
					"message": "Server error",
				})
				return
			}
		} else {
			updateError := forumModels.UpdateStatusPostLike(existingLikeId, "delete", user.ID)
			if updateError != nil {
				fmt.Println("Update like error:", updateError.Error())
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{
					"success": false,
					"message": "Server error",
				})
				return
			}

			if existingLikeType != opinion { //this is duplicated like or duplicated dislike so we should update it to disable
				post := &forumModels.PostLike{
					Type:   opinion,
					PostId: req.PostID,
					UserId: user.ID,
				}
				_, insertError := forumModels.InsertPostLike(post)
				if insertError != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]any{
						"success": false,
						"message": "Server error",
					})
					return
				}
			}
		}
	} else if postType == "comment" {
		existingLikeId, existingLikeType := forumModels.CommentHasLiked(user.ID, req.PostID)

		if existingLikeId == -1 {
			insertError := forumModels.InsertCommentLike(opinion, req.PostID, user.ID)
			if insertError != nil {
				fmt.Println(opinion, req.PostID, user.ID)
				fmt.Println("like comment:", insertError)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{
					"success": false,
					"message": "Server error",
				})
				return
			}
		} else {
			updateError := forumModels.UpdateCommentLikesStatus(existingLikeId, "delete", user.ID)
			if updateError != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{
					"success": false,
					"message": "Server error",
				})
				return
			}

			if existingLikeType != opinion { //this is duplicated like or duplicated dislike so we should update it to disable
				insertError := forumModels.InsertCommentLike(opinion, req.PostID, user.ID)

				if insertError != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]any{
						"success": false,
						"message": "Server error",
					})
					return
				}
			}
		}
	}

	// Create post and send to all connections

	msg.Updated = true
	msg.MsgType = postType
	msg.UserUUID = user.UUID
	if postType == "post" {
		msg.Post, err = forumModels.ReadPostById(req.PostID, user.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"success": false,
				"message": "Server error",
			})
			return
		}
	} else if postType == "comment" {
		msg.Comment, err = forumModels.ReadCommentById(req.PostID, user.ID)
		if err != nil {
			fmt.Println("read comment:", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"success": false,
				"message": "Server error",
			})
			return
		}
	}

	broadcast <- msg // Send to all WebSocket clients

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
	})
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
	parentType := r.URL.Query().Get("parentType")
	if parentType == "" {
		fmt.Println("No parent type at replying")
		http.Error(w, "Missing parent type", http.StatusBadRequest)
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
		msg.IsReplied = true
		msg.Comment.Description = html.EscapeString(strings.TrimSpace(requestData.Content))

		parentPost, parentComment := requestData.ParentId, requestData.ParentId
		if parentType == "post" {
			parentComment = 0
			msg.Comment.PostId = requestData.ParentId
		} else if parentType == "comment" {
			parentPost = 0
			msg.Comment.CommentId = requestData.ParentId
		}
		var err error
		msg.Comment.ID, err = forumModels.InsertComment(parentPost, parentComment, user.ID, msg.Comment.Description)

		if err != nil {
			fmt.Println("Error inserting comment", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"success": false,
				"message": "Server error",
			})
			return
		}

		if parentType == "post" {
			msg.NumberOfReplis, err = forumModels.CountCommentsForPost(msg.Comment.PostId)
			if err != nil {
				fmt.Println(err)
				return
			}
		} else if parentType == "comment" {
			msg.NumberOfReplis, err = forumModels.CountCommentsForComment(msg.Comment.CommentId)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		msg.Comment.User = user
		msg.UserUUID = user.UUID
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
	loginStatus, user, _, _ := userControllers.ValidateSession(w, r)
	if !loginStatus {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Not logged in",
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
		comments, err = forumModels.ReadAllCommentsForPostByUserID(parentID, user.ID)
	} else if parentType == "comment" {
		fmt.Println(parentID)
		comments, err = forumModels.ReadAllCommentsForComment(parentID, user.ID)
		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println(comments)
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
