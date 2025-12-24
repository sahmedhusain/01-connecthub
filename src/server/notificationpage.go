package server

import (
	"01connecthub/database"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"
)

func NotificationsPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/notifications" {
		log.Println("Invalid URL path")
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		ErrHandler(w, r, &err)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Database connection failed")
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &err)
		return
	}
	defer db.Close()

	var hasSession bool
	var userID int
	var userName string
	seshCok, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println("No cookie found, treated as guest")
	} else if seshCok.Value == "" {
		hasSession = false
	} else {
		hasSession = true

		seshVal := seshCok.Value

		err = db.QueryRow("SELECT userid, Username FROM user WHERE current_session = ?", seshVal).Scan(&userID, &userName)
		if err == sql.ErrNoRows {
			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    "",
				Expires:  time.Now().Add(-time.Hour),
				HttpOnly: true,
			})
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else if err != nil {
			log.Println("Error fetching userid ID from user table:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		} else {
			log.Println("User is logged in:", userName)
		}
	}

	var roleID int
	err = db.QueryRow("SELECT role_id FROM user WHERE userid = ?", userID).Scan(&roleID)
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &err)
		return
	}

	var roleName string
	if roleID == 1 {
		roleName = "Admin"
	} else if roleID == 2 {
		roleName = "Moderator"
	} else {
		roleName = "User"
	}

	if hasSession {
		var avatar sql.NullString
		err = db.QueryRow("SELECT avatar, role_id FROM user WHERE userID = ?", userID).Scan(&avatar, &roleID)
		if err == sql.ErrNoRows {
			log.Println("No user found with the given ID:", userID)
			err := ErrorPageData{Code: "404", ErrorMsg: "USER NOT FOUND"}
			ErrHandler(w, r, &err)
			return
		} else if err != nil {
			log.Println("Failed to fetch user data:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}

		notifications, err := database.GetLastNotifications(db, userID)
		if err != nil {
			log.Println("Failed to fetch notifications")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &errData)
			return
		}

		var totalLikes, totalPosts int
		err = db.QueryRow("SELECT COUNT(*) FROM likes WHERE user_userID = ?", userID).Scan(&totalLikes)
		if err != nil {
			log.Println("Failed to fetch total likes:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}

		err = db.QueryRow("SELECT COUNT(*) FROM post WHERE user_userID = ?", userID).Scan(&totalPosts)
		if err != nil {
			log.Println("Failed to fetch total posts:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}

		data := struct {
			HasSession    bool
			UserID        int
			UserName      string
			Avatar        string
			Notifications []database.Notification
			RoleName      string
			TotalLikes    int
			TotalPosts    int
			RoleID        int
			SelectedTab   string
		}{
			HasSession:    hasSession,
			UserID:        userID,
			UserName:      userName,
			Avatar:        avatar.String,
			RoleName:      roleName,
			Notifications: notifications,
			RoleID:        roleID,
			TotalLikes:    totalLikes,
			TotalPosts:    totalPosts,
			SelectedTab:   "notifications",
		}

		err = templates.ExecuteTemplate(w, "notifications.html", data)
		if err != nil {
			log.Println("Error rendering notifications page:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}
	}
}
