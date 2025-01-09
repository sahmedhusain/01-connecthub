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
	err = session.Save(r, w)
	if err != nil {
		log.Println("Error saving session:", err)
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &errData)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
