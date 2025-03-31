package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"real-time-forum/config"
	errorManagementControllers "real-time-forum/modules/errorManagement/controllers"
	"real-time-forum/modules/forumManagement/models"
	forumModels "real-time-forum/modules/forumManagement/models"
	userManagementControllers "real-time-forum/modules/userManagement/controllers"
	userManagementModels "real-time-forum/modules/userManagement/models"
	"real-time-forum/utils"
	"strconv"
	"strings"
	"text/template"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func HandlePosts(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		HandleNewPost(w, r) // Call function to process new post
		return
	}
	if r.Method == http.MethodGet {
		HandleGetPosts(w, r) // Call function to fetch all posts
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(map[string]any{
		"success": false,
	})
}

// Get all posts
func HandleGetPosts(w http.ResponseWriter, r *http.Request) {
	loginStatus, user, _, validateErr := userManagementControllers.ValidateSession(w, r)

	if !loginStatus || validateErr != nil {
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Not logged in",
		})
		return
	}

	// Get category from query
	categoryIdString := r.URL.Query().Get("categoryid")
	if categoryIdString == "" {
		fmt.Println("faulty category id:", categoryIdString)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Missing category",
		})
		return
	}

	catId, err := strconv.Atoi(categoryIdString)
	if err != nil {
		fmt.Println("faulty category id:", categoryIdString, catId, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Missing category",
		})
		return
	}

	var posts []forumModels.Post
	if catId == 0 {
		posts, err = forumModels.ReadAllPosts(user.ID)
	} else if catId > 0 {
		posts, err = forumModels.ReadPostsByCategoryId(user.ID, catId)
	}

	if err != nil {
		fmt.Println("error getting posts:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Server error",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"posts":   posts,
	})
}

// Handle new post submissions
func HandleNewPost(w http.ResponseWriter, r *http.Request) {
	loginStatus, user, _, validateErr := userManagementControllers.ValidateSession(w, r)

	if !loginStatus || validateErr != nil {
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Not logged in",
		})
		return
	}

	var msg config.Message
	var requestData struct {
		Title      string `json:"title"`
		Content    string `json:"content"`
		Categories []int  `json:"categoryIds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		fmt.Println("json parse error:", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Invalid request",
		})
		return
	}

	if requestData.Title == "" || requestData.Content == "" || len(requestData.Categories) == 0 {
		fmt.Println("Missing title, content or categories in post")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Invalid request",
		})
		return
	}

	// Trim input
	title := strings.TrimSpace(requestData.Title)
	description := strings.TrimSpace(requestData.Content) // Displaying as texcontent prevents execution

	// Create a Post struct
	msg.MsgType = "post"
	msg.Updated = false
	msg.Post = forumModels.Post{
		Title:       title,
		Description: description,
		CreatedAt:   time.Now(),
		User:        user,
	}
	msg.UserUUID = user.UUID

	// Store post in DB
	var err error
	msg.Post.ID, err = forumModels.InsertPost(&msg.Post, requestData.Categories)
	if err != nil {
		fmt.Println("error inserting post:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
		})
		return
	}

	msg.Post.Categories, err = forumModels.ReadCategoriesByPostId(msg.Post.ID)
	if err != nil {
		fmt.Println("error reading categories:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
		})
		return
	}

	// Broadcast the post
	config.Broadcast <- msg

	// Send response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
	})

}

func CategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fmt.Println("Wrong method on getting categories")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
		})
		return
	}

	categories, err := forumModels.ReadAllCategories()
	if err != nil {
		fmt.Println("Error reading categories:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
		})
		return
	}

	type dataToSend struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}

	var data []dataToSend
	for i := range categories {
		data = append(data, dataToSend{Id: categories[i].ID, Name: categories[i].Name})
	}

	json.NewEncoder(w).Encode(map[string]any{
		"success":    true,
		"categories": data,
	})

}

func FilterPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.MethodNotAllowedError)
		return
	}
	searchTerm := r.URL.Query().Get("post_info")

	categories, err := models.ReadAllCategories()
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	posts, err := models.FilterPosts(searchTerm)
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	data_obj_sender := struct {
		LoginUser  userManagementModels.User
		Posts      []models.Post
		Categories []models.Category
		SearchTerm string
	}{
		LoginUser:  userManagementModels.User{},
		Posts:      posts,
		Categories: categories,
		SearchTerm: searchTerm,
	}

	loginStatus, loginUser, _, checkLoginError := userManagementControllers.ValidateSession(w, r)
	if checkLoginError != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}
	if loginStatus {
		data_obj_sender.LoginUser = loginUser
	}

	// Create a template with a function map
	tmpl, err := template.New("search_posts.html").Funcs(template.FuncMap{
		"formatDate": utils.FormatDate, // Register function globally
	}).ParseFiles(
		publicUrl+"search_posts.html",
		publicUrl+"templates/header.html",
		publicUrl+"templates/navbar.html",
		publicUrl+"templates/hero.html",
		publicUrl+"templates/posts.html",
		publicUrl+"templates/footer.html",
	)
	if err != nil {
		fmt.Println(err)
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	err = tmpl.Execute(w, data_obj_sender)
	if err != nil {
		fmt.Println(err)
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}
}

func ReadMyCreatedPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.MethodNotAllowedError)
		return
	}

	if r.URL.Path != "/myCreatedPosts/" {
		// If the URL is not exactly "/myCreatedPosts/", respond with 404
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.NotFoundError)
		return
	}

	loginStatus, loginUser, _, checkLoginError := userManagementControllers.ValidateSession(w, r)
	if checkLoginError != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}
	if loginStatus {
		fmt.Println("logged in userid is: ", loginUser.ID)
	} else {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.UnauthorizedError)
		return
	}

	categories, err := models.ReadAllCategories()
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	posts, err := models.ReadPostsByUserId(loginUser.ID)
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	data_obj_sender := struct {
		LoginUser  userManagementModels.User
		Posts      []models.Post
		Categories []models.Category
	}{
		LoginUser:  loginUser,
		Posts:      posts,
		Categories: categories,
	}

	// Create a template with a function map
	tmpl, err := template.New("my_created_posts.html").Funcs(template.FuncMap{
		"formatDate": utils.FormatDate, // Register function globally
	}).ParseFiles(
		publicUrl+"my_created_posts.html",
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

func ReadMyLikedPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.MethodNotAllowedError)
		return
	}

	if r.URL.Path != "/myLikedPosts/" {
		// If the URL is not exactly "/myLikedPosts/", respond with 404
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.NotFoundError)
		return
	}

	loginStatus, loginUser, _, checkLoginError := userManagementControllers.ValidateSession(w, r)
	if checkLoginError != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}
	if loginStatus {
		fmt.Println("logged in userid is: ", loginUser.ID)
	} else {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.UnauthorizedError)
		return
	}

	categories, err := models.ReadAllCategories()
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	posts, err := models.ReadPostsLikedByUserId(loginUser.ID)
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	data_obj_sender := struct {
		LoginUser  userManagementModels.User
		Posts      []models.Post
		Categories []models.Category
	}{
		LoginUser:  loginUser,
		Posts:      posts,
		Categories: categories,
	}

	// Create a template with a function map
	tmpl, err := template.New("my_liked_posts.html").Funcs(template.FuncMap{
		"formatDate": utils.FormatDate, // Register function globally
	}).ParseFiles(
		publicUrl+"my_liked_posts.html",
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

func ReadPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.MethodNotAllowedError)
		return
	}

	loginStatus, loginUser, _, checkLoginError := userManagementControllers.ValidateSession(w, r)
	if checkLoginError != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}
	if loginStatus {
		fmt.Println("logged in userid is: ", loginUser.ID)
		// return
	} else {
		fmt.Println("user is not logged in")
	}

	uuid, errUrl := utils.ExtractUUIDFromUrl(r.URL.Path, "post")
	if errUrl == "not found" {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.NotFoundError)
		return
	}

	post, err := models.ReadPostByUUID(uuid, loginUser.ID)
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	data_obj_sender := struct {
		LoginUser userManagementModels.User
		Post      models.Post
		Comments  []models.Comment
	}{
		LoginUser: loginUser,
		Post:      post,
		Comments:  nil,
	}

	if loginStatus {
		comments, err := models.ReadAllCommentsForPostByUserID(post.ID, loginUser.ID)
		if err != nil {
			fmt.Println(err)
			errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
			return
		}

		data_obj_sender.Comments = comments
	} else {
		comments, err := models.ReadAllCommentsForPost(post.ID)
		if err != nil {
			errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
			return
		}

		data_obj_sender.Comments = comments
	}

	// Create a template with a function map
	tmpl, err := template.New("post_details.html").Funcs(template.FuncMap{
		"formatDate": utils.FormatDate, // Register function globally
	}).ParseFiles(
		publicUrl+"post_details.html",
		publicUrl+"templates/header.html",
		publicUrl+"templates/navbar.html",
		publicUrl+"templates/footer.html",
	)
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	// Execute template with data
	err = tmpl.Execute(w, data_obj_sender)
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
	}
}

func EditPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.MethodNotAllowedError)
		return
	}

	loginStatus, loginUser, _, checkLoginError := userManagementControllers.ValidateSession(w, r)
	if checkLoginError != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}
	if loginStatus {
		fmt.Println("logged in userid is: ", loginUser.ID)
		// return
	} else {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.UnauthorizedError)
		return
	}

	uuid, errUrl := utils.ExtractUUIDFromUrl(r.URL.Path, "editPost")
	if errUrl == "not found" {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.NotFoundError)
		return
	}

	post, err := models.ReadPostByUUID(uuid, loginUser.ID)
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	categories, err := models.ReadAllCategories()
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	data_obj_sender := struct {
		LoginUser  userManagementModels.User
		Post       models.Post
		Categories []models.Category
	}{
		LoginUser:  loginUser,
		Post:       post,
		Categories: categories,
	}

	tmpl, err := template.ParseFiles(
		publicUrl+"edit_post.html",
		publicUrl+"templates/header.html",
		publicUrl+"templates/navbar.html",
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

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.MethodNotAllowedError)
		return
	}

	loginStatus, loginUser, _, checkLoginError := userManagementControllers.ValidateSession(w, r)
	if checkLoginError != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}
	if loginStatus {
		fmt.Println("logged in userid is: ", loginUser.ID)
		// return
	} else {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.UnauthorizedError)
		return
	}

	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.BadRequestError)
		return
	}

	idStr := r.FormValue("id")
	uuid := r.FormValue("uuid")
	title := r.FormValue("title")
	description := r.FormValue("description")
	categories := r.Form["categories"]

	if len(idStr) == 0 || len(title) == 0 || len(description) == 0 || len(categories) == 0 {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.BadRequestError)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	post := &models.Post{
		ID:          id,
		Title:       title,
		Description: description,
		UserId:      loginUser.ID,
	}

	// Convert the string slice to an int slice
	categoryIds := make([]int, 0, len(categories))
	for _, category := range categories {
		if id, err := strconv.Atoi(category); err == nil {
			categoryIds = append(categoryIds, id)
		} else {
			// Handle error if conversion fails (for example, invalid input)
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}
	}

	// Update a record while checking duplicates
	updateError := models.UpdatePost(post, categoryIds, loginUser.ID)
	if updateError != nil {
		if errors.Is(updateError, sql.ErrNoRows) {
			// todo show toast
			fmt.Println("Post already exists!")
		} else {
			errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		}
		return
	} else {
		fmt.Println("Post updated successfully!")
	}

	http.Redirect(w, r, "/post/"+uuid, http.StatusFound)
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.MethodNotAllowedError)
		return
	}

	loginStatus, loginUser, _, checkLoginError := userManagementControllers.ValidateSession(w, r)
	if checkLoginError != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}
	if loginStatus {
		fmt.Println("logged in userid is: ", loginUser.ID)
		// return
	} else {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.UnauthorizedError)
		return
	}

	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.BadRequestError)
		return
	}

	idStr := r.FormValue("id")

	if len(idStr) == 0 {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.BadRequestError)
		return
	}

	post_id, err := strconv.Atoi(idStr)
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	// Update a record while checking duplicates
	updateError := models.UpdateStatusPost(post_id, "delete", loginUser.ID)
	if updateError != nil {
		if errors.Is(updateError, sql.ErrNoRows) {
			// todo show toast
			fmt.Println("Post already exists!")
		} else {
			errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		}
		return
	} else {
		fmt.Println("Post delete successfully!")
	}

	//userManagementControllers.RedirectToIndex(w, r)
}

func LikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.MethodNotAllowedError)
		return
	}
	loginStatus, loginUser, _, checkLoginError := userManagementControllers.ValidateSession(w, r)
	if checkLoginError != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}
	if loginStatus {
		fmt.Println("logged in userid is: ", loginUser.ID)
	} else {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.UnauthorizedError)
		return
	}

	err := r.ParseForm()
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.BadRequestError)
		return
	}
	postID := r.FormValue("post_id")
	postIDInt, _ := strconv.Atoi(postID)
	var Type string
	like := r.FormValue("like_post")
	dislike := r.FormValue("dislike_post")
	if like == "like" {
		Type = like
	} else if dislike == "dislike" {
		Type = dislike
	}

	existingLikeId, existingLikeType := models.PostHasLike(loginUser.ID, postIDInt)

	if existingLikeId == -1 {
		post := &models.PostLike{
			Type:   Type,
			PostId: postIDInt,
			UserId: loginUser.ID,
		}
		_, insertError := models.InsertPostLike(post)
		if insertError != nil {
			errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
			return
		}
		//userManagementControllers.RedirectToPrevPage(w, r)
	} else {
		updateError := models.UpdateStatusPostLike(existingLikeId, "delete", loginUser.ID)
		if updateError != nil {
			errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
			return
		}

		if existingLikeType != Type { //this is duplicated like or duplicated dislike so we should update it to disable
			post := &models.PostLike{
				Type:   Type,
				PostId: postIDInt,
				UserId: loginUser.ID,
			}
			_, insertError := models.InsertPostLike(post)
			if insertError != nil {
				errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
				return
			}
		}
		//userManagementControllers.RedirectToPrevPage(w, r)
		return
	}
}
