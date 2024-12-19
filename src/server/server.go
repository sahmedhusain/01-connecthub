package server

import (
	"database/sql"
	"fmt"
	"forum/database"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

var templates *template.Template

func init() {
	templates = template.Must(template.ParseGlob(filepath.Join("templates", "*.html")))
}

type ErrorPageData struct {
	Code     string
	ErrorMsg string
}

type PageData struct {
	UserID         string
	UserName       string
	Avatar         string
	Categories     []database.Category
	Users          []database.User
	Posts          []database.Post
	SelectedTab    string
	SelectedFilter string
	Notifications  []database.Notification
}

func errHandler(w http.ResponseWriter, _ *http.Request, errData *ErrorPageData) {
	err := templates.ExecuteTemplate(w, "error.html", errData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func MainPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		errHandler(w, r, &err)
		return
	}

	if r.Method != "GET" {
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		errHandler(w, r, &err)
		return
	}

	// Redirect to /?tab=posts&filter=all if no tab is specified
	if r.URL.Query().Get("tab") == "" {
		http.Redirect(w, r, "/?tab=posts&filter=all", http.StatusFound)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Database connection failed"}
		errHandler(w, r, &err)
		return
	}
	defer db.Close()

	categories, err := database.GetAllCategories(db)
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch categories"}
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
		err := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch posts"}
		errHandler(w, r, &err)
		return
	}

	users, err := database.GetAllUsers(db)
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch users"}
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login" {
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		errHandler(w, r, &err)
		return
	}

	if r.Method == "POST" {
		email := r.FormValue("email")
		password := r.FormValue("password")

		db, err := sql.Open("sqlite3", "./database/main.db")
		if err != nil {
			err := ErrorPageData{Code: "500", ErrorMsg: "Database connection failed"}
			errHandler(w, r, &err)
			return
		}
		defer db.Close()

		var userID int
		var dbPassword string
		err = db.QueryRow("SELECT userid, password FROM user WHERE email = ?", email).Scan(&userID, &dbPassword)
		if err != nil {
			if err == sql.ErrNoRows {
				// No user found with the given email
				err = templates.ExecuteTemplate(w, "login.html", map[string]interface{}{
					"ErrorMsg": "Invalid email or password",
				})
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}
			err := ErrorPageData{Code: "500", ErrorMsg: "Database query failed"}
			errHandler(w, r, &err)
			return
		}

		// Check if the password is correct
		if password != dbPassword {
			err = templates.ExecuteTemplate(w, "login.html", map[string]interface{}{
				"ErrorMsg": "Invalid email or password",
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// If login is successful, redirect to the Home page with user ID
		http.Redirect(w, r, fmt.Sprintf("/home?user=%d&tab=posts&filter=all", userID), http.StatusSeeOther)
		return
	}

	err := templates.ExecuteTemplate(w, "login.html", nil)
	if err != nil {
		errData := ErrorPageData{Code: "500", ErrorMsg: "Failed to parse template"}
		errHandler(w, r, &errData)
	}
}

func SignupPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/signup" {
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		errHandler(w, r, &err)
		return
	}

	// Render the template
	err := templates.ExecuteTemplate(w, "signup.html", nil)
	if err != nil {
		errData := ErrorPageData{Code: "500", ErrorMsg: "Failed to parse template"}
		errHandler(w, r, &errData)
	}
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/home" {
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		errHandler(w, r, &err)
		return
	}

	if r.Method != "GET" {
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		errHandler(w, r, &err)
		return
	}

	// Redirect to /home?tab=posts&filter=all if no tab is specified
	if r.URL.Query().Get("tab") == "" {
		userID := r.URL.Query().Get("user")
		http.Redirect(w, r, fmt.Sprintf("/home?user=%s&tab=posts&filter=all", userID), http.StatusFound)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Database connection failed"}
		errHandler(w, r, &err)
		return
	}
	defer db.Close()

	categories, err := database.GetAllCategories(db)
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch categories"}
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
		err := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch posts"}
		errHandler(w, r, &err)
		return
	}

	users, err := database.GetAllUsers(db)
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch users"}
		errHandler(w, r, &err)
		return
	}

	userID := r.URL.Query().Get("user")
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		err := ErrorPageData{Code: "400", ErrorMsg: "Invalid user ID"}
		errHandler(w, r, &err)
		return
	}

	var userName string
	var avatar sql.NullString
	err = db.QueryRow("SELECT username, avatar FROM user WHERE userid = ?", userID).Scan(&userName, &avatar)
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch user data"}
		errHandler(w, r, &err)
		return
	}

	notifications, err := database.GetLastNotifications(db, userIDInt)
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch notifications"}
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
	}

	err = templates.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
