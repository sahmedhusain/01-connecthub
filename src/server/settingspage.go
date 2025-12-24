package server

import (
	"01connecthub/database"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func SettingsPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/settings" {
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
		var user database.User
		err := db.QueryRow("SELECT F_name, L_name, username, email, avatar FROM user WHERE userid = ?", userID).Scan(&user.FirstName, &user.LastName, &user.Username, &user.Email, &user.Avatar)
		if err != nil {
			log.Println(err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}
		var totalLikes, totalPosts int
		err = db.QueryRow("SELECT COUNT(*) FROM likes WHERE user_userID = ?", userID).Scan(&totalLikes)
		if err != nil {
			log.Println(err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}

		err = db.QueryRow("SELECT COUNT(*) FROM post WHERE user_userID = ?", userID).Scan(&totalPosts)
		if err != nil {
			log.Println(err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}

		switch r.Method {
		case "GET":

			var user database.User
			err := db.QueryRow("SELECT F_name, L_name, username, email, avatar FROM user WHERE userid = ?", userID).Scan(&user.FirstName, &user.LastName, &user.Username, &user.Email, &user.Avatar)
			if err != nil {
				log.Println("Failed to fetch user data:", err)
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
				RoleName      string
				UserID        int
				FirstName     string
				LastName      string
				UserName      string
				Email         string
				Avatar        string
				Password      string
				PasswordShown bool
				Notifications []database.Notification
				TotalLikes    int
				TotalPosts    int
				SelectedTab   string
				RoleID        int
			}{
				HasSession:    hasSession,
				RoleName:      roleName,
				UserID:        userID,
				FirstName:     user.FirstName,
				LastName:      user.LastName,
				UserName:      user.Username,
				Email:         user.Email,
				Avatar:        user.Avatar.String,
				Password:      "",
				Notifications: []database.Notification{},
				TotalLikes:    totalLikes,
				TotalPosts:    totalPosts,
				SelectedTab:   "settings",
				RoleID:        roleID,
			}

			err = templates.ExecuteTemplate(w, "settings.html", data)
			if err != nil {
				log.Println("Error rendering settings page:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}
		case "POST":
			err := r.ParseMultipartForm(10 << 20)
			if err != nil {
				log.Println("Failed to parse form data")
				err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
				ErrHandler(w, r, &err)
				return
			}

			firstName := r.FormValue("first_name")
			lastName := r.FormValue("last_name")
			username := r.FormValue("username")
			email := r.FormValue("email")
			password := r.FormValue("password")

			var avatarPath sql.NullString
			file, handler, err := r.FormFile("avatar")
			if err == nil {
				defer file.Close()
				avatarPath.String = fmt.Sprintf("static/uploads/%s", handler.Filename)
				avatarPath.Valid = true

				os.MkdirAll("static/uploads", os.ModePerm)

				f, err := os.OpenFile(avatarPath.String, os.O_WRONLY|os.O_CREATE, 0666)
				if err != nil {
					log.Println("Failed to open file for writing")
					err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
					ErrHandler(w, r, &err)
					return
				}
				defer f.Close()
				io.Copy(f, file)
			} else if err != http.ErrMissingFile {
				log.Println("Failed to upload avatar")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}

			if password != "" {
				_, err = db.Exec("UPDATE user SET first_name = ?, last_name = ?, username = ?, email = ?, password = ?, avatar = ? WHERE id = ?", firstName, lastName, username, email, password, avatarPath, userID)
			} else {
				_, err = db.Exec("UPDATE user SET first_name = ?, last_name = ?, username = ?, email = ?, avatar = ? WHERE id = ?", firstName, lastName, username, email, avatarPath, userID)
			}

			if err != nil {
				log.Println("Failed to update user data")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}
			log.Println("User data updated with ID:", userID)
			http.Redirect(w, r, "/settings", http.StatusSeeOther)
		default:
			log.Println("Method not allowed")
			err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
			ErrHandler(w, r, &err)
			return
		}
	}

}
