package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
)

func ReverseMiddleware(next http.HandlerFunc) http.HandlerFunc {
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
			fmt.Println("This user has no cookie")
		} else {
			//Set username from cookie value
			userName := usrCok.Value

			var sessionID string
			err = db.QueryRow("SELECT current_session FROM user WHERE Username = ?", userName).Scan(&sessionID)
			if err != nil {
				log.Println("Error fetching session ID from user table:", err)
				return
			}

			seshCok, _ := r.Cookie("session_token")
			if seshCok.Value == sessionID {
				fmt.Println("Valid cookie, redirected to home")
				http.Redirect(w, r, "/home", http.StatusFound)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
