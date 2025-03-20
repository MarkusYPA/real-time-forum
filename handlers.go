package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"real-time-forum/db"
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

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	json.NewDecoder(r.Body).Decode(&creds)

	var userID int
	err := db.DB.QueryRow("SELECT id FROM users WHERE username = ? AND password = ?", creds.Username, creds.Password).Scan(&userID)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create session
	sessionID, _ := uuid.NewV4() // TODO handle error
	_, err = db.DB.Exec("INSERT INTO sessions (session_id, user_id, expires_at) VALUES (?, ?, ?)", sessionID, userID, time.Now().Add(24*time.Hour))
	if err != nil {
		http.Error(w, "Session creation failed", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID.String(),
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		db.DB.Exec("DELETE FROM sessions WHERE session_id = ?", cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour),
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
	fmt.Println(creds)

	_, emailErr := mail.ParseAddress(creds.Email)
	if emailErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid email format")
	}

	hashPass, cryptErr := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if cryptErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Server error")
	}

	userId, idErr := uuid.NewV4() // Generate a new UUID user id
	if idErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Server error")
	}

	_, dbErr := db.DB.Exec("INSERT INTO users (id, username, age, gender, firstname, lastname, email, password) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", userId, creds.Username, creds.Age, creds.Gender, creds.FirstName, creds.LastName, creds.Email, hashPass)
	if dbErr != nil {
		// should return error at duplicate username or email
		fmt.Println(dbErr.Error())
		//http.Error(w, "User registration failed", http.StatusBadRequest)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("User registration failed")
		//return
	}

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// Handle new post submissions
func handleNewPost(w http.ResponseWriter, r *http.Request) {
	var post Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	db.InsertPost(w, post.Content)

	// Broadcast the new post
	broadcast <- post

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

// Get all posts
func handleGetPosts(w http.ResponseWriter, r *http.Request) {
	rows := db.GetPosts(w)
	var posts []Post
	for rows.Next() {
		var post Post
		rows.Scan(&post.ID, &post.Content)
		posts = append(posts, post)
	}
	json.NewEncoder(w).Encode(posts)
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
