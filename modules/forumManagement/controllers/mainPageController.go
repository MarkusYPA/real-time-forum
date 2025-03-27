package controller

import (
	"net/http"
	errorManagementControllers "real-time-forum/modules/errorManagement/controllers"
	"real-time-forum/modules/forumManagement/models"
	"real-time-forum/utils"
	"text/template"

	userManagementControllers "real-time-forum/modules/userManagement/controllers"
	userManagementModels "real-time-forum/modules/userManagement/models"

	_ "github.com/mattn/go-sqlite3"
)

const publicUrl = "modules/forumManagement/views/"

func MainPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.MethodNotAllowedError)
		return
	}

	if r.URL.Path != "/" {
		// If the URL is not exactly "/", respond with 404
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.NotFoundError)
		return
	}

	categories, err := models.ReadAllCategories()
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	loginStatus, loginUser, _, checkLoginError := userManagementControllers.ValidateSession(w, r)
	if checkLoginError != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	posts, err := models.ReadAllPosts(loginUser.ID)
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	data_obj_sender := struct {
		LoginUser  userManagementModels.User
		Posts      []models.Post
		Categories []models.Category
	}{
		LoginUser:  userManagementModels.User{},
		Posts:      posts,
		Categories: categories,
	}

	if loginStatus {
		data_obj_sender.LoginUser = loginUser
	}

	// Create a template with a function map
	tmpl, err := template.New("index.html").Funcs(template.FuncMap{
		"formatDate": utils.FormatDate, // Register function globally
	}).ParseFiles(
		publicUrl+"index.html",
		publicUrl+"templates/header.html",
		publicUrl+"templates/navbar.html",
		publicUrl+"templates/hero.html",
		publicUrl+"templates/posts.html",
		publicUrl+"templates/footer.html",
	)
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	err = tmpl.Execute(w, data_obj_sender)
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}
}
