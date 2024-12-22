package server

import (
	"database/sql"
	"fmt"
	"forum/database"
	"html/template"
	"io"
	"log"
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
	RoleID         int
	Post           database.Post
	Comments       []database.Comment
}

func errHandler(w http.ResponseWriter, _ *http.Request, errData *ErrorPageData) {
	err := templates.ExecuteTemplate(w, "error.html", errData)
	if err != nil {
		log.Println("Error rendering error page:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func MainPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
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

	// Redirect to /?tab=posts&filter=all if no tab is specified
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

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login" {
		log.Println("Redirecting to Home page")
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		errHandler(w, r, &err)
		return
	}

	if r.Method == "POST" {
		email := r.FormValue("email")
		password := r.FormValue("password")

		db, err := sql.Open("sqlite3", "./database/main.db")
		if err != nil {
			log.Println("Database connection failed")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
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
					log.Println("Error rendering login page:", err)
					errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
					errHandler(w, r, &errData)
				}
				return
			}
			log.Println("Failed to fetch user data")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		// Check if the password is correct
		if password != dbPassword {
			err = templates.ExecuteTemplate(w, "login.html", map[string]interface{}{
				"ErrorMsg": "Invalid email or password",
			})
			if err != nil {
				log.Println("Error rendering login page:", err)
				errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &errData)
			}
			return
		}

		// Set the userID in the session
		session, _ := store.Get(r, "session-name")
		session.Values["userID"] = strconv.Itoa(userID)
		err = session.Save(r, w)
		if err != nil {
			log.Println("Error saving session:", err)
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}

		log.Println("User logged in with userID:", userID)

		// If login is successful, redirect to the Home page with user ID
		log.Println("Redirecting to Home page with user ID")
		http.Redirect(w, r, fmt.Sprintf("/home?user=%d&tab=posts&filter=all", userID), http.StatusSeeOther)
		return
	}

	err := templates.ExecuteTemplate(w, "login.html", nil)
	if err != nil {
		log.Println("Error rendering login page:", err)
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &errData)
	}
}

func SignupPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/signup" {
		log.Println("Redirecting to Home page")
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		errHandler(w, r, &err)
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm-password")

		if password != confirmPassword {
			err := templates.ExecuteTemplate(w, "signup.html", map[string]string{
				"ErrorMessage": "Passwords do not match",
			})
			if err != nil {
				log.Println("Error rendering signup page:", err)
				errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &errData)
			}
			return
		}

		db, err := sql.Open("sqlite3", "./database/main.db")
		if err != nil {
			log.Println("Database connection failed")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}
		defer db.Close()

		var usernameExists, emailExists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM user WHERE username = ?)", username).Scan(&usernameExists)
		if err != nil {
			log.Println("Failed to check if username exists")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}

		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM user WHERE email = ?)", email).Scan(&emailExists)
		if err != nil {
			log.Println("Failed to check if email exists")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}

		if usernameExists {
			err := templates.ExecuteTemplate(w, "signup.html", map[string]string{
				"ErrorMessage": "Username already exists",
			})
			if err != nil {
				log.Println("Error rendering signup page:", err)
				errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &errData)
			}
			return
		}

		if emailExists {
			err := templates.ExecuteTemplate(w, "signup.html", map[string]string{
				"ErrorMessage": "Email already exists",
			})
			if err != nil {
				log.Println("Error rendering signup page:", err)
				errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &errData)
			}
			return
		}

		_, err = db.Exec("INSERT INTO user (username, email, password) VALUES (?, ?, ?)", username, email, password)
		if err != nil {
			log.Println("Failed to insert user data")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}
		log.Println("User registered with username:", username)
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	err := templates.ExecuteTemplate(w, "signup.html", nil)
	if err != nil {
		log.Println("Error rendering signup page:", err)
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &errData)
	}
}

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

func NewPostPage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Render the new post page template
		err := templates.ExecuteTemplate(w, "newpost.html", nil)
		if err != nil {
			log.Println("Error rendering new post page:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

	case "POST":
		// Parse the form data
		err := r.ParseForm()
		if err != nil {
			log.Println("Failed to parse form data")
			err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
			errHandler(w, r, &err)
			return
		}

		// Get the post content and user ID
		content := r.FormValue("content")
		userID := r.FormValue("user")

		if content == "" || userID == "" {
			log.Println("Invalid form data")
			err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
			errHandler(w, r, &err)
			return
		}

		// Insert the new post into the database
		db, err := sql.Open("sqlite3", "./database/main.db")
		if err != nil {
			log.Println("Database connection failed")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}
		defer db.Close()

		image := sql.NullString{String: r.FormValue("image"), Valid: r.FormValue("image") != ""}
		postID, err := database.InsertPost(db, content, image, userID)
		if err != nil {
			log.Println("Failed to insert post data")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		// Associate the post with categories
		categoryIDs := r.Form["categories"]
		for _, categoryID := range categoryIDs {
			categoryIDInt, err := strconv.Atoi(categoryID)
			if err != nil {
				log.Println("Failed to parse category ID")
				err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
				errHandler(w, r, &err)
				return
			}
			err = database.InsertPostCategory(db, postID, categoryIDInt)
			if err != nil {
				log.Println("Failed to insert post category data")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
		}
		log.Println("New post created with ID:", postID)
		http.Redirect(w, r, "/home?user="+userID, http.StatusSeeOther)

	default:
		log.Println("Method not allowed")
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		errHandler(w, r, &err)
		return
	}
}

func SettingsPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/settings" {
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

	log.Println("UserID retrieved from session:", userID)

	switch r.Method {
	case "GET":
		var user database.User
		err := db.QueryRow("SELECT id, first_name, last_name, username, email, avatar FROM user WHERE id = ?", userID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Email, &user.Avatar)
		if err != nil {
			log.Println("Failed to fetch user data")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
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
			log.Println("Error rendering settings page:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}
	case "POST":
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			log.Println("Failed to parse form data")
			err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
			errHandler(w, r, &err)
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
				log.Println("Failed to open file for writing")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
			defer f.Close()
			io.Copy(f, file)
		} else if err != http.ErrMissingFile {
			log.Println("Failed to upload avatar")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
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
			errHandler(w, r, &err)
			return
		}
		log.Println("User data updated with ID:", userID)
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
	default:
		log.Println("Method not allowed")
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		errHandler(w, r, &err)
		return
	}
}

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

	// Retrieve UserID from session
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

func MyProfilePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/myprofile" {
		http.NotFound(w, r)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	err = db.QueryRow("SELECT userid, F_name, L_name, Username, Avatar FROM user WHERE userid = ?", userID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Avatar)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	posts, err := database.GetUserPosts(db, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	followersCount, err := database.GetFollowersCount(db, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	followingCount, err := database.GetFollowingCount(db, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	friendsCount, err := database.GetFriendsCount(db, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view := r.URL.Query().Get("view")
	var followers, following, friends []database.User

	if view == "followers" {
		followers, err = database.GetFollowers(db, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if view == "following" {
		following, err = database.GetFollowing(db, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if view == "friends" {
		friends, err = database.GetFriends(db, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
		View           string
		Followers      []database.User
		Following      []database.User
		Friends        []database.User
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
		View:           view,
		Followers:      followers,
		Following:      following,
		Friends:        friends,
	}

	err = templates.ExecuteTemplate(w, "myprofile.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ProfilePage(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/profile" {
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

    // Retrieve ProfileUserID from query parameters
    profileUserID := r.URL.Query().Get("user")
    if profileUserID == "" {
        log.Println("ProfileUserID not found in query parameters")
        err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
        errHandler(w, r, &err)
        return
    }

    var user database.User
    err = db.QueryRow("SELECT userid, F_name, L_name, Username, Avatar FROM user WHERE userid = ?", profileUserID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Avatar)
    if err != nil {
        log.Println("Failed to fetch user data")
        errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
        errHandler(w, r, &errData)
        return
    }

    posts, err := database.GetUserPosts(db, profileUserID)
    if err != nil {
        log.Println("Failed to fetch user posts")
        errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
        errHandler(w, r, &errData)
        return
    }

    followersCount, err := database.GetFollowersCount(db, profileUserID)
    if err != nil {
        log.Println("Failed to fetch followers count")
        errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
        errHandler(w, r, &errData)
        return
    }

    followingCount, err := database.GetFollowingCount(db, profileUserID)
    if err != nil {
        log.Println("Failed to fetch following count")
        errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
        errHandler(w, r, &errData)
        return
    }

    friendsCount, err := database.GetFriendsCount(db, profileUserID)
    if err != nil {
        log.Println("Failed to fetch friends count")
        errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
        errHandler(w, r, &errData)
        return
    }

    isFollowing, err := database.IsFollowing(db, userID, profileUserID)
    if err != nil {
        log.Println("Failed to check if user is following")
        errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
        errHandler(w, r, &errData)
        return
    }

    view := r.URL.Query().Get("view")
    var followers, following []database.User

    if view == "followers" {
        followers, err = database.GetFollowers(db, profileUserID)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    } else if view == "following" {
        following, err = database.GetFollowing(db, profileUserID)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
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
        View                  string
        Followers             []database.User
        Following             []database.User
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
        View:                  view,
        Followers:             followers,
        Following:             following,
    }

    err = templates.ExecuteTemplate(w, "profile.html", data)
    if err != nil {
        log.Println("Error rendering profile page:", err)
        err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
        errHandler(w, r, &err)
        return
    }
}

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

	// Check if the user is an admin
	var roleID int
	err = db.QueryRow("SELECT role_id FROM user WHERE id = ?", userID).Scan(&roleID)
	if err != nil || roleID != 1 {
		log.Println("User is not an admin, redirecting to Home page")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
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
					log.Println("Failed to fetch user sessions")
					errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
					errHandler(w, r, &errData)
					return
				}
			}
		}

		data := struct {
			UserID          string
			Avatar          string
			Users           []database.User
			Posts           []database.Post
			Categories      []database.Category
			Reports         []database.Report
			TotalUsers      int
			TotalPosts      int
			TotalCategories int
			UserLogs        []database.UserLog
			UserSessions    []database.UserSession
		}{
			UserID:          userID,
			Avatar:          session.Values["avatar"].(string),
			Users:           users,
			Posts:           posts,
			Categories:      categories,
			Reports:         reports,
			TotalUsers:      totalUsers,
			TotalPosts:      totalPosts,
			TotalCategories: totalCategories,
			UserLogs:        userLogs,
			UserSessions:    userSessions,
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
				log.Println("Failed to delete user")
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
		} else {
			for key, values := range r.Form {
				if len(values) > 0 && key[:5] == "role_" {
					userID := key[5:]
					roleID := values[0]
					_, err := db.Exec("UPDATE user SET role_id = ? WHERE id = ?", roleID, userID)
					if err != nil {
						log.Println("Failed to update user role")
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

func ModeratorPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/moderator" {
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

	// Check if the user is a moderator
	var roleID int
	err = db.QueryRow("SELECT role_id FROM user WHERE id = ?", userID).Scan(&roleID)
	if err != nil || roleID != 2 {
		log.Println("User is not a moderator, redirecting to Home page")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	switch r.Method {
	case "GET":
		posts, err := database.GetAllPosts(db)
		if err != nil {
			log.Println("Failed to fetch posts")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}

		data := struct {
			UserID string
			Avatar string
			Posts  []database.Post
		}{
			UserID: userID,
			Avatar: session.Values["avatar"].(string),
			Posts:  posts,
		}

		err = templates.ExecuteTemplate(w, "moderator.html", data)
		if err != nil {
			log.Println("Error rendering moderator page:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}
	case "POST":
		r.ParseForm()
		if r.FormValue("delete_post") != "" {
			postID := r.FormValue("delete_post")
			_, err := db.Exec("DELETE FROM post WHERE postid = ?", postID)
			if err != nil {
				log.Println("Failed to delete post")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
		} else if r.FormValue("report_post") != "" {
			postID := r.FormValue("report_post")
			reportReason := r.FormValue("report_reason")
			_, err := db.Exec("INSERT INTO reports (post_id, reported_by, report_reason) VALUES (?, ?, ?)", postID, userID, reportReason)
			if err != nil {
				log.Println("Failed to report post")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
		}
		log.Println("Moderator action completed")
		http.Redirect(w, r, "/moderator", http.StatusSeeOther)
	default:
		log.Println("Method not allowed")
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		errHandler(w, r, &err)
		return
	}
}

func PostPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/post" {
		log.Println("Redirecting to Home page")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	if r.Method != "GET" {
		log.Println("Method not allowed")
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		errHandler(w, r, &err)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Error opening database:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}
	defer db.Close()

	postID := r.URL.Query().Get("id")
	if postID == "" {
		log.Println("Post ID not found in query parameters")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	var post database.Post
	err = db.QueryRow(`
        SELECT post.postid, post.image, post.content, post.post_at, post.user_userid, user.Username, user.F_name, user.L_name, user.Avatar,
               (SELECT COUNT(*) FROM likes WHERE likes.post_postid = post.postid) AS Likes,
               (SELECT COUNT(*) FROM dislikes WHERE dislikes.post_postid = post.postid) AS Dislikes,
               (SELECT COUNT(*) FROM comment WHERE comment.post_postid = post.postid) AS Comments
        FROM post
        JOIN user ON post.user_userid = user.userid
        WHERE post.postid = ?
    `, postID).Scan(&post.PostID, &post.Image, &post.Content, &post.PostAt, &post.UserUserID, &post.Username, &post.FirstName, &post.LastName, &post.Avatar, &post.Likes, &post.Dislikes, &post.Comments)
	if err != nil {
		log.Println("Error querying post:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}

	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		log.Println("Error converting post ID to integer:", err)
		err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
		errHandler(w, r, &err)
		return
	}
	comments, err := database.GetCommentsForPost(db, postIDInt)
	if err != nil {
		log.Println("Error getting comments for post:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}

	data := PageData{
		Post:     post,
		Comments: comments,
	}

	err = templates.ExecuteTemplate(w, "post.html", data)
	if err != nil {
		log.Println("Error executing template:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}
}

func LikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Println("Method not allowed")
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		errHandler(w, r, &err)
		return
	}

	postID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil {
		log.Println("Invalid post ID")
		err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
		errHandler(w, r, &err)
		return
	}

	userID := r.FormValue("user_id")

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Error opening database:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}
	defer db.Close()

	err = database.ToggleLike(db, postID, userID)
	if err != nil {
		log.Println("Error toggling like:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}
	log.Println("Like toggled")
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func DislikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Println("Method not allowed")
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		errHandler(w, r, &err)
		return
	}

	postID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil {
		log.Println("Invalid post ID")
		err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
		errHandler(w, r, &err)
		return
	}

	userID := r.FormValue("user_id")

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Error opening database:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}
	defer db.Close()

	err = database.ToggleDislike(db, postID, userID)
	if err != nil {
		log.Println("Error toggling dislike:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}
	log.Println("Dislike toggled")
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}
