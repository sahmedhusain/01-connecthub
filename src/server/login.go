package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"
)

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		log.Println("Redirecting to Home page")
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		errHandler(w, r, &err)
		return
	}

	if r.Method == "POST" {
		email := r.FormValue("email")
		password := r.FormValue("password")

		db, err := sql.Open("sqlite3", "./database/main.db")
		if err != nil {
			log.Println("Database connection failed")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}
		defer db.Close()

		var userID int
		var dbPassword, userName string
		err = db.QueryRow("SELECT userid, password, username FROM user WHERE email = ?", email).Scan(&userID, &dbPassword, &userName)
		if err != nil {
			if err == sql.ErrNoRows {
				// No user found with the given email
				err = templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
					"ErrorMsg": "Invalid email or password",
				})
				if err != nil {
					log.Println("Error rendering login page:", err)
					errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
					errHandler(w, r, &errData)
				}
				return
			}
			log.Println("Failed to fetch user data")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		// Check if the password is correct
		if !VerifyPassword(password, dbPassword) {
			err = templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
				"ErrorMsg": "Invalid email or password",
			})
			if err != nil {
				log.Println("Error rendering login page:", err)
				errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &errData)
			}
			return
		}

		//End session when user already has a session
		result, err := db.Exec("DELETE FROM session WHERE userid = ?", userID)
		if err != nil {
			log.Println("Error rendering login page:", err)
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
		} else if rowsAffected, err := result.RowsAffected(); err == nil && rowsAffected == 1 {
			session, _ := store.Get(r, "session")
			delete(session.Values, "userID")
			delete(session.Values, "sessionID")
			err = session.Save(r, w)
			if err != nil {
				log.Println("Error saving session:", err)
				errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &errData)
				return
			}
		}

		// Create a new session in the database
		var sessionID int
		startTime := time.Now()
		err = db.QueryRow(
			"INSERT INTO session (userid, start) VALUES (?, ?) RETURNING sessionid",
			userID, startTime).Scan(&sessionID)
		if err != nil {
			log.Println("Error creating session:", err)
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}

		//create a new session associated with the user
		session, _ := store.Get(r, "session")
		session.Values["userID"] = userID
		session.Values["sessionID"] = sessionID

		err = session.Save(r, w)
		if err != nil {
			log.Println("Error saving session:", err)
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}

		// Update the user's session ID in the database
		_, err = db.Exec("UPDATE user SET session_sessionid = ? WHERE userid = ?", sessionID, userID)
		if err != nil {
			log.Println("Error updating user session ID:", err)
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}

		log.Println("User logged in with userID:", userID)

		// If login is successful, redirect to the Home page with user ID
		log.Println("Redirecting to Home page with user ID")
		http.Redirect(w, r, fmt.Sprintf("/home?user=%d&tab=posts&filter=all", userID), http.StatusSeeOther)
		return
	}

	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Println("Error rendering login page:", err)
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &errData)
	}
}
