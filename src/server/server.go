package server

import (
	"database/sql"
	"forum/database"
	"html/template"
	"net/http"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type ErrorPageData struct {
	Code     string
	ErrorMsg string
}

var once sync.Once

func errHandler(w http.ResponseWriter, _ *http.Request, err *ErrorPageData) {
	errorTemp, erra := template.ParseFiles("templates/error.html")
	if erra != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	errorTemp.Execute(w, err)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles("templates/" + tmpl)
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Failed to parse template"}
		w.WriteHeader(http.StatusInternalServerError)
		errHandler(w, nil, &err)
		return
	}
	t.Execute(w, data)
}

func MainPage(w http.ResponseWriter, r *http.Request) {
	// Validating the request path
	if r.URL.Path != "/" {
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		w.WriteHeader(http.StatusNotFound)
		errHandler(w, r, &err)
		return
	}

	// Validating the request method
	if r.Method != "GET" {
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		errHandler(w, r, &err)
		return
	}

	// Open DB connection
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Database connection failed"}
		w.WriteHeader(http.StatusInternalServerError)
		errHandler(w, r, &err)
		return
	}
	defer db.Close()

	// Fetch categories, users, comments, and posts
	categories, err := database.GetAllCategories(db)
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch categories"}
		w.WriteHeader(http.StatusInternalServerError)
		errHandler(w, r, &err)
		return
	}

	users, err := database.GetAllUsers(db)
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch users"}
		w.WriteHeader(http.StatusInternalServerError)
		errHandler(w, r, &err)
		return
	}

	comments, err := database.GetComments(db)
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch comments"}
		w.WriteHeader(http.StatusInternalServerError)
		errHandler(w, r, &err)
		return
	}

	posts, err := database.GetAllPosts(db)
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch posts"}
		w.WriteHeader(http.StatusInternalServerError)
		errHandler(w, r, &err)
		return
	}

	// Combine data for template
	data := struct {
		Categories []database.Category
		Users      []database.User
		Comments   []database.Comment
		Posts      []database.Post
	}{
		Categories: categories,
		Users:      users,
		Comments:   comments,
		Posts:      posts,
	}

	// Render template
	renderTemplate(w, "index.html", data)
}

func AboutPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		errHandler(w, r, &err)
		return
	}
	renderTemplate(w, "about.html", nil)
}

func HelpPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		errHandler(w, r, &err)
		return
	}
	renderTemplate(w, "help.html", nil)
}

func PrivacyPolicyPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		errHandler(w, r, &err)
		return
	}
	renderTemplate(w, "privacy_policy.html", nil)
}

func ActivityCentrePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		errHandler(w, r, &err)
		return
	}
	renderTemplate(w, "activity_centre.html", nil)
}

func ConnectionsPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		errHandler(w, r, &err)
		return
	}
	renderTemplate(w, "connections.html", nil)
}

func ContentPolicyPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		errHandler(w, r, &err)
		return
	}
	renderTemplate(w, "content_policy.html", nil)
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		renderTemplate(w, "login.html", nil)
	} else if r.Method == "POST" {
		// Handle login form submission
		// Add your login logic here
		// For now, just redirect to the main page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		errHandler(w, r, &err)
	}
}

func NotificationsPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		errHandler(w, r, &err)
		return
	}
	renderTemplate(w, "notifications.html", nil)
}

func ProfilePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		errHandler(w, r, &err)
		return
	}
	renderTemplate(w, "profile.html", nil)
}

func SignupPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		renderTemplate(w, "signup.html", nil)
	} else if r.Method == "POST" {
		// Handle signup form submission
		// Add your signup logic here
		// For now, just redirect to the login page
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		errHandler(w, r, &err)
	}
}

func UserAgreementPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		errHandler(w, r, &err)
		return
	}
	renderTemplate(w, "user_agreement.html", nil)
}

func IndexsPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		errHandler(w, r, &err)
		return
	}
	renderTemplate(w, "indexs.html", nil)
}
