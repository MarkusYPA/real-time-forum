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
	user, _, err := userModels.SelectSession(sessionToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
		})
		return
	}

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
		})
		return
	}

	defer conn.Close()
	Mu.Lock()
	Clients[user.UUID] = conn
	Mu.Unlock()
	TellAllToUpdateClients()

	for {
		var msg map[string]string
		err := conn.ReadJSON(&msg)

		if err != nil {
			break
		}

		// Broadcast typing event
		if msg["type"] == "typing" {
			if _, ok := Clients[msg["to"]]; ok {
				Clients[msg["to"]].WriteJSON(map[string]string{
					"msgType":  "typing",
					"userFrom": msg["from"],
				})
			}
		}
		if msg["type"] == "stopped_typing" {
			if _, ok := Clients[msg["to"]]; ok {
				Clients[msg["to"]].WriteJSON(map[string]string{
					"msgType":  "stopped_typing",
					"userFrom": msg["from"],
				})
			}
		}
	}

	Mu.Lock()
	delete(Clients, user.UUID)
	userModels.UpdateOnlineTime(user.UUID)
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

		// Broadcast to self
		if exists && msg.MsgType != "" {
			err := specificClient.WriteJSON(msg)
			if err != nil {
				specificClient.Close()
				Mu.Lock()
				delete(Clients, msg.UserUUID)
				userModels.UpdateOnlineTime(msg.UserUUID)
				Mu.Unlock()

				TellAllToUpdateClients()
			}

			if msg.MsgType == "listOfChat" || msg.MsgType == "showMessages" {
				continue
			}
		}

		// Broadcast to one recipient
		if msg.MsgType == "sendMessage" {

			msg.PrivateMessage.IsCreatedBy = false
			msg.SendNotification = true
			if receiverConn, ok := Clients[msg.ReciverUserUUID]; ok {
				Mu.Lock()
				err := receiverConn.WriteJSON(msg)
				Mu.Unlock()
				if err != nil {
					receiverConn.Close()
					delete(Clients, msg.ReciverUserUUID)
					userModels.UpdateOnlineTime(msg.ReciverUserUUID)
				}
			} // already checked receiver exists
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
				userModels.UpdateOnlineTime(uuid)
				TellAllToUpdateClients()
			}
		}
		Mu.Unlock()
	}
}
