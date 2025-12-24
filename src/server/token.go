package server

import (
	UUID "01connecthub/src/security"
	"database/sql"
	"log"
	"net/http"
	"time"
)

func CreateSession(w http.ResponseWriter, r *http.Request, userID int) {

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Database connection failed")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &errData)
		return
	}
	defer db.Close()
	sessionToken, err := UUID.GenerateToken()
	if err != nil {
		log.Println("Error generating UUID:", err)
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &errData)
	}

	stringToken := sessionToken.String()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    stringToken,
		Expires:  time.Now().Add(1 * time.Hour),
		HttpOnly: true,
	})

	result, err := db.Exec("UPDATE session SET sessionid = ? WHERE userid = ?", stringToken, userID)
	if err != nil {
		log.Println("Error updating session ID in session table:", err)
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &errData)
		return
	} else if rowsAffected, err := result.RowsAffected(); err == nil && rowsAffected == 0 { //only insert a new row if no record is updated (i.e., no session is found)
		_, err := db.Exec("INSERT INTO session (sessionid, userid, endtime) VALUES (?, ?, ?) RETURNING sessionid",
			stringToken, userID, time.Now().Add(1*time.Hour))
		if err != nil {
			log.Println("Error creating new session:", err)
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &errData)
			return
		}
	}

	_, err = db.Exec("UPDATE user SET current_session = ? WHERE userid = ?", stringToken, userID)
	if err != nil {
		log.Println("Error updating session ID in user table:", err)
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &errData)
		return
	}

}
