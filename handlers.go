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
	expiresAt := time.Now().Add(30 * time.Minute)

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
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	post.ID, post.Date = db.InsertPost(w, userName, usrId, post.Title, post.Content)
	db.InsertCategories(post.Categories, post.ID)

	day, time, _ := timeStrings(post.Date)
	post.Date = day + " " + time
	post.Author = userName

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
		day, time, _ := timeStrings(posts[i].Date)
		posts[i].Date = day + " " + time
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
