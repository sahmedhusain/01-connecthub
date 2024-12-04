package server

import (
	"forum/database"
	"database/sql"
	"html/template"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type ErrorPageData struct {
	Code     string
	ErrorMsg string
}

func errHandler(w http.ResponseWriter, r *http.Request, err *ErrorPageData) {
	errorTemp, erra := template.ParseFiles("templates/index.html")
	if erra != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	errorTemp.Execute(w, err)
}

func MainPage(w http.ResponseWriter, r *http.Request) {
	// Open DB connection
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Database connection failed")
		return
	}
	defer db.Close()

	// Fetch categories and users
	categories, err := database.GetAllCategories(db)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}

	users, err := database.GetAllUsers(db)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	// Combine data for template
	data := struct {
		Categories []database.Category
		Users      []database.User
	}{
		Categories: categories,
		Users:      users,
	}

	// Render template
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "")
		return
	}
	tmpl.Execute(w, data)

}

func ErrorHandler(w http.ResponseWriter, r *http.Request, statusCode int, errM string) {
	var errorData ErrorPageData
	switch statusCode {
	case http.StatusNotFound:
		errorData = ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
	case http.StatusBadRequest:
		errorData = ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
		if errM != "" {
			errorData.ErrorMsg += ": " + errM
		}
	case http.StatusInternalServerError:
		errorData = ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
	case http.StatusMethodNotAllowed:
		errorData = ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
	default:
		errorData = ErrorPageData{Code: "000", ErrorMsg: "UNEXPECTED ERROR"}
	}
	w.WriteHeader(statusCode)
	errHandler(w, r, &errorData)
}
