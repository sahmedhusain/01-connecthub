package server

import (
	"database/sql"
	"forum/database"
	"log"
	"net/http"
)

func NotificationsPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/notifications" {
		log.Println("Redirecting to Home page")
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

	session, _ := store.Get(r, "session-name")
	userID, ok := session.Values["userID"].(string)
	if !ok || userID == "" {
		log.Println("UserID not found in session, redirecting to login page")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
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
		UserID        string
		Avatar        string
		Notifications []database.Notification
	}{
		UserID:        userID,
		Avatar:        session.Values["avatar"].(string),
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
