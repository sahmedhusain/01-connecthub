package server

import (
	"database/sql"
	"fmt"
	"forum/database"
	"log"
	"net/http"
)

func HomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/home" {
		log.Println("Redirecting to Home page")
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		errHandler(w, r, &err)
		return
	}

	if r.Method != "GET" {
		log.Println("Method not allowed")
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		errHandler(w, r, &err)
		return
	}

	var hasSession bool
	seshCok, err := r.Cookie("session_token")
	if err != nil || seshCok.Value == "" {
		hasSession = false
	} else {
		hasSession = true
	}

	// Redirect to /home?tab=posts&filter=all if no tab is specified
	// if r.URL.Query().Get("tab") == "" {
	// 	userID := r.URL.Query().Get("user")
	// 	if userID == "" {
	// 		log.Println("User ID is missing")
	// 		http.Redirect(w, r, "/", http.StatusSeeOther)
	// 		return
	// 	}
	// 	log.Println("Redirecting to Home page with tab=posts&filter=all")
	// 	http.Redirect(w, r, fmt.Sprintf("/home?user=%s&tab=posts&filter=all", userID), http.StatusFound)
	// 	return
	// }

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Database connection failed:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}
	defer db.Close()

	categories, err := database.GetAllCategories(db)
	if err != nil {
		log.Println("Failed to fetch categories:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}

	var posts []database.Post
	filter := r.URL.Query().Get("filter")
	selectedTab := r.URL.Query().Get("tab")

	if selectedTab == "" {
		selectedTab = "posts"
	}

	if filter == "" {
		if selectedTab == "your+posts" {
			filter = "newest"
		} else if selectedTab == "your+replies" {
			filter = "newest"
		} else if selectedTab == "your+reactions" {
			filter = "likes"
		} else {
			filter = "all"
		}
	}

	if selectedTab == "tags" && filter != "all" {
		posts, err = database.GetPostsByCategory(db, filter)
	} else if filter == "all" {
		posts, err = database.GetAllPosts(db)
	} else {
		posts, err = database.GetFilteredPosts(db, filter)
	}
	if err != nil {
		log.Println("Failed to fetch posts:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}

	users, err := database.GetAllUsers(db)
	if err != nil {
		log.Println("Failed to fetch users:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}

	if hasSession {
		// userID := r.URL.Query().Get("user")
		// if userID == "" {
		// 	log.Println("User ID is missing")
		// 	http.Redirect(w, r, "/", http.StatusSeeOther)
		// 	return
		// }

		// userIDInt, err := strconv.Atoi(userID)
		// if err != nil {
		// 	log.Println("Failed to parse user ID:", err)
		// 	err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
		// 	errHandler(w, r, &err)
		// 	return
		// }

		// Retrieve username cookie
		usrCok, err := r.Cookie("dotcom_user")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			fmt.Println("Error fetching username from cookie")
			return
		}

		//Set username from cookie value
		userName := usrCok.Value

		var userID string
		err = db.QueryRow("SELECT userid FROM user WHERE Username = ?", userName).Scan(&userID)
		if err != nil {
			log.Println("Error fetching session ID from user table:", err)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		var avatar sql.NullString
		var roleID int
		err = db.QueryRow("SELECT avatar, role_id FROM user WHERE userid = ?", userID).Scan(&avatar, &roleID)
		if err == sql.ErrNoRows {
			log.Println("No user found with the given ID:", userID)
			err := ErrorPageData{Code: "404", ErrorMsg: "USER NOT FOUND"}
			errHandler(w, r, &err)
			return
		} else if err != nil {
			log.Println("Failed to fetch user data:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
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

		var totalLikes, totalPosts int
		err = db.QueryRow("SELECT COUNT(*) FROM likes WHERE user_userid = ?", userID).Scan(&totalLikes)
		if err != nil {
			log.Println("Failed to fetch total likes:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		err = db.QueryRow("SELECT COUNT(*) FROM post WHERE user_userid = ?", userID).Scan(&totalPosts)
		if err != nil {
			log.Println("Failed to fetch total posts:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		notifications, err := database.GetLastNotifications(db, userID)
		if err != nil {
			log.Println("Failed to fetch notifications:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		data := PageData{
			HasSession:     hasSession,
			UserID:         userID,
			UserName:       userName,
			Avatar:         avatar.String,
			RoleName:       roleName,
			TotalLikes:     totalLikes,
			TotalPosts:     totalPosts,
			Categories:     categories,
			Users:          users,
			Posts:          posts,
			SelectedTab:    selectedTab,
			SelectedFilter: filter,
			Notifications:  notifications,
			RoleID:         roleID,
		}

		err = templates.ExecuteTemplate(w, "home.html", data)
		if err != nil {
			log.Println("Error rendering home page:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}
	} else {
		data := PageData{
			HasSession:     hasSession,
			Categories:     categories,
			Users:          users,
			Posts:          posts,
			SelectedTab:    selectedTab,
			SelectedFilter: filter,
		}

		err = templates.ExecuteTemplate(w, "home.html", data)
		if err != nil {
			log.Println("Error rendering home page:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}
	}
}
