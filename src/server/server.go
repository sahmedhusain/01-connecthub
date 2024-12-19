package server

import (
	"database/sql"
	"fmt"
	"forum/database"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/sessions"

	_ "github.com/mattn/go-sqlite3"
)

var (
	templates *template.Template
	store     = sessions.NewCookieStore([]byte("your-secret-key"))
)

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

func errHandler(w http.ResponseWriter, r *http.Request, errData *ErrorPageData) {
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

	notifications, err := database.GetLastNotifications(db, strconv.Itoa(userIDInt))
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

func NewPostPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/newpost" {
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		errHandler(w, r, &err)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Database connection failed"}
		errHandler(w, r, &err)
		return
	}
	defer db.Close()

	switch r.Method {
	case "GET":
		categories, err := database.GetAllCategories(db)
		if err != nil {
			err := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch categories"}
			errHandler(w, r, &err)
			return
		}
		data := struct {
			Categories []database.Category
		}{
			Categories: categories,
		}
		err = templates.ExecuteTemplate(w, "newpost.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case "POST":
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Could not parse form", http.StatusBadRequest)
			return
		}

		content := r.FormValue("content")
		categoryIDs := r.Form["categories"]
		userID := r.FormValue("user_id") // Replace with actual user ID retrieval logic

		var imagePath sql.NullString
		file, handler, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			imagePath.String = fmt.Sprintf("static/uploads/%s", handler.Filename)
			imagePath.Valid = true
			f, err := os.OpenFile(imagePath.String, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				http.Error(w, "Unable to save image", http.StatusInternalServerError)
				return
			}
			defer f.Close()
			io.Copy(f, file)
		}

		postID, err := database.InsertPost(db, content, imagePath, userID)
		if err != nil {
			err := ErrorPageData{Code: "500", ErrorMsg: "Failed to create post"}
			errHandler(w, r, &err)
			return
		}

		for _, catID := range categoryIDs {
			categoryID, err := strconv.Atoi(catID)
			if err != nil {
				continue
			}
			err = database.InsertPostCategory(db, postID, categoryID)
			if err != nil {
				err := ErrorPageData{Code: "500", ErrorMsg: "Failed to associate categories"}
				errHandler(w, r, &err)
				return
			}
		}

		http.Redirect(w, r, "/home", http.StatusSeeOther)
	default:
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		errHandler(w, r, &err)
	}
}

func SettingsPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/settings" {
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		errHandler(w, r, &err)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Database connection failed"}
		errHandler(w, r, &err)
		return
	}
	defer db.Close()

	// Retrieve UserID from session
	session, _ := store.Get(r, "session-name")
	userID, ok := session.Values["userID"].(string)
	if !ok || userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	switch r.Method {
	case "GET":
		var user database.User
		err := db.QueryRow("SELECT id, first_name, last_name, username, email, avatar FROM user WHERE id = ?", userID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Email, &user.Avatar)
		if err != nil {
			errData := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch user data"}
			errHandler(w, r, &errData)
			return
		}

		data := struct {
			UserID    string
			FirstName string
			LastName  string
			Username  string
			Email     string
			Avatar    string
		}{
			UserID:    userID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Username:  user.Username,
			Email:     user.Email,
			Avatar:    user.Avatar.String,
		}

		err = templates.ExecuteTemplate(w, "settings.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "POST":
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Could not parse form", http.StatusBadRequest)
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

			// Ensure the directory exists
			os.MkdirAll("static/uploads", os.ModePerm)

			f, err := os.OpenFile(avatarPath.String, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				http.Error(w, "Unable to save avatar", http.StatusInternalServerError)
				return
			}
			defer f.Close()
			io.Copy(f, file)
		} else if err != http.ErrMissingFile {
			http.Error(w, "Error uploading file", http.StatusInternalServerError)
			return
		}

		if password != "" {
			_, err = db.Exec("UPDATE user SET first_name = ?, last_name = ?, username = ?, email = ?, password = ?, avatar = ? WHERE id = ?", firstName, lastName, username, email, password, avatarPath, userID)
		} else {
			_, err = db.Exec("UPDATE user SET first_name = ?, last_name = ?, username = ?, email = ?, avatar = ? WHERE id = ?", firstName, lastName, username, email, avatarPath, userID)
		}

		if err != nil {
			http.Error(w, "Failed to update user data", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/settings", http.StatusSeeOther)
	default:
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		errHandler(w, r, &err)
		return
	}
}

func NotificationsPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/notifications" {
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		errHandler(w, r, &err)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Database connection failed"}
		errHandler(w, r, &err)
		return
	}
	defer db.Close()

	// Retrieve UserID from session
	session, _ := store.Get(r, "session-name")
	userID, ok := session.Values["userID"].(string)
	if !ok || userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	notifications, err := database.GetLastNotifications(db, userID)
	if err != nil {
		errData := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch notifications"}
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func MyProfilePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/myprofile" {
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		errHandler(w, r, &err)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Database connection failed"}
		errHandler(w, r, &err)
		return
	}
	defer db.Close()

	// Retrieve UserID from session
	session, _ := store.Get(r, "session-name")
	userID, ok := session.Values["userID"].(string)
	if !ok || userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var user database.User
	err = db.QueryRow("SELECT id, first_name, last_name, username, avatar FROM user WHERE id = ?", userID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Avatar)
	if err != nil {
		errData := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch user data"}
		errHandler(w, r, &errData)
		return
	}

	posts, err := database.GetUserPosts(db, userID)
	if err != nil {
		errData := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch user posts"}
		errHandler(w, r, &errData)
		return
	}

	followersCount, err := database.GetFollowersCount(db, userID)
	if err != nil {
		errData := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch followers count"}
		errHandler(w, r, &errData)
		return
	}

	followingCount, err := database.GetFollowingCount(db, userID)
	if err != nil {
		errData := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch following count"}
		errHandler(w, r, &errData)
		return
	}

	friendsCount, err := database.GetFriendsCount(db, userID)
	if err != nil {
		errData := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch friends count"}
		errHandler(w, r, &errData)
		return
	}

	data := struct {
		UserID         string
		FirstName      string
		LastName       string
		Username       string
		Avatar         string
		PostsCount     int
		FollowersCount int
		FollowingCount int
		FriendsCount   int
		Posts          []database.Post
	}{
		UserID:         userID,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Username:       user.Username,
		Avatar:         user.Avatar.String,
		PostsCount:     len(posts),
		FollowersCount: followersCount,
		FollowingCount: followingCount,
		FriendsCount:   friendsCount,
		Posts:          posts,
	}

	err = templates.ExecuteTemplate(w, "myprofile.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ProfilePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/profile" {
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		errHandler(w, r, &err)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Database connection failed"}
		errHandler(w, r, &err)
		return
	}
	defer db.Close()

	// Retrieve UserID from session
	session, _ := store.Get(r, "session-name")
	userID, ok := session.Values["userID"].(string)
	if !ok || userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Retrieve ProfileUserID from query parameters
	profileUserID := r.URL.Query().Get("user")
	if profileUserID == "" {
		err := ErrorPageData{Code: "400", ErrorMsg: "Bad Request"}
		errHandler(w, r, &err)
		return
	}

	var user database.User
	err = db.QueryRow("SELECT id, first_name, last_name, username, avatar FROM user WHERE id = ?", profileUserID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Avatar)
	if err != nil {
		errData := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch user data"}
		errHandler(w, r, &errData)
		return
	}

	posts, err := database.GetUserPosts(db, profileUserID)
	if err != nil {
		errData := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch user posts"}
		errHandler(w, r, &errData)
		return
	}

	followersCount, err := database.GetFollowersCount(db, profileUserID)
	if err != nil {
		errData := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch followers count"}
		errHandler(w, r, &errData)
		return
	}

	followingCount, err := database.GetFollowingCount(db, profileUserID)
	if err != nil {
		errData := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch following count"}
		errHandler(w, r, &errData)
		return
	}

	friendsCount, err := database.GetFriendsCount(db, profileUserID)
	if err != nil {
		errData := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch friends count"}
		errHandler(w, r, &errData)
		return
	}

	isFollowing, err := database.IsFollowing(db, userID, profileUserID)
	if err != nil {
		errData := ErrorPageData{Code: "500", ErrorMsg: "Failed to check following status"}
		errHandler(w, r, &errData)
		return
	}

	data := struct {
		UserID                string
		Avatar                string
		ProfileUserID         string
		ProfileFirstName      string
		ProfileLastName       string
		ProfileUsername       string
		ProfileAvatar         string
		ProfilePostsCount     int
		ProfileFollowersCount int
		ProfileFollowingCount int
		ProfileFriendsCount   int
		ProfilePosts          []database.Post
		IsFollowing           bool
	}{
		UserID:                userID,
		Avatar:                session.Values["avatar"].(string),
		ProfileUserID:         profileUserID,
		ProfileFirstName:      user.FirstName,
		ProfileLastName:       user.LastName,
		ProfileUsername:       user.Username,
		ProfileAvatar:         user.Avatar.String,
		ProfilePostsCount:     len(posts),
		ProfileFollowersCount: followersCount,
		ProfileFollowingCount: followingCount,
		ProfileFriendsCount:   friendsCount,
		ProfilePosts:          posts,
		IsFollowing:           isFollowing,
	}

	err = templates.ExecuteTemplate(w, "profile.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
