package server

import (
	"database/sql"
	"forum/src/security"
	"log"
	"net/http"
	"regexp"
	"time"
)

func SignupPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/signup" {
		log.Println("Invalid URL path")
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		ErrHandler(w, r, &err)
		return
	}

	if r.Method == "POST" {
		F_name := r.FormValue("first_name")
		L_name := r.FormValue("last_name")
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm-password")

		emailRegex := regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)
		if !emailRegex.MatchString(email) {
			err := templates.ExecuteTemplate(w, "signup.html", map[string]string{
				"ErrorMessage": "Invalid email format",
			})
			if err != nil {
				log.Println("Error rendering signup page:", err)
				errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &errData)
			}
			return
		}

		if password != confirmPassword {
			err := templates.ExecuteTemplate(w, "signup.html", map[string]string{
				"ErrorMessage": "Passwords do not match",
			})
			if err != nil {
				log.Println("Error rendering signup page:", err)
				errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &errData)
			}
			return
		}

		db, err := sql.Open("sqlite3", "./database/main.db")
		if err != nil {
			log.Println("Database connection failed")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &errData)
			return
		}
		defer db.Close()

		var usernameExists, emailExists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM user WHERE username = ?)", username).Scan(&usernameExists)
		if err != nil {
			log.Println("Failed to check if username exists")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &errData)
			return
		}

		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM user WHERE email = ?)", email).Scan(&emailExists)
		if err != nil {
			log.Println("Failed to check if email exists")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &errData)
			return
		}

		if usernameExists {
			err := templates.ExecuteTemplate(w, "signup.html", map[string]string{
				"ErrorMessage": "Username already exists",
			})
			if err != nil {
				log.Println("Error rendering signup page:", err)
				errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &errData)
			}
			return
		}

		if emailExists {
			err := templates.ExecuteTemplate(w, "signup.html", map[string]string{
				"ErrorMessage": "Email already exists",
			})
			if err != nil {
				log.Println("Error rendering signup page:", err)
				errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &errData)
			}
			return
		}

		// Begin transaction for atomic operations
		tx, err := db.Begin()
		if err != nil {
			log.Println("Failed to begin transaction:", err)
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &errData)
			return
		}

		// Generate session token
		sessionToken, err := security.GenerateToken()
		if err != nil {
			log.Println("Failed to generate session token:", err)
			tx.Rollback()
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &errData)
			return
		}

		hashedPassword, _ := HashPassword(password)
		defaultAvatar := "static/assets/default-avatar.png"

		var roleID int
		if email == "sayedahmed97.sad@gmail.com" {
			roleID = 1
		} else if email == "qassimhassan9@gmail.com" {
			roleID = 2
		} else {
			roleID = 3
		}

		// Insert user with session token
		result, err := tx.Exec("INSERT INTO user (F_name, L_name, Username, Email, password, current_session, role_id, Avatar, provider) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			F_name, L_name, username, email, hashedPassword, sessionToken.String(), roleID, defaultAvatar, "normal")
		if err != nil {
			log.Println("Failed to insert user data:", err)
			tx.Rollback()
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &errData)
			return
		}

		// Get user ID for session creation
		lastID, err := result.LastInsertId()
		if err != nil {
			log.Println("Failed to get last insert ID:", err)
			tx.Rollback()
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &errData)
			return
		}

		userID := int(lastID)

		// Create session record
		_, err = tx.Exec("INSERT INTO session (sessionid, userid, endtime) VALUES (?, ?, ?)",
			sessionToken.String(), userID, time.Now().Add(1*time.Hour))
		if err != nil {
			log.Println("Error creating session:", err)
			tx.Rollback()
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &errData)
			return
		}

		// Commit transaction
		if err = tx.Commit(); err != nil {
			log.Println("Failed to commit transaction:", err)
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &errData)
			return
		}

		// Create session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    sessionToken.String(),
			Expires:  time.Now().Add(1 * time.Hour),
			HttpOnly: true,
		})

		// Redirect to home page
		http.Redirect(w, r, "/home?tab=posts&filter=all", http.StatusSeeOther)
		return
	}

	err := templates.ExecuteTemplate(w, "signup.html", nil)
	if err != nil {
		log.Println("Error rendering signup page:", err)
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &errData)
	}
}
