package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"
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

		// Fetch session cookie
		seshCok, err := r.Cookie("session_toekn")
		if err != nil {
			fmt.Println("This user has no cookie")
		} else {
			//Set session token from cookie value
			seshVal := seshCok.Value

			var exists bool
			err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM user WHERE current_session = ?)", seshVal).Scan(&exists)
			if err != nil {
				log.Println("Error :", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			} else if !exists {
				log.Println("Inavlid Session")
				http.SetCookie(w, &http.Cookie{
					Name:     "session_token",
					Value:    "",
					Expires:  time.Now().Add(-time.Hour),
					HttpOnly: true,
				})
				http.Redirect(w, r, "/", http.StatusBadRequest)
			} else if exists {
				fmt.Println("Valid cookie")
				http.Redirect(w, r, "/home", http.StatusUnauthorized)
			}
		}
		next.ServeHTTP(w, r)
	})
}
