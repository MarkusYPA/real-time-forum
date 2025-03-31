package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"real-time-forum/config"
	"real-time-forum/modules/forumManagement/models"
	userManagementControllers "real-time-forum/modules/userManagement/controllers"
)

func LikeOrDislike(w http.ResponseWriter, r *http.Request, opinion string) {
	var msg config.Message
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

	loginStatus, user, _, validateErr := userManagementControllers.ValidateSession(w, r)

	if !loginStatus || validateErr != nil {
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
		existingLikeId, existingLikeType := models.PostHasLike(user.ID, req.PostID)

		if existingLikeId == -1 {
			post := &models.PostLike{
				Type:   opinion,
				PostId: req.PostID,
				UserId: user.ID,
			}
			_, insertError := models.InsertPostLike(post)
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
			updateError := models.UpdateStatusPostLike(existingLikeId, "delete", user.ID)
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
				post := &models.PostLike{
					Type:   opinion,
					PostId: req.PostID,
					UserId: user.ID,
				}
				_, insertError := models.InsertPostLike(post)
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
		existingLikeId, existingLikeType := models.CommentHasLiked(user.ID, req.PostID)

		if existingLikeId == -1 {
			insertError := models.InsertCommentLike(opinion, req.PostID, user.ID)
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
			updateError := models.UpdateCommentLikesStatus(existingLikeId, "delete", user.ID)
			if updateError != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{
					"success": false,
					"message": "Server error",
				})
				return
			}

			if existingLikeType != opinion { //this is duplicated like or duplicated dislike so we should update it to disable
				insertError := models.InsertCommentLike(opinion, req.PostID, user.ID)

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
		msg.Post, err = models.ReadPostById(req.PostID, user.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"success": false,
				"message": "Server error",
			})
			return
		}
	} else if postType == "comment" {
		msg.Comment, err = models.ReadCommentById(req.PostID, user.ID)
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

	config.Broadcast <- msg // Send to all WebSocket Clients

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
	})
}

func LikeHandler(w http.ResponseWriter, r *http.Request) {
	LikeOrDislike(w, r, "like")
}

func DislikeHandler(w http.ResponseWriter, r *http.Request) {
	LikeOrDislike(w, r, "dislike")
}
