package server

import (
	"forum/database"
	"html/template"
	"net/http"
)

type ErrorPageData struct {
	Code     string
	ErrorMsg string
}

func errHandler(w http.ResponseWriter, r *http.Request, err *ErrorPageData) {
	errorTemp, erra := template.ParseFiles("templates/error.html")
	if erra != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	errorTemp.Execute(w, err)
}

func MainPage(w http.ResponseWriter, r *http.Request) {
	database.DataBase()
	if r.URL.Path != "/" {
		// error 404
		ErrorHandler(w, r, http.StatusNotFound, "")
		http.ServeFile(w, r, "templates/error.html")
		return
	}
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		// error 500
		ErrorHandler(w, r, http.StatusInternalServerError, "")
		http.ServeFile(w, r, "templates/error.html")
		return
	}
	tmpl.Execute(w, nil)
}


func ErrorHandler(w http.ResponseWriter, r *http.Request, statusCode int, errM string) {
	var errorData ErrorPageData
	switch statusCode {
	case http.StatusNotFound:
		//404
		errorData.ErrorMsg = "404 - Page not found"
		http.ServeFile(w, r, "templates/error.html")
	case http.StatusBadRequest:
		//400
		errorData.ErrorMsg = "400 - Bad request"
		http.ServeFile(w, r, "templates/error.html")
		if errM != "" {
			errorData.ErrorMsg += ": " + errM
		}
	case http.StatusInternalServerError:
		//500
		errorData.ErrorMsg = "500 - Internal server error"
		http.ServeFile(w, r, "templates/error.html")
	case http.StatusMethodNotAllowed:
		//405
		errorData.ErrorMsg = "405 - Method not allowed"
		http.ServeFile(w, r, "templates/error.html")
	default:
		errorData.ErrorMsg = "Unexpected error"
		http.ServeFile(w, r, "templates/error.html")
	}
	w.WriteHeader(statusCode)
	errHandler(w, r, &errorData)
}
