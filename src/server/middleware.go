package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := sql.Open("sqlite3", "./database/main.db")
		if err != nil {
			log.Println("Database connection failed")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}
		defer db.Close()

		// Retrieve username cookie
		usrCok, err := r.Cookie("dotcom_user")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			fmt.Println("Error fetching username from cookie")
			return
		}

		//Set username from cookie value
		userName := usrCok.Value

		var sessionID string
		err = db.QueryRow("SELECT current_session FROM user WHERE Username = ?", userName).Scan(&sessionID)
		if err != nil {
			log.Println("Error fetching session ID from user table:", err)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		seshCok, err := r.Cookie("session_token")
		fmt.Println("Invalid session")
		if err != nil || seshCok.Value == "" || seshCok.Value != sessionID {
			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    "",
				Expires:  time.Now().Add(-time.Hour),
				HttpOnly: true,
			})

			http.SetCookie(w, &http.Cookie{
				Name:     "dotcom_user",
				Value:    "",
				Expires:  time.Now().Add(-time.Hour),
				HttpOnly: true,
			})
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
