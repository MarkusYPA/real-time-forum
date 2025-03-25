package main

import (
	"fmt"
	"net/http"
	"os"
	forumModels "real-time-forum/modules/forumManagement/models"
	"sync"
	"text/template"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

type Message struct {
	Post    forumModels.Post    `json:"post"`
	Comment forumModels.Comment `json:"comment"`
	MsgType string              `json:"msgType"`
	Updated bool                `json:"updated"`
}

var (
	homeTmpl  *template.Template
	upgrader  = websocket.Upgrader{}
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan Message)
	mu        sync.Mutex
)

const (
	titleMaxLen   int = 100
	contentMaxLen int = 3000
)

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

	http.HandleFunc("/api/like", likeHandler)
	http.HandleFunc("/api/dislike", dislikeHandler)
	http.HandleFunc("/api/addreply", replyHandler)
	http.HandleFunc("/api/replies", getRepliesHandler)
}

func main() {
	setHandlers()
	makeTemplate()
	fmt.Println("Server is running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
