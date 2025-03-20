package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"real-time-forum/db"
	"sync"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

var (
	homeTmpl  *template.Template
	upgrader  = websocket.Upgrader{}
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan Post)
	mu        sync.Mutex
)

type Post struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
}

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

func setHandlers() {
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
	})
	http.HandleFunc("/", homeHandler)

	go handleBroadcasts()

	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/api/posts", handlePosts)
}

func makeTemplate() {
	var err error
	homeTmpl, err = template.ParseFiles("index.html")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func main() {

	err := db.OpenDB() // Open database connection
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.DB.Close()
	db.MakeTables()
	db.DataCleanup(6*time.Hour, db.RemoveExpiredSessions, "session")    // Clean up sessions every interval
	db.DataCleanup(12*time.Hour, db.RemoveUnusedCategories, "category") // Clean up categories every interval

	setHandlers()
	makeTemplate()

	fmt.Println("Server is running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
