package server

import (
	"database/sql"
	"forum/database"
	"log"
	"net/http"
)

func MainPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		log.Println("Invalid URL path")
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

	if r.URL.Query().Get("tab") == "" {
		log.Println("Redirecting to Home page with tab=posts&filter=all")
		http.Redirect(w, r, "/?tab=posts&filter=all", http.StatusFound)
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
	if filter == "" {
		filter = "all"
	}
	selectedTab := r.URL.Query().Get("tab")
	if selectedTab == "" {
		selectedTab = "posts"
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

	data := PageData{
		Categories:     categories,
		Users:          users,
		Posts:          posts,
		SelectedTab:    selectedTab,
		SelectedFilter: filter,
	}

	err = templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		log.Println("Error rendering index page:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}
}
