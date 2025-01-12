package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
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
		session, err := store.Get(r, "session")
		if err != nil {
			log.Println("Error getting session:", err)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		_, userOk := session.Values["userID"]
		_, sessionOk := session.Values["sessionID"]
		if !userOk || !sessionOk {
			http.Redirect(w, r, "/", http.StatusFound)
			fmt.Println("no session, redirected to login")
			return
		}

		userID := session.Values["userID"]
		var sessionID int
		err = db.QueryRow("SELECT session_sessionid FROM user WHERE userid = ?", userID).Scan(&sessionID)
		if err != nil {
			log.Println("Error fetching session ID:", err)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		fmt.Print("gorilla sessionID: ", session.Values["sessionID"])
		fmt.Print("sessionID: ", sessionID)
		if session.Values["sessionID"] != sessionID {
			http.Redirect(w, r, "/", http.StatusFound)
			fmt.Println("gorilla session id does not not match db session id")
			return
		}

		next.ServeHTTP(w, r)
	})
}
