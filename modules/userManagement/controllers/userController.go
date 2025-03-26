package controller

import (
	"fmt"
	"net/http"
	errorManagementControllers "real-time-forum/modules/errorManagement/controllers"
	"real-time-forum/modules/userManagement/models"
	"time"

	"github.com/gofrs/uuid/v5"
	_ "github.com/mattn/go-sqlite3"
)

const publicUrl = "modules/userManagement/views/"

var u1 = uuid.Must(uuid.NewV4())

type AuthPageErrorData struct {
	ErrorMessage string
}

/* func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.MethodNotAllowedError)
		return
	}

	loginStatus, user, _, checkLoginError := CheckLogin(w, r)
	if checkLoginError != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}
	if loginStatus {
		fmt.Println("logged in userid is: ", user.ID)
		RedirectToIndex(w, r)
		return
	}
	err := r.ParseForm()
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.BadRequestError)
		return
	}
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	if len(username) == 0 || len(email) == 0 || len(password) == 0 {
		// errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.BadRequestError)
		renderAuthPage(w, "Username, email and password are required.")
		return
	}
	if !strings.Contains(email, "@") {
		renderAuthPage(w, "Invalid email address!")
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}

	newUser := &models.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}

	// Insert a record while checking duplicates
	userId, insertError := models.InsertUser(newUser)
	if insertError != nil {
		if insertError.Error() == "duplicateEmail" {
			renderAuthPage(w, "User with this email already exists!")
			return
		} else if insertError.Error() == "duplicateUsername" {
			renderAuthPage(w, "User with this username already exists!")
			return
		} else {
			errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		}
		return
	} else {
		fmt.Println("User added successfully!")
	}

	SessionGenerator(w, r, userId)

	RedirectToIndex(w, r)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.MethodNotAllowedError)
		return
	}

	loginStatus, user, _, checkLoginError := CheckLogin(w, r)
	if checkLoginError != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}
	if loginStatus {
		fmt.Println("logged in userid is: ", user.ID)
		RedirectToIndex(w, r)
		return
	}

	err := r.ParseForm()
	if err != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.BadRequestError)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	if len(username) == 0 || len(password) == 0 {
		// errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.BadRequestError)
		// return
		renderAuthPage(w, "Username and password are required.")
		return
	}

	// Insert a record while checking duplicates
	authStatus, userId, authError := models.AuthenticateUser(username, password)
	if authError != nil {
		// errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		renderAuthPage(w, authError.Error())
		return
	} else if authStatus {
		fmt.Println("User logged in successfully!")
	}

	SessionGenerator(w, r, userId)

	RedirectToIndex(w, r)
}

// Render the login page with an optional error message
func renderAuthPage(w http.ResponseWriter, errorMsg string) {
	tmpl := template.Must(template.ParseFiles(publicUrl + "authPage.html"))
	tmpl.Execute(w, AuthPageErrorData{ErrorMessage: errorMsg})
} */

func SessionGenerator(w http.ResponseWriter, r *http.Request, userId int) {
	session := &models.Session{
		UserId: userId,
	}
	session, insertError := models.InsertSession(session)
	if insertError != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return
	}
	SetCookie(w, session.SessionToken, session.ExpiresAt)
	fmt.Println("cookie SET")
	// Set the session token in a cookie

}

// Middleware to check for valid user session in cookie
func ValidateSession(w http.ResponseWriter, r *http.Request) (bool, models.User, string, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println("in the ValidateSession")
		return false, models.User{}, "", err
	}

	sessionToken := cookie.Value
	user, expirationTime, selectError := models.SelectSession(sessionToken)
	if selectError != nil {
		if selectError.Error() == "sql: no rows in result set" {
			DeleteCookie(w, "session_token")
			return false, models.User{}, "", nil
		} else {
			return false, models.User{}, "", selectError
		}
	}

	// Check if the cookie has expired
	if time.Now().After(expirationTime) {
		// Cookie expired, redirect to login

		return false, models.User{}, "", nil
	}
	fmt.Println("user:", user)
	return true, user, sessionToken, nil
}

func RedirectToIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusFound)
}

func RedirectToHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/home/", http.StatusFound)
}

func RedirectToPrevPage(w http.ResponseWriter, r *http.Request) {
	referrer := r.Header.Get("Referer")
	if referrer == "" {
		referrer = "/"
	}

	// Redirect back to the original page to reload it
	http.Redirect(w, r, referrer, http.StatusSeeOther)
}

func DeleteCookie(w http.ResponseWriter, cookieName string) {
	http.SetCookie(w, &http.Cookie{
		Name:    cookieName,
		Value:   "",              // Optional but recommended
		Expires: time.Unix(0, 0), // Set expiration to a past date
		MaxAge:  -1,              // Ensure immediate removal
		Path:    "/",             // Must match the original cookie path
	})
}

func SetCookie(w http.ResponseWriter, sessionToken string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   false,
	})
}
