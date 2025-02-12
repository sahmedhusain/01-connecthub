package Handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"forum/src/security"
	"forum/src/server"
	html "html/template"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

var (
	oauthConfigGithub = &oauth2.Config{
		ClientID:     "Ov23lijcOVao9JkId97d",
		ClientSecret: "45bcae291c72da86dbdd8b65129c950f6bbf773a",
		RedirectURL:  "http://localhost:8080/auth/github/callback",
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
	username := strings.Split(primaryEmail, "@")[0]
	hashed, err := server.HashPassword(token.AccessToken)
	if err != nil {
		http.Error(w, "Error hashing password: "+err.Error(), http.StatusInternalServerError)
		return
	}

	proGithub := false
	exists, _ := Emailexists(primaryEmail, w, r)

	if exists {
		db, err := sql.Open("sqlite3", "./database/main.db")
		if err != nil {
			log.Println("Database connection failed")
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}
		defer db.Close()


		provid, err := db.Query("SELECT provider FROM user WHERE email = ?", primaryEmail)
		if err != nil {
			log.Println("Error executing query:", err)
			return
		}
		defer provid.Close()

		for provid.Next() {
			var pro string
			err := provid.Scan(&pro)
			if err != nil {
				log.Println("Error scanning provider:", err)
				return
			}
			if pro == "GitHub" {
				proGithub = true
			}
		}
	}

	if !proGithub {

		db, err := sql.Open("sqlite3", "./database/main.db")
		if err != nil {
			log.Println("Database connection failed")
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}
		defer db.Close()

		defaultAvatar := "static/assets/default-avatar.png"

		stmt, err := db.Prepare("INSERT INTO user (F_name, L_name, Username, Email, password, current_session, role_id, Avatar, provider) VALUES (?, ?, ?, ?, ?, ?, ?, ?,?)")
		if err != nil {
			log.Println("Failed to prepare insert statement:", err)
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec("Github", "User", username, primaryEmail, hashed, "", 3, defaultAvatar, "Github")
		if err != nil {
			log.Println("Failed to insert user data:", err)
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}

		githubID, ok := user["id"].(float64)
		if !ok {
			http.Error(w, "GitHub ID is missing or invalid", http.StatusInternalServerError)
			return
		}
		stmtt, err := db.Prepare("INSERT INTO google (gituserid, gitF_name, gitL_name, gitUsername, gitEmail, gitpassword, gitAvatar, user_userid) VALUES (?, ?, ?, ?, ?, ?, ?,?)")
		if err != nil {
			log.Println("Failed to prepare insert statement:", err)
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}
		defer stmtt.Close()

		var userid int
		err = db.QueryRow("SELECT userid FROM user WHERE email = ?", primaryEmail).Scan(&userid)
		if err != nil {
			log.Println("Error executing query:", err)
			return
		}

		_, err = stmtt.Exec(githubID, "", "", username, primaryEmail, hashed, defaultAvatar, userid)
		if err != nil {
			log.Println("Failed to insert user data:", err)
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}
	}
	sid, _, err := Login(primaryEmail, token.AccessToken, "GitHub", w, r)
	if err != nil {
		http.Error(w, "Login failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	server.CreateSession(w, r, sid)
	http.Redirect(w, r, "/", http.StatusSeeOther)

}
var (
	oauthConfigGoogle = &oauth2.Config{
		ClientID:     "45bcae291c72da86dbdd8b65129c950f6bbf773a270576066421-puugu8n2v7om91no9u1kq116l0uf345e.apps.googleusercontent.com",
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
	firstName, _ := user["given_name"].(string)
	lastName, _ := user["family_name"].(string)
	profilePicture, _ := user["picture"].(string)
	username := strings.Split(email, "@")[0]
	hashed, err := server.HashPassword(token.AccessToken)
	if err != nil {
		http.Error(w, "Error hashing password: "+err.Error(), http.StatusInternalServerError)
		return
	}
	proGoogle := false
	exists, _ := Emailexists(email, w, r)
	if exists {
		db, err := sql.Open("sqlite3", "./database/main.db")
		if err != nil {
			log.Println("Database connection failed")
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}
		defer db.Close()

		provid, err := db.Query("SELECT provider FROM user WHERE email = ?", email)
		if err != nil {
			log.Println("Error executing query:", err)
			return
		}
		defer provid.Close()

		for provid.Next() {
			var pro string
			err := provid.Scan(&pro)
			if err != nil {
				log.Println("Error scanning provider:", err)
				return
			}
			if pro == "Google" {
				proGoogle = true
			}
		}
	}
	if !proGoogle {

		db, err := sql.Open("sqlite3", "./database/main.db")
		if err != nil {
			log.Println("Database connection failed")
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}
		defer db.Close()

		stmt, err := db.Prepare("INSERT INTO user (F_name, L_name, Username, Email, password, current_session, role_id, Avatar, provider) VALUES (?, ?, ?, ?, ?, ?, ?, ?,?)")
		if err != nil {
			log.Println("Failed to prepare insert statement:", err)
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(firstName, lastName, username, email, hashed, "", 3, profilePicture, "Google")
		if err != nil {
			log.Println("Failed to insert user data:", err)
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}

		var userid int
		err = db.QueryRow("SELECT userid FROM user WHERE email = ?", email).Scan(&userid)
		if err != nil {
			log.Println("Error executing query:", err)
			return
		}

		stmtt, err := db.Prepare("INSERT INTO google (googleuserid, googleF_name, googleL_name, googleUsername, googleEmail, googlepassword, googleAvatar, user_userid) VALUES (?, ?, ?, ?, ?, ?, ?,?)")
		if err != nil {
			log.Println("Failed to prepare insert statement:", err)
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}
		defer stmtt.Close()

		_, err = stmtt.Exec(googleID, firstName, lastName, username, email, hashed, profilePicture, userid)
		if err != nil {
			log.Println("Failed to insert user data:", err)
			errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.ErrHandler(w, r, &errData)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Log the user in (set session or cookie)
	sid, _, err := Login(email, token.AccessToken, "Google", w, r)
	if err != nil {
		errData := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		server.ErrHandler(w, r, &errData)
		return
	}
	server.CreateSession(w, r, sid)
	http.Redirect(w, r, "/", http.StatusSeeOther)
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

	statement, err := db.Prepare("SELECT id, F_name, L_name, username, email, password, current_session ,provider FROM user WHERE email = ? AND provider = ?")
	if err != nil {
		return 0, uuid.Nil, err
	}
	defer statement.Close()

	var id int
	var email string
	var username string
	var password string
	var Provider string
	var sessionID sql.NullString

	row := statement.QueryRow(InputEmail, provider)
	err = row.Scan(&id, &email, &username, &password, &sessionID, &Provider)
	if err != nil {
		return 0, uuid.Nil, err
	}

	if !server.VerifyPassword(inputPassword, password) {
		return 0, uuid.Nil, fmt.Errorf("invalid credentials")
	}

	cook, valid := r.Cookie("session_id")

	// Check if the user already has an active session
	if valid != nil || cook.Value != "" {
		return 0, uuid.Nil, fmt.Errorf("user already has an active session")
	}

	// Generate a new session ID (you can use a library like UUID for this)
	newSessionID, _ := security.GenerateToken()

	// Update the database with the new session ID
	_, err = db.Exec("UPDATE user SET current_session = ? WHERE id = ?", newSessionID, id, Provider)
	if err != nil {
		return 0, uuid.Nil, err
	}
	Provider = ""

	return id, newSessionID, nil
}
