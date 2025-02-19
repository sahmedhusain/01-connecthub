package server

import (
	"database/sql"
	"fmt"
	"forum/database"
	"log"
	"net/http"
	"strconv"
	"time"
)

func AdminPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/admin" {
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


	//check if user is an admin
	var roleID int
	err = db.QueryRow("SELECT role_id FROM user WHERE userid = ?", userID).Scan(&roleID)
	if err == sql.ErrNoRows {
		log.Println("No user found with the given user ID:", userID)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	} else if err != nil || roleID != 1 {
		log.Println("User is not an admin or error occurred")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
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

	var roleName string
	if roleID == 1 {
		roleName = "Admin"
	} else if roleID == 2 {
		roleName = "Moderator"
	} else {
		roleName = "User"
	}

	switch r.Method {
	case "GET":
		users, err := database.GetAllUsers(db)
		if err != nil {
			log.Println("Failed to fetch users")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &errData)
			return
		}

		posts, err := database.GetAllPosts(db)
		if err != nil {
			log.Println("Failed to fetch posts")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &errData)
			return
		}

		categories, err := database.GetAllCategories(db)
		if err != nil {
			log.Println("Failed to fetch categories")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &errData)
			return
		}

		reports, err := database.GetAllReports(db)
		if err != nil {
			log.Println("Failed to fetch reports")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &errData)
			return
		}

		totalUsers, err := database.GetTotalUsersCount(db)
		if err != nil {
			log.Println("Failed to fetch total users count")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &errData)
			return
		}

		totalPosts, err := database.GetTotalPostsCount(db)
		if err != nil {
			log.Println("Failed to fetch total posts count")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &errData)
			return
		}

		totalCategories, err := database.GetTotalCategoriesCount(db)
		if err != nil {
			log.Println("Failed to fetch total categories count")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &errData)
			return
		}

		notifications, err := database.GetLastNotifications(db, userID)
		if err != nil {
			log.Println("Failed to fetch notifications:", err)
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &errData)
			return
		}

		var userLogs []database.UserLog
		var userSessions []database.UserSession
		if userID := r.URL.Query().Get("user_logs"); userID != "" {
			userIDInt, err := strconv.Atoi(userID)
			if err == nil {
				userLogs, err = database.GetUserLogs(db, userIDInt)
				if err != nil {
					log.Println("Failed to fetch user logs")
					errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
					ErrHandler(w, r, &errData)
					return
				}
			}
		}

		if userID := r.URL.Query().Get("user_sessions"); userID != "" {
			userIDInt, err := strconv.Atoi(userID)
			if err == nil {
				userSessions, err = database.GetUserSessions(db, userIDInt)
				if err != nil {
					log.Println("Failed to fetch user sessions:", err)
					errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
					ErrHandler(w, r, &errData)
					return
				}
			}
		}

		var totalLikes int
		err = db.QueryRow("SELECT COUNT(*) FROM likes WHERE userid = ?", userID).Scan(&totalLikes)
		if err != nil {
			log.Println("Failed to fetch total likes:", err)
			totalLikes = 0
		}

		data := PageData{
			HasSession:     hasSession,
			UserID:   userID,
			UserName: userName,
			RoleID: 	   roleID,
			RoleName:        roleName,
			Users:           users,
			Posts:           posts,
			Categories:      categories,
			Reports:         reports,
			TotalUsers:      totalUsers,
			TotalPosts:      totalPosts,
			TotalCategories: totalCategories,
			UserLogs:        userLogs,
			UserSessions:    userSessions,
			Notifications:   notifications,
			TotalLikes:      totalLikes,
			SelectedTab:     "admin", // Set the default selected tab
		}

		err = templates.ExecuteTemplate(w, "admin.html", data)
		if err != nil {
			log.Println("Error rendering admin page:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}
	case "POST":
		r.ParseForm()
		if r.FormValue("delete_user") != "" {
			userID := r.FormValue("delete_user")
			_, err := db.Exec("DELETE FROM user WHERE id = ?", userID)
			if err != nil {
				log.Println("Failed to delete user:", err)
				errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &errData)
				return
			}
		} else if r.FormValue("delete_post") != "" {
			postID := r.FormValue("delete_post")
			_, err := db.Exec("DELETE FROM post WHERE postid = ?", postID)
			if err != nil {
				log.Println("Failed to delete post")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}
		} else if r.FormValue("delete_category") != "" {
			categoryID := r.FormValue("delete_category")
			_, err := db.Exec("DELETE FROM categories WHERE idcategories = ?", categoryID)
			if err != nil {
				log.Println("Failed to delete category")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}
		} else if r.FormValue("add_category") != "" {
			categoryName := r.FormValue("new_category")
			_, err := db.Exec("INSERT INTO categories (name) VALUES (?)", categoryName)
			if err != nil {
				log.Println("Failed to add category")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}
		} else if r.FormValue("resolve_report") != "" {
			reportID := r.FormValue("resolve_report")
			_, err := db.Exec("DELETE FROM reports WHERE id = ?", reportID)
			if err != nil {
				log.Println("Failed to resolve report")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}
		} else if r.FormValue("delete_comment") != "" {
			commentID := r.FormValue("delete_comment")
			_, err := db.Exec("DELETE FROM comment WHERE commentid = ?", commentID)
			if err != nil {
				log.Println("Failed to delete comment")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}
		} else {
			for key, values := range r.Form {
				if len(values) > 0 && key[:5] == "role_" {
					userID := key[5:]
					roleID := values[0]
					_, err := db.Exec("UPDATE user SET role_id = ? WHERE id = ?", roleID, userID)
					if err != nil {
						log.Println("Failed to update user role:", err)
						err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
						ErrHandler(w, r, &err)
						return
					}
				}
			}
		}
		log.Println("Admin action completed")
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	default:
		log.Println("Method not allowed")
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		ErrHandler(w, r, &err)
		return
	}
}
}