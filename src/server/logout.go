package server

import (
	"database/sql"
	"log"
	"net/http"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM session WHERE userid = ?", session.Values["userID"])
	if err != nil {
		log.Fatal(err)
	}

	delete(session.Values, "userID")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
