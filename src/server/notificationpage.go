package server

import (
	"database/sql"
	"forum/database"
	"log"
	"net/http"
	"strconv"
)

func NotificationsPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/notifications" {
		log.Println("Invalid URL path")
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		errHandler(w, r, &err)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Database connection failed")
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}
	defer db.Close()

	userID, err := strconv.Atoi(r.FormValue("user"))
	if err != nil {
		log.Println("Error converting userID to int:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}

	notifications, err := database.GetLastNotifications(db, userID)
	if err != nil {
		log.Println("Failed to fetch notifications")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &errData)
		return
	}

	data := struct {
		UserID        int
		Avatar        string
		Notifications []database.Notification
	}{
		UserID: userID,
		// Avatar:        session.Values["avatar"].(string),
		Notifications: notifications,
	}

	err = templates.ExecuteTemplate(w, "notifications.html", data)
	if err != nil {
		log.Println("Error rendering notifications page:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}
}
