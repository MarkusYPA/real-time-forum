package main

import (
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
	ID         int      `json:"id"`
	Title      string   `json:"title"`
	Author     string   `json:"author"`
	Date       string   `json:"date"`
	Content    string   `json:"content"`
	Categories []string `json:"categories"`
}

func makeTemplate() {
	var err error
	homeTmpl, err = template.ParseFiles("index.html")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func setHandlers() {
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
	})
	http.HandleFunc("/", homeHandler)

	go handleBroadcasts()

	http.HandleFunc("/api/session", handleSessionCheck)
	http.HandleFunc("/api/login", handleLogin)
	http.HandleFunc("/api/register", handleRegister)
	http.HandleFunc("/api/logout", handleLogout)
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/api/posts", handlePosts)
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
