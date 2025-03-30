package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"real-time-forum/config"
	errorManagementControllers "real-time-forum/modules/errorManagement/controllers"
	userModels "real-time-forum/modules/userManagement/models"
	"time"

	"github.com/gofrs/uuid/v5"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const publicUrl = "modules/userManagement/views/"

var u1 = uuid.Must(uuid.NewV4())

type AuthPageErrorData struct {
	ErrorMessage string
}

func SessionGenerator(w http.ResponseWriter, r *http.Request, userId int) string {
	session := &userModels.Session{
		UserId: userId,
	}
	session, insertError := userModels.InsertSession(session)
	if insertError != nil {
		errorManagementControllers.HandleErrorPage(w, r, errorManagementControllers.InternalServerError)
		return ""
	}
	SetCookie(w, session.SessionToken, session.ExpiresAt)
	// Set the session token in a cookie
	return session.SessionToken

}

// Middleware to check for valid user session in cookie
func ValidateSession(w http.ResponseWriter, r *http.Request) (bool, userModels.User, string, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println("in the ValidateSession")
		return false, userModels.User{}, "", err
	}

	sessionToken := cookie.Value
	user, expirationTime, selectError := userModels.SelectSession(sessionToken)
	if selectError != nil {
		if selectError.Error() == "sql: no rows in result set" {
			DeleteCookie(w, "session_token")
			return false, userModels.User{}, "", nil
		} else {
			return false, userModels.User{}, "", selectError
		}
	}

	// Check if the cookie has expired
	if time.Now().After(expirationTime) {
		// Cookie expired, redirect to login

		return false, userModels.User{}, "", nil
	}
	return true, user, sessionToken, nil
}

// handleSessionCheck is used to check if the user hasa valid session at first loading the page
func HandleSessionCheck(w http.ResponseWriter, r *http.Request) {
	loginStatus, _, sessionToken, _ := ValidateSession(w, r)

	if !loginStatus {
		json.NewEncoder(w).Encode(map[string]any{"loggedIn": false})
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"loggedIn": true, "token": sessionToken})
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		UsernameOrEmail string `json:"usernameOrEmail"`
		Password        string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&creds)
	hasFound, userID, err := userModels.AuthenticateUser(creds.UsernameOrEmail, creds.Password)
	if !hasFound {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	sessionToken := SessionGenerator(w, r, userID)
	json.NewEncoder(w).Encode(map[string]any{"success": true, "token": sessionToken})
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	loginStatus, user, sessionToken, _ := ValidateSession(w, r)
	if !loginStatus {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Not logged in",
		})
		return
	}
	err := userModels.DeleteSession(sessionToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "config Error",
		})
		return
	}

	DeleteCookie(w, "session_token")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})

	config.Mu.Lock()
	delete(config.Clients, user.UUID)
	config.Mu.Unlock()

	config.TellAllToUpdateClients()
}
func HandleMyProfile(w http.ResponseWriter, r *http.Request) {
	loginStatus, user, _, _ := ValidateSession(w, r)
	if !loginStatus {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Not logged in",
		})
		return
	}
	user.ID = 0
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"user":    user,
	})

}
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	// allow registering new user while logged in

	var creds userModels.User
	json.NewDecoder(r.Body).Decode(&creds)

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
			"message": "config error",
		})
		return
	}
	creds.Password = string(hashPass)
	// Insert a record while checking duplicates
	_, insertError := userModels.InsertUser(&creds)
	if insertError != nil {
		fmt.Println(insertError.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "User registration failed",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
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
