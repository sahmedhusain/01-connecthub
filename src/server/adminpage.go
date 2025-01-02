package server

import (
	"database/sql"
	"forum/database"
	"log"
	"net/http"
	"strconv"
)

func AdminPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/admin" {
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

	// Retrieve UserID from session
	session, _ := store.Get(r, "session-name")
	userID, ok := session.Values["userID"].(string)
	if !ok || userID == "" {
		log.Println("UserID not found in session, redirecting to login page")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Retrieve UserName from session
	userName, ok := session.Values["username"].(string)
	if !ok || userName == "" {
		log.Println("UserName not found in session, redirecting to login page")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check if the user is an admin
	var roleID int
	err = db.QueryRow("SELECT role_id FROM user WHERE userid = ?", userID).Scan(&roleID)
	if (err == sql.ErrNoRows) {
		log.Println("No user found with the given ID:", userID)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	} else if err != nil || roleID != 1 {
		log.Println("User is not an admin or error occurred")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Retrieve avatar from session
	avatar, ok := session.Values["avatar"].(string)
	if (!ok) {
		// Set a default avatar if not found in session
		avatar = "/static/assets/default-avatar.png"
	}

	// Determine role name
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
			errHandler(w, r, &errData)
			return
		}

		posts, err := database.GetAllPosts(db)
		if err != nil {
			log.Println("Failed to fetch posts")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}

		categories, err := database.GetAllCategories(db)
		if err != nil {
			log.Println("Failed to fetch categories")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}

		reports, err := database.GetAllReports(db)
		if err != nil {
			log.Println("Failed to fetch reports")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}

		totalUsers, err := database.GetTotalUsersCount(db)
		if err != nil {
			log.Println("Failed to fetch total users count")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}

		totalPosts, err := database.GetTotalPostsCount(db)
		if err != nil {
			log.Println("Failed to fetch total posts count")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}

		totalCategories, err := database.GetTotalCategoriesCount(db)
		if err != nil {
			log.Println("Failed to fetch total categories count")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}

		notifications, err := database.GetLastNotifications(db, userID)
		if err != nil {
			log.Println("Failed to fetch notifications:", err)
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
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
					errHandler(w, r, &errData)
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
					errHandler(w, r, &errData)
					return
				}
			}
		}

		// Retrieve total likes for the user
		var totalLikes int
		err = db.QueryRow("SELECT COUNT(*) FROM likes WHERE userid = ?", userID).Scan(&totalLikes)
		if err != nil {
			log.Println("Failed to fetch total likes:", err)
			totalLikes = 0
		}

		data := PageData{
			UserID:          userID,
			UserName:        userName,
			Avatar:          avatar,
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
			SelectedTab:    "admin", // Set the default selected tab
		}

		err = templates.ExecuteTemplate(w, "admin.html", data)
		if err != nil {
			log.Println("Error rendering admin page:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
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
				errHandler(w, r, &errData)
				return
			}
		} else if r.FormValue("delete_post") != "" {
			postID := r.FormValue("delete_post")
			_, err := db.Exec("DELETE FROM post WHERE postid = ?", postID)
			if err != nil {
				log.Println("Failed to delete post")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
		} else if r.FormValue("delete_category") != "" {
			categoryID := r.FormValue("delete_category")
			_, err := db.Exec("DELETE FROM categories WHERE idcategories = ?", categoryID)
			if err != nil {
				log.Println("Failed to delete category")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
		} else if r.FormValue("add_category") != "" {
			categoryName := r.FormValue("new_category")
			_, err := db.Exec("INSERT INTO categories (name) VALUES (?)", categoryName)
			if err != nil {
				log.Println("Failed to add category")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
		} else if r.FormValue("resolve_report") != "" {
			reportID := r.FormValue("resolve_report")
			_, err := db.Exec("DELETE FROM reports WHERE id = ?", reportID)
			if err != nil {
				log.Println("Failed to resolve report")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
		} else if r.FormValue("delete_comment") != "" {
			commentID := r.FormValue("delete_comment")
			_, err := db.Exec("DELETE FROM comment WHERE commentid = ?", commentID)
			if err != nil {
				log.Println("Failed to delete comment")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
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
						errHandler(w, r, &err)
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
		errHandler(w, r, &err)
		return
	}
}
