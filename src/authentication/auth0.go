package Handlers

import (
	"01connecthub/src/security"
	"01connecthub/src/server"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	html "html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

var (
	oauthConfigGithub = &oauth2.Config{
		ClientID:     "Ov23liYl2L30q080Nif5",
		ClientSecret: "b8815439273a6e564a6f0d7f82c775c3d7e2383a",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"user"},
		Endpoint:     github.Endpoint,
	}
	oauthStateStringGit = "randomstring"
)

var templates = html.Must(html.ParseGlob("templates/*.html"))

func LoginPageGit(w http.ResponseWriter, r *http.Request) {
	url := oauthConfigGithub.AuthCodeURL(oauthStateStringGit)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func Callback(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != oauthStateStringGit {
		http.Error(w, "State is invalid", http.StatusBadRequest)
		return
	}

	token, err := oauthConfigGithub.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		http.Error(w, "Code exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := oauthConfigGithub.Client(context.Background(), token)
	userInfo, err := client.Get("https://api.github.com/user")
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer userInfo.Body.Close()

	var user map[string]interface{}
	if err := json.NewDecoder(userInfo.Body).Decode(&user); err != nil {
		http.Error(w, "Failed to decode user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	emailInfo, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		http.Error(w, "Failed to get user emails: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer emailInfo.Body.Close()

	var emails []map[string]interface{}
	if err := json.NewDecoder(emailInfo.Body).Decode(&emails); err != nil {
		http.Error(w, "Failed to decode user emails: "+err.Error(), http.StatusInternalServerError)
		return
	}

	primaryEmail := ""
	for _, email := range emails {
		if isPrimary, ok := email["primary"].(bool); ok && isPrimary {
			if emailStr, ok := email["email"].(string); ok {
				primaryEmail = emailStr
				break
			}
		}
	}

	if primaryEmail == "" {
		http.Error(w, "No primary email found", http.StatusInternalServerError)
		return
	}

	// Determine role_id based on email
	var roleID int
	if primaryEmail == "sayedahmed97.sad@gmail.com" {
		roleID = 1 // Admin
	} else if primaryEmail == "qassimhassan9@gmail.com" {
		roleID = 2 // Moderator
	} else {
		roleID = 3 // Regular user
	}

	username := strings.Split(primaryEmail, "@")[0]
	hashed, err := server.HashPassword("github_oauth") // Use a short consistent password
	if err != nil {
		http.Error(w, "Error hashing password: "+err.Error(), http.StatusInternalServerError)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Database connection failed")
		errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		server.ErrHandler(w, r, &errData)
		return
	}
	defer db.Close()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		log.Println("Failed to begin transaction:", err)
		errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		server.ErrHandler(w, r, &errData)
		return
	}

	// Generate session token
	sessionToken, err := security.GenerateToken()
	if err != nil {
		log.Println("Failed to generate session token:", err)
		tx.Rollback()
		errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		server.ErrHandler(w, r, &errData)
		return
	}

	// Check if user exists
	var userExists bool
	var userID int

	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM user WHERE email = ?)", primaryEmail).Scan(&userExists)
	if err != nil {
		log.Println("Error checking if user exists:", err)
		tx.Rollback()
		errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		server.ErrHandler(w, r, &errData)
		return
	}

	defaultAvatar := "static/assets/default-avatar.png"
	githubAvatar, _ := user["avatar_url"].(string)
	if githubAvatar == "" {
		githubAvatar = defaultAvatar
	}

	if userExists {
		// User exists - update user information and session
		err = db.QueryRow("SELECT userid FROM user WHERE email = ?", primaryEmail).Scan(&userID)
		if err != nil {
			log.Println("Error getting user ID:", err)
			tx.Rollback()
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}

		// Update user details
		_, err = tx.Exec("UPDATE user SET provider = ?, current_session = ? WHERE userid = ?",
			"Github", sessionToken, userID)
		if err != nil {
			log.Println("Error updating user:", err)
			tx.Rollback()
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}

		// Check if GitHub record exists
		var githubExists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM github WHERE user_userid = ?)", userID).Scan(&githubExists)
		if err != nil {
			log.Println("Error checking if GitHub record exists:", err)
			tx.Rollback()
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}

		githubID, ok := user["id"].(float64)
		if !ok {
			http.Error(w, "GitHub ID is missing or invalid", http.StatusInternalServerError)
			tx.Rollback()
			return
		}
		githubIDString := fmt.Sprintf("%.0f", githubID)

		if githubExists {
			// Update GitHub record
			_, err = tx.Exec("UPDATE github SET gitAvatar = ? WHERE user_userid = ?",
				githubAvatar, userID)
			if err != nil {
				log.Println("Error updating GitHub record:", err)
				tx.Rollback()
				errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				server.ErrHandler(w, r, &errData)
				return
			}
		} else {
			// Insert GitHub record
			_, err = tx.Exec("INSERT INTO github (gituserid, gitF_name, gitL_name, gitUsername, gitEmail, gitpassword, gitAvatar, user_userid) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
				githubIDString, "", "", username, primaryEmail, hashed, githubAvatar, userID)
			if err != nil {
				log.Println("Error inserting GitHub record:", err)
				tx.Rollback()
				errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				server.ErrHandler(w, r, &errData)
				return
			}
		}
	} else {
		// Insert new user
		res, err := tx.Exec("INSERT INTO user (F_name, L_name, Username, Email, password, current_session, role_id, Avatar, provider) VALUES (?, ?, ?, ?, ?, ?, ?, ?,?)",
			username, "", username, primaryEmail, hashed, sessionToken.String(), roleID, githubAvatar, "Github")
		if err != nil {
			log.Println("Failed to insert user data:", err)
			tx.Rollback()
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}

		lastID, err := res.LastInsertId()
		if err != nil {
			log.Println("Failed to get last insert ID:", err)
			tx.Rollback()
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}

		userID = int(lastID)

		githubID, ok := user["id"].(float64)
		if !ok {
			http.Error(w, "GitHub ID is missing or invalid", http.StatusInternalServerError)
			tx.Rollback()
			return
		}
		githubIDString := fmt.Sprintf("%.0f", githubID)

		// Insert GitHub record
		_, err = tx.Exec("INSERT INTO github (gituserid, gitF_name, gitL_name, gitUsername, gitEmail, gitpassword, gitAvatar, user_userid) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			githubIDString, "", "", username, primaryEmail, hashed, githubAvatar, userID)
		if err != nil {
			log.Println("Failed to insert GitHub data:", err)
			tx.Rollback()
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Println("Failed to commit transaction:", err)
		errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		server.ErrHandler(w, r, &errData)
		return
	}

	// Create session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken.String(),
		Expires:  time.Now().Add(1 * time.Hour),
		HttpOnly: true,
	})

	// Insert or update session record
	_, err = db.Exec("INSERT OR REPLACE INTO session (sessionid, userid, endtime) VALUES (?, ?, ?)",
		sessionToken.String(), userID, time.Now().Add(1*time.Hour))
	if err != nil {
		log.Println("Error creating session:", err)
		// Non-fatal error, continue
	}

	// Redirect to home page
	http.Redirect(w, r, "/home?tab=posts&filter=all", http.StatusSeeOther)
}

var (
	oauthConfigGoogle = &oauth2.Config{
		ClientID:     "308640975130-ejeugakggjsq2n0gco97minddk4b8elt.apps.googleusercontent.com",
		ClientSecret: "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",
		RedirectURL:  "http://localhost:8080/callbackGoogle",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
	oauthStateStringGoogle = "randomstring"
)

func LoginPageGoogle(w http.ResponseWriter, r *http.Request) {
	url := oauthConfigGoogle.AuthCodeURL(oauthStateStringGoogle)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
func CallbackGoogle(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != oauthStateStringGoogle {
		http.Error(w, "State is invalid", http.StatusBadRequest)
		return
	}

	token, err := oauthConfigGoogle.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		http.Error(w, "Code exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := oauthConfigGoogle.Client(context.Background(), token)
	userInfo, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer userInfo.Body.Close()

	var user map[string]interface{}
	err = json.NewDecoder(userInfo.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Failed to decode user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	googleID, ok := user["id"].(string)
	if !ok {
		http.Error(w, "Google ID is missing", http.StatusInternalServerError)
		return
	}

	email, ok := user["email"].(string)
	if !ok {
		http.Error(w, "Email is missing", http.StatusInternalServerError)
		return
	}

	// Determine role_id based on email
	var roleID int
	if email == "sayedahmed97.sad@gmail.com" {
		roleID = 1 // Admin
	} else if email == "qassimhassan9@gmail.com" {
		roleID = 2 // Moderator
	} else {
		roleID = 3 // Regular user
	}

	firstName, _ := user["given_name"].(string)
	lastName, _ := user["family_name"].(string)
	profilePicture, _ := user["picture"].(string)
	username := strings.Split(email, "@")[0]

	// Use a short password for OAuth logins
	hashed, err := server.HashPassword("google_oauth")
	if err != nil {
		http.Error(w, "Error hashing password: "+err.Error(), http.StatusInternalServerError)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Database connection failed")
		errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		server.ErrHandler(w, r, &errData)
		return
	}
	defer db.Close()

	// Begin a transaction to ensure database consistency
	tx, err := db.Begin()
	if err != nil {
		log.Println("Failed to begin transaction:", err)
		errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		server.ErrHandler(w, r, &errData)
		return
	}

	// Generate a session token
	sessionToken, err := security.GenerateToken()
	if err != nil {
		log.Println("Failed to generate session token:", err)
		tx.Rollback()
		errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		server.ErrHandler(w, r, &errData)
		return
	}

	// Check if user exists with this email
	var userExists bool
	var userID int

	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM user WHERE email = ?)", email).Scan(&userExists)
	if err != nil {
		log.Println("Error checking if user exists:", err)
		tx.Rollback()
		errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		server.ErrHandler(w, r, &errData)
		return
	}

	if userExists {
		// User exists - update user information and session
		err = db.QueryRow("SELECT userid FROM user WHERE email = ?", email).Scan(&userID)
		if err != nil {
			log.Println("Error getting user ID:", err)
			tx.Rollback()
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}

		// Update user details
		_, err = tx.Exec("UPDATE user SET   provider = ?, current_session = ? WHERE userid = ?",
			 "Google", sessionToken,  userID)
		if err != nil {
			log.Println("Error updating user:", err)
			tx.Rollback()
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}

		// Check if Google record exists for this user
		var googleExists bool

		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM google WHERE user_userid = ?)", userID).Scan(&googleExists)
		if err != nil {
			log.Println("Error checking if Google record exists:", err)
			tx.Rollback()
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}

		if googleExists {
			// Update Google record
			_, err = tx.Exec("UPDATE google SET googleF_name = ?, googleL_name = ?, googleAvatar = ? WHERE user_userid = ?",
				firstName, lastName, profilePicture, userID)
			if err != nil {
				log.Println("Error updating Google record:", err)
				tx.Rollback()
				errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				server.ErrHandler(w, r, &errData)
				return
			}
		} else {
			// Insert Google record
			_, err = tx.Exec("INSERT INTO google (google_api_id, googleF_name, googleL_name, googleUsername, googleEmail, googlepassword, googleAvatar, user_userid) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
				googleID, firstName, lastName, username, email, hashed, profilePicture, userID)
			if err != nil {
				log.Println("Error inserting Google record:", err)
				tx.Rollback()
				errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				server.ErrHandler(w, r, &errData)
				return
			}
		}
	} else {
		// Insert new user
		res, err := tx.Exec("INSERT INTO user (F_name, L_name, Username, Email, password, current_session, role_id, Avatar, provider) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			firstName, lastName, username, email, hashed, sessionToken.String(), roleID, profilePicture, "Google")
		if err != nil {
			log.Println("Failed to insert user data:", err)
			tx.Rollback()
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}

		lastID, err := res.LastInsertId()
		if err != nil {
			log.Println("Failed to get last insert ID:", err)
			tx.Rollback()
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}

		userID = int(lastID)

		// Insert Google record
		_, err = tx.Exec("INSERT INTO google (google_api_id, googleF_name, googleL_name, googleUsername, googleEmail, googlepassword, googleAvatar, user_userid) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			googleID, firstName, lastName, username, email, hashed, profilePicture, userID)
		if err != nil {
			log.Println("Failed to insert Google data:", err)
			tx.Rollback()
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Println("Failed to commit transaction:", err)
		errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		server.ErrHandler(w, r, &errData)
		return
	}

	// Create session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken.String(),
		Expires:  time.Now().Add(1 * time.Hour),
		HttpOnly: true,
	})

	// Insert or update session record
	_, err = db.Exec("INSERT OR REPLACE INTO session (sessionid, userid, endtime) VALUES (?, ?, ?)",
		sessionToken.String(), userID, time.Now().Add(1*time.Hour))
	if err != nil {
		log.Println("Error creating session:", err)
		// Non-fatal error, continue
	}

	// Redirect to home page
	http.Redirect(w, r, "/home?tab=posts&filter=all", http.StatusSeeOther)
}
func Emailexists(email string, w http.ResponseWriter, r *http.Request) (bool, error) {
	var emailExists bool

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Database connection failed")
		errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		server.ErrHandler(w, r, &errData)
		return false, err
	}

	defer db.Close()
	if r.Method == "POST" {

		email := r.FormValue("email")

		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM user WHERE email = ?)", email).Scan(&emailExists)
		if err != nil {
			log.Println("Failed to check if email exists")
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return false, err
		}

		if emailExists {
			err := templates.ExecuteTemplate(w, "signup.html", map[string]string{
				"ErrorMessage": "Email already exists",
			})
			if err != nil {
				log.Println("Error rendering signup page:", err)
				errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				server.ErrHandler(w, r, &errData)
			}
			return true, err
		}
	}
	return false, nil
}

func Login(InputEmail, inputPassword, provider string, w http.ResponseWriter, r *http.Request) (int, uuid.UUID, error) {
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Database connection failed")
		errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		server.ErrHandler(nil, nil, &errData)
		return 0, uuid.Nil, err
	}
	defer db.Close()

	statement, err := db.Prepare("SELECT userid, F_name, L_name, username, email, password, current_session, provider FROM user WHERE email = ? AND provider = ?")
	if err != nil {
		return 0, uuid.Nil, err
	}
	defer statement.Close()

	var id int
	var firstName string
	var lastName string
	var username string
	var email string
	var password string
	var sessionID sql.NullString
	var Provider string

	row := statement.QueryRow(InputEmail, provider)
	err = row.Scan(&id, &firstName, &lastName, &username, &email, &password, &sessionID, &Provider)
	if err != nil {
		log.Println("Login scan error:", err)
		return 0, uuid.Nil, err
	}

	// For OAuth logins, skip password verification or use a different approach
	if provider == "Github" || provider == "Google" {
		// OAuth login - we stored a shortened token or "ggg", not a real password
	} else {
		if !server.VerifyPassword(inputPassword, password) {
			return 0, uuid.Nil, fmt.Errorf("invalid credentials")
		}
	}

	// Check for existing session
	cook, valid := r.Cookie("session_id")
	if valid == nil && cook.Value != "" {
		return 0, uuid.Nil, fmt.Errorf("user already has an active session")
	}

	newSessionID, _ := security.GenerateToken()

	_, err = db.Exec("UPDATE user SET current_session = ? WHERE userid = ?", newSessionID, id)
	if err != nil {
		return 0, uuid.Nil, err
	}

	return id, newSessionID, nil
}
