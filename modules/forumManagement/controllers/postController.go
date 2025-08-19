package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"real-time-forum/config"
	forumModels "real-time-forum/modules/forumManagement/models"
	userManagementControllers "real-time-forum/modules/userManagement/controllers"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func HandlePosts(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		HandleNewPost(w, r) // Call function to process new post
		return
	}
	if r.Method == http.MethodGet {
		HandleGetPosts(w, r) // Call function to fetch all posts
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(map[string]any{
		"success": false,
	})
}

// Get all posts
func HandleGetPosts(w http.ResponseWriter, r *http.Request) {
	loginStatus, user, _, validateErr := userManagementControllers.ValidateSession(w, r)

	if !loginStatus || validateErr != nil {
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

// Handle new post submissions
func HandleNewPost(w http.ResponseWriter, r *http.Request) {
	loginStatus, user, _, validateErr := userManagementControllers.ValidateSession(w, r)

	if !loginStatus || validateErr != nil {
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Not logged in",
		})
		return
	}

	var msg config.Message
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

	if requestData.Title == "" || requestData.Content == "" || len(requestData.Categories) == 0 {
		fmt.Println("Missing title, content or categories in post")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Invalid request",
		})
		return
	}

	// Trim input
	title := strings.TrimSpace(requestData.Title)
	description := strings.TrimSpace(requestData.Content) // Displaying as texcontent prevents execution

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
	config.Broadcast <- msg

	// Send response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
	})

}

func CategoryHandler(w http.ResponseWriter, r *http.Request) {
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
