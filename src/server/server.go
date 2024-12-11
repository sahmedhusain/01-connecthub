package server

import (
	"database/sql"
	"forum/database"
	"html/template"
	"net/http"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var templates *template.Template

func init() {
	// Parse all templates
	templates = template.Must(template.ParseGlob(filepath.Join("templates", "*.html")))
}

type ErrorPageData struct {
	Code     string
	ErrorMsg string
}

func errHandler(w http.ResponseWriter, _ *http.Request, errData *ErrorPageData) {
	err := templates.ExecuteTemplate(w, "error.html", errData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Failed to parse template"}
		w.WriteHeader(http.StatusInternalServerError)
		errHandler(w, nil, &err)
		return
	}
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
	renderTemplate(w, "index.html", data)
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		renderTemplate(w, "login.html", nil)
	} else if r.Method == "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		errHandler(w, r, &err)
	}
}

func SignupPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		renderTemplate(w, "signup.html", nil)
	} else if r.Method == "POST" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		errHandler(w, r, &err)
	}
}

func IndexsPage(w http.ResponseWriter, r *http.Request) {
	// Open DB connection
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Database connection failed"}
		w.WriteHeader(http.StatusInternalServerError)
		errHandler(w, r, &err)
		return
	}
	defer db.Close()

	// Fetch posts from the database
	posts, err := database.GetAllPosts(db)
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "Failed to fetch posts"}
		w.WriteHeader(http.StatusInternalServerError)
		errHandler(w, r, &err)
		return
	}

	// Render the template
	data := struct {
		Posts []database.Post 
	}{
		Posts: posts,
	}

	err = templates.ExecuteTemplate(w, "indexs.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
