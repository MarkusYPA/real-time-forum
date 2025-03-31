package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"real-time-forum/config"
	userModels "real-time-forum/modules/userManagement/models"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type AuthPageErrorData struct {
	ErrorMessage string
}

func SessionGenerator(w http.ResponseWriter, r *http.Request, userId int) (string, error) {
	session := &userModels.Session{
		UserId: userId,
	}
	session, insertError := userModels.InsertSession(session)
	if insertError != nil {
		return "", insertError
	}

	// Set the session token in a cookie
	SetCookie(w, session.SessionToken, session.ExpiresAt)

	return session.SessionToken, nil

}

// Middleware to check for valid user session in cookie
func ValidateSession(w http.ResponseWriter, r *http.Request) (bool, userModels.User, string, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		//fmt.Println("Getting cookie in ValidateSession:", err)
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
		return false, userModels.User{}, "", nil
	}
	return true, user, sessionToken, nil
}

// handleSessionCheck checks if the user has a valid session at first loading of the page
func HandleSessionCheck(w http.ResponseWriter, r *http.Request) {
	loginStatus, user, sessionToken, validateErr := ValidateSession(w, r)

	if !loginStatus || validateErr != nil {
		json.NewEncoder(w).Encode(map[string]any{
			"success":  false,
			"message":  "Not logged in",
			"loggedIn": false,
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"loggedIn": true, "token": sessionToken, "username": user.Username})
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		UsernameOrEmail string `json:"usernameOrEmail"`
		Password        string `json:"password"`
	}

	json.NewDecoder(r.Body).Decode(&creds)
	userID, err := userModels.AuthenticateUser(creds.UsernameOrEmail, creds.Password)

	if err != nil {
		fmt.Println("Error authenticating user:", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
		})
		return
	}

	sessionToken, sessionErr := SessionGenerator(w, r, userID)
	if sessionErr != nil {
		fmt.Println("Error creating session:", sessionErr.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
		})
	}

	username, err := userModels.FindUsernameByID(userID)
	if err != nil {
		fmt.Println("Error creating session:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
		})
	}

	json.NewEncoder(w).Encode(map[string]any{"success": true, "token": sessionToken, "username": username})
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	_, user, sessionToken, _ := ValidateSession(w, r)

	if sessionToken != "" {
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

		config.Mu.Lock()
		delete(config.Clients, user.UUID)
		config.Mu.Unlock()
	}

	json.NewEncoder(w).Encode(map[string]bool{"success": true})

	config.TellAllToUpdateClients()
}

func HandleMyProfile(w http.ResponseWriter, r *http.Request) {
	loginStatus, user, _, validateErr := ValidateSession(w, r)

	if !loginStatus || validateErr != nil {
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
		fmt.Println("Error parsing email", emailErr.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Invalid e-mail",
		})
		return
	}
	hashPass, cryptErr := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if cryptErr != nil {
		fmt.Println("Error hashing password", cryptErr.Error())
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
		fmt.Println("Error inserting user", insertError.Error())
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
