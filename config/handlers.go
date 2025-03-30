package config

import (
	"encoding/json"
	"log"
	"net/http"
	userModels "real-time-forum/modules/userManagement/models"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 Not Found", http.StatusNotFound) // error 404
		return
	}
	if r.Method == http.MethodGet {
		err := HomeTmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
	}
}

// Tell all connected Clients to update Clients list
func TellAllToUpdateClients() {
	var msg Message
	msg.MsgType = "updateClients"
	Broadcast <- msg
}

// Handle WebSocket connections
func HandleConnections(w http.ResponseWriter, r *http.Request) {
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
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket error:", err)
		return
	}
	defer conn.Close()
	Mu.Lock()
	Clients[user.UUID] = conn
	Mu.Unlock()
	TellAllToUpdateClients()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			//fmt.Println("Error reading message:", err)
			break
		}
	}
	Mu.Lock()
	delete(Clients, user.UUID)
	Mu.Unlock()

	TellAllToUpdateClients()
}

// Broadcast new posts
func HandleBroadcasts() {
	for {
		msg := <-Broadcast
		Mu.Lock()
		specificClient, exists := Clients[msg.UserUUID]
		Mu.Unlock()

		//sendOnlyToUser := false
		//fmt.Println("Message type on Broadcast", msg.MsgType, exists)

		// Broadcast to self
		if exists && msg.MsgType != "" {
			err := specificClient.WriteJSON(msg)
			if err != nil {
				specificClient.Close()
				Mu.Lock()
				delete(Clients, msg.UserUUID)
				Mu.Unlock()

				TellAllToUpdateClients()
			}

			if msg.MsgType == "listOfChat" || msg.MsgType == "showMessages" {
				//sendOnlyToUser = true
				continue
			}
		}

		// Broadcast to one recipient
		if msg.MsgType == "sendMessage" {

			msg.PrivateMessage.IsCreatedBy = false
			msg.SendNotoification = true
			if receiverConn, ok := Clients[msg.ReciverUserUUID]; ok {
				Mu.Lock()
				err := receiverConn.WriteJSON(msg)
				Mu.Unlock()
				if err != nil {
					receiverConn.Close()
					delete(Clients, msg.ReciverUserUUID)
				}
			} else {
				// Send error message "no receiver found" to original client?
			}
			continue
		}

		// Broadcast to all other Clients
		Mu.Lock()
		for uuid, client := range Clients {

			if uuid == msg.UserUUID {
				continue
			}

			msg.Comment.IsLikedByUser = false
			msg.Comment.IsDislikedByUser = false
			msg.Post.IsDislikedByUser = false
			msg.Post.IsLikedByUser = false
			msg.IsLikAction = false

			Mu.Unlock()
			err := client.WriteJSON(msg)
			Mu.Lock()

			if err != nil {
				client.Close()
				delete(Clients, uuid)

				TellAllToUpdateClients()
			}
		}
		Mu.Unlock()
	}
}
