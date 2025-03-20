package controller

import (
	"fmt"
	errorManagementControllers "forum/modules/errorManagement/controllers"
	"forum/modules/forumManagement/models"
	"net/http"
	"strconv"

	userManagementControllers "forum/modules/userManagement/controllers"

	_ "github.com/mattn/go-sqlite3"
)

func ReadAllComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.MethodNotAllowedError)
		return
	}

	loginStatus, loginUser, _, checkLoginError := userManagementControllers.CheckLogin(w, r)
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

	// comments, err := models.ReadAllComments()
	// if err != nil {
	// 	errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
	// 	return
	// }

	// tmpl, err := template.ParseFiles(
	// 	publicUrl + "comments.html",
	// )
	// if err != nil {
	// 	errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
	// 	return
	// }

	// err = tmpl.Execute(w, comments)
	// if err != nil {
	// 	errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
	// 	return
	// }
}

func readPostComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.MethodNotAllowedError)
		return
	}

	loginStatus, loginUser, _, checkLoginError := userManagementControllers.CheckLogin(w, r)
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

	// comments, err := models.ReadCommentsByPostId()
	// if err != nil {
	// 	errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
	// 	return
	// }

	// tmpl, err := template.ParseFiles(
	// 	publicUrl + "post_comments.html",
	// )
	// if err != nil {
	// 	errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
	// 	return
	// }

	// err = tmpl.Execute(w, comments)
	// if err != nil {
	// 	errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
	// 	return
	// }
}

func createComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.MethodNotAllowedError)
		return
	}

	loginStatus, loginUser, _, checkLoginError := userManagementControllers.CheckLogin(w, r)
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

	// tmpl, err := template.ParseFiles(
	// 	publicUrl + "new_comment.html",
	// )
	// if err != nil {
	// 	errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
	// 	return
	// }

	// err = tmpl.Execute(w, nil)
	// if err != nil {
	// 	errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
	// 	return
	// }
}

func SubmitComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.MethodNotAllowedError)
		return
	}

	loginStatus, loginUser, _, checkLoginError := userManagementControllers.CheckLogin(w, r)
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

	err := r.ParseForm()
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.BadRequestError)
		return
	}
	post_id_str := r.FormValue("post_id")
	description := r.FormValue("description")
	if len(post_id_str) == 0 || len(description) == 0 {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.BadRequestError)
		return
	}

	post_id, err := strconv.Atoi(post_id_str)
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	// Insert a record while checking duplicates
	_, insertError := models.InsertComment(post_id, loginUser.ID, description)
	if insertError != nil {
		fmt.Println(insertError)
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	} else {
		fmt.Println("Comment added successfully!")
	}
	userManagementControllers.RedirectToPrevPage(w, r)

}

func LikeComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.MethodNotAllowedError)
		return
	}
	loginStatus, loginUser, _, checkLoginError := userManagementControllers.CheckLogin(w, r)
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

	err := r.ParseForm()
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.BadRequestError)
		return
	}
	commentID := r.FormValue("comment_id")
	commentIDInt, _ := strconv.Atoi(commentID)
	var Type string
	like := r.FormValue("like")
	dislike := r.FormValue("dislike")
	if like == "like" {
		Type = like
	} else if dislike == "dislike" {
		Type = dislike
	}

	existingLikeId, existingLikeType := models.CommentHasLiked(loginUser.ID, commentIDInt)
	if existingLikeId == -1 {
		models.InsertCommentLike(Type, commentIDInt, loginUser.ID)
		userManagementControllers.RedirectToPrevPage(w, r)
	} else {
		updateError := models.UpdateCommentLikesStatus(existingLikeId, "delete", loginUser.ID)
		if updateError != nil {
			errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
			return
		}

		if existingLikeType != Type { //this is duplicated like or duplicated dislike so we should update it to disable
			models.InsertCommentLike(Type, commentIDInt, loginUser.ID)

		}
		userManagementControllers.RedirectToPrevPage(w, r)
		return
	}
}

func UpdateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.MethodNotAllowedError)
		return
	}

	loginStatus, loginUser, _, checkLoginError := userManagementControllers.CheckLogin(w, r)
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

	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.BadRequestError)
		return
	}

	idStr := r.FormValue("comment_id")
	post_uuid := r.FormValue("post_uuid")
	description := r.FormValue("description")

	if len(idStr) == 0 || len(post_uuid) == 0 || len(description) == 0 {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.BadRequestError)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	comment := &models.Comment{
		ID:          id,
		Description: description,
		UserId:      loginUser.ID,
	}

	// Update a record while checking duplicates
	updateError := models.UpdateComment(comment, loginUser.ID, description)
	if updateError != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	} else {
		fmt.Println("Comment updated successfully!")
	}

	http.Redirect(w, r, "/post/"+post_uuid, http.StatusFound)
}

func DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.MethodNotAllowedError)
		return
	}

	loginStatus, loginUser, _, checkLoginError := userManagementControllers.CheckLogin(w, r)
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

	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.BadRequestError)
		return
	}

	idStr := r.FormValue("comment_id")
	post_uuid := r.FormValue("post_uuid")

	if len(idStr) == 0 || len(post_uuid) == 0 {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.BadRequestError)
		return
	}

	comment_id, err := strconv.Atoi(idStr)
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	// Update a record while checking duplicates
	updateError := models.UpdateCommentStatus(comment_id, "delete", loginUser.ID)
	if updateError != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	} else {
		fmt.Println("Post delete successfully!")
	}

	http.Redirect(w, r, "/post/"+post_uuid, http.StatusFound)
}
