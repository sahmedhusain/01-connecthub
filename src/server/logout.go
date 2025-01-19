package server

import (
	"database/sql"
	"log"
	"net/http"
	"time"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	})

	_, err = db.Exec("DELETE FROM session WHERE userid = ?", userID)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("UPDATE user SET current_session = NULL WHERE userid = ?", userID)
	if err != nil {
		log.Fatal(err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
