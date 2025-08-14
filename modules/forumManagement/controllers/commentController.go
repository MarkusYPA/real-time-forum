package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"real-time-forum/config"
	"real-time-forum/modules/forumManagement/models"
	userManagementControllers "real-time-forum/modules/userManagement/controllers"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func ReplyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fmt.Println("Bad method at replying")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Method Not Allowed",
		})
		return
	}

	// Get the parent type from the query parameter
	parentType := r.URL.Query().Get("parentType")
	if parentType == "" {
		fmt.Println("No parent type at replying")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Missing parent type",
		})
		return
	}

	loginStatus, user, _, validateErr := userManagementControllers.ValidateSession(w, r)

	if !loginStatus || validateErr != nil {
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Not logged in",
		})
		return
	}

	if loginStatus {
		var msg config.Message

		var requestData struct {
			Content  string `json:"content"`
			ParentId int    `json:"parentid"`
		}

		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"success": false,
				"message": "Invalid request",
			})
			return
		}

		if requestData.Content == "" || requestData.ParentId == 0 {
			fmt.Println("Missing content or parent id in comment")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"success": false,
				"message": "Invalid request",
			})
			return
		}

		msg.MsgType = "comment"
		msg.Updated = false
		msg.IsReplied = true
		msg.Comment.Description = strings.TrimSpace(requestData.Content)

		parentPost, parentComment := requestData.ParentId, requestData.ParentId
		if parentType == "post" {
			parentComment = 0
			msg.Comment.PostId = requestData.ParentId
		} else if parentType == "comment" {
			parentPost = 0
			msg.Comment.CommentId = requestData.ParentId
		}
		var err error
		msg.Comment.ID, err = models.InsertComment(parentPost, parentComment, user.ID, msg.Comment.Description)

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
			msg.NumberOfReplis, err = models.CountCommentsForPost(msg.Comment.PostId)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{
					"success": false,
					"message": "Server error",
				})
				return
			}
		} else if parentType == "comment" {
			msg.NumberOfReplis, err = models.CountCommentsForComment(msg.Comment.CommentId)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{
					"success": false,
					"message": "Server error",
				})
				return
			}
		}
		msg.Comment.User = user
		msg.UserUUID = user.UUID
		msg.Comment.CreatedAt = time.Now()

		// Broadcast the new reply
		config.Broadcast <- msg

		// Also Broadcast parent to update number of replies?

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

func GetRepliesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Method not allowed",
		})
		return
	}
	loginStatus, user, _, validateErr := userManagementControllers.ValidateSession(w, r)

	if !loginStatus || validateErr != nil {
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
	if err != nil {
		fmt.Println("parentID atoi error:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"success": false})
		return
	}

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
	var comments []models.Comment

	if parentType == "post" {
		comments, err = models.ReadAllCommentsForPostByUserID(parentID, user.ID)
	} else if parentType == "comment" {
		comments, err = models.ReadAllCommentsForComment(parentID, user.ID)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Getting comments failed",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success":  true,
		"comments": comments,
	})
}
