package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"real-time-forum/config"
	"real-time-forum/modules/forumManagement/models"
	userManagementControllers "real-time-forum/modules/userManagement/controllers"
	userModels "real-time-forum/modules/userManagement/models"
	"time"
)

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	loginStatus, user, _, validateErr := userManagementControllers.ValidateSession(w, r)

	if !loginStatus || validateErr != nil {
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Not logged in",
		})
		return
	}
	var err error
	var msg config.Message
	msg.ChattedUsers, msg.UnchattedUsers, err = models.ReadAllUsers(user.ID)
	if err != nil {
		fmt.Println("Error getting list of users:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "config Error",
		})
		return
	}

	for i, usr := range msg.ChattedUsers {
		if _, ok := config.Clients[usr.UserUUID]; ok {
			msg.ChattedUsers[i].IsOnline = true
		}
	}

	for i, usr := range msg.UnchattedUsers {
		if _, ok := config.Clients[usr.UserUUID]; ok {
			msg.UnchattedUsers[i].IsOnline = true
		}
	}

	msg.MsgType = "listOfChat"
	msg.Updated = false
	msg.UserUUID = user.UUID

	config.Broadcast <- msg

	// Send response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
	})
}

func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	loginStatus, sendUser, _, validateErr := userManagementControllers.ValidateSession(w, r)

	var msg config.Message

	if !loginStatus || validateErr != nil {
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Not logged in",
		})
		return
	}

	reciverUserUUID := r.URL.Query().Get("UserUUID")
	_, exists := config.Clients[reciverUserUUID]
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"message": "user is offline try later",
		})
		return
	}

	chatUUID := r.URL.Query().Get("ChatUUID")

	if chatUUID == "" {
		reciverID, err := userModels.FindUserByUUID(reciverUserUUID)
		if err != nil {
			fmt.Println("find user : ", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"success": false,
			})
			return
		}

		// Double check for chat with both user IDs (If two users open chat before any message is sent)
		chatUUID, err = models.FindChatUUIDbyUserIDS(sendUser.ID, reciverID)

		if chatUUID == "" && err == nil {
			chatUUID, err = models.InsertChat(sendUser.ID, reciverID)
			if err != nil {
				fmt.Println("create chat: ", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{
					"success": false,
				})
				return
			}
		} else if err != nil {
			fmt.Println("find chat: ", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"success": false,
			})
			return
		}

	}

	var dataReq struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&dataReq); err != nil {
		fmt.Println("decoding json at sendMessageHandler: ", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
		})
		return
	}

	if dataReq.Content == "" {
		fmt.Println("Empty message attempt")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Empty message not accepted",
		})
		return
	}

	err := models.InsertMessage(dataReq.Content, sendUser.ID, chatUUID)
	if err != nil {
		fmt.Println("InsertMessage error at sendMessageHandler", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
		})
		return
	}

	msg.MsgType = "sendMessage"
	msg.Updated = false
	msg.UserUUID = sendUser.UUID
	msg.PrivateMessage.Message.CreatedAt = time.Now()
	msg.PrivateMessage.Message.SenderUsername = sendUser.Username
	msg.PrivateMessage.Message.Content = dataReq.Content
	msg.PrivateMessage.Message.ChatUUID = chatUUID
	msg.ReciverUserUUID = reciverUserUUID
	msg.PrivateMessage.IsCreatedBy = true // change when sent to other user at Broadcast handler

	config.Broadcast <- msg

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Chat message sent",
	})
}

func ShowMessagesHandler(w http.ResponseWriter, r *http.Request) {
	loginStatus, user, _, validateErr := userManagementControllers.ValidateSession(w, r)

	var msg config.Message
	if !loginStatus || validateErr != nil {
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Not logged in",
		})
		return
	}

	msg.MsgType = "showMessages"
	msg.Updated = false
	msg.UserUUID = user.UUID
	chatUUID := r.URL.Query().Get("ChatUUID")
	msg.ReciverUserUUID = r.URL.Query().Get("UserUUID")
	var dataReq struct {
		NumberOfMessages int `json:"numberOfMessages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&dataReq); err != nil {
		fmt.Println(r.Body, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Error decoding JSON",
		})
		return
	}

	var err error
	msg.Messages, err = models.ReadAllMessages(chatUUID, dataReq.NumberOfMessages, user.ID)
	if err != nil {
		fmt.Println("Error reading messages", err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
		})
		return
	}

	// Write down if all the messages requested came through
	if len(msg.Messages) == dataReq.NumberOfMessages {
		msg.GotAllMessagesRequested = true
	}

	msg.ReceiverUserName, err = userModels.FindUsername(msg.ReciverUserUUID)

	if err != nil {
		fmt.Println("Error finding username", err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
		})
		return
	}

	config.Broadcast <- msg

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
	})
}
