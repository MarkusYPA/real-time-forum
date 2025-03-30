package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"real-time-forum/config"
	"real-time-forum/db"
	forumManagementControllers "real-time-forum/modules/forumManagement/controllers"
	userManagementControllers "real-time-forum/modules/userManagement/controllers"
)

func MakeTemplate() {
	var err error
	config.HomeTmpl, err = template.ParseFiles("index.html")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func SetHandlers() {
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
	})
	http.HandleFunc("/", config.HomeHandler)

	go config.HandleBroadcasts()

	http.HandleFunc("/api/category", forumManagementControllers.CategoryHandler)
	http.HandleFunc("/api/session", userManagementControllers.HandleSessionCheck)
	http.HandleFunc("/api/login", userManagementControllers.HandleLogin)
	http.HandleFunc("/api/register", userManagementControllers.HandleRegister)
	http.HandleFunc("/api/logout", userManagementControllers.HandleLogout)
	http.HandleFunc("/ws", config.HandleConnections)
	http.HandleFunc("/api/posts", forumManagementControllers.HandlePosts)
	http.HandleFunc("/api/like", forumManagementControllers.LikeHandler)
	http.HandleFunc("/api/dislike", forumManagementControllers.DislikeHandler)
	http.HandleFunc("/api/addreply", forumManagementControllers.ReplyHandler)
	http.HandleFunc("/api/replies", forumManagementControllers.GetRepliesHandler)
	http.HandleFunc("/api/sendmessage", forumManagementControllers.SendMessageHandler)
	http.HandleFunc("/api/showmessages", forumManagementControllers.ShowMessagesHandler)
	http.HandleFunc("/api/userslist", forumManagementControllers.GetUsersHandler)
}

func main() {
	db.ExecuteSQLFile("db/forum.sql")
	db.ExecuteSQLFile("db/whosonfirst.sql")
	db.ExecuteSQLFile("db/allyourbase.sql")

	SetHandlers()
	MakeTemplate()
	fmt.Println("Server is running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
