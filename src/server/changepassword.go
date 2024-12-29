package server

import (
	"database/sql"
	"log"
	"net/http"
)

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Println("Method not allowed")
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		errHandler(w, r, &err)
		return
	}

	userID := r.FormValue("user_id")
	currentPassword := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Error opening database:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}
	defer db.Close()

	var storedPassword string
	err = db.QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&storedPassword)
	if err != nil {
		log.Println("Error fetching user:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}

	if storedPassword != currentPassword {
		log.Println("Current password is incorrect")
		err := ErrorPageData{Code: "400", ErrorMsg: "Current password is incorrect"}
		errHandler(w, r, &err)
		return
	}

	_, err = db.Exec("UPDATE users SET password = ? WHERE id = ?", newPassword, userID)
	if err != nil {
		log.Println("Error updating password:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}

	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}
