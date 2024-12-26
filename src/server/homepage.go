package server

import (
	"database/sql"
	"fmt"
	"forum/database"
	"log"
	"net/http"
	"strconv"
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

	// Redirect to /home?tab=posts&filter=all if no tab is specified
	if r.URL.Query().Get("tab") == "" {
		userID := r.URL.Query().Get("user")
		log.Println("Redirecting to Home page with tab=posts&filter=all")
		http.Redirect(w, r, fmt.Sprintf("/home?user=%s&tab=posts&filter=all", userID), http.StatusFound)
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

	categories, err := database.GetAllCategories(db)
	if err != nil {
		log.Println("Failed to fetch categories")
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
		log.Println("Failed to fetch posts")
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}

	users, err := database.GetAllUsers(db)
	if err != nil {
		log.Println("Failed to fetch users")
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}

	userID := r.URL.Query().Get("user")
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		log.Println("Failed to parse user ID")
		err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
		errHandler(w, r, &err)
		return
	}

	var userName string
	var avatar sql.NullString
	var roleID int
	err = db.QueryRow("SELECT username, avatar, role_id FROM user WHERE userid = ?", userID).Scan(&userName, &avatar, &roleID)
	if err != nil {
		log.Println("Failed to fetch user data")
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}

	notifications, err := database.GetLastNotifications(db, strconv.Itoa(userIDInt))
	if err != nil {
		log.Println("Failed to fetch notifications")
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}

	data := PageData{
		UserID:         userID,
		UserName:       userName,
		Avatar:         avatar.String,
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
}
