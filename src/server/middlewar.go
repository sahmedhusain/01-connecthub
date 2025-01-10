package server

import (
	"log"
	"net/http"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			return
		}

		next.ServeHTTP(w, r)
	})
}
