package server

import (
	"fmt"
	"html/template"
	"net/http"
	
)

func MainPage(w http.ResponseWriter, r *http.Request) {
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
	tmpl.Execute(w, r)
}


func ErrorHandler(w http.ResponseWriter, r *http.Request, statusCode int, errM string) {
	var errorMessage string
	switch statusCode {
	case http.StatusNotFound:
		//404
		errorMessage = " 404 - Page not found"
		http.ServeFile(w, r, "templates/error.html")
	case http.StatusBadRequest:
		//400
		errorMessage = "400 - Bad request"
		http.ServeFile(w, r, "templates/error.html")
		if errM != "" {
			//400 with extra message
			errorMessage += ": " + errM
		}
	case http.StatusInternalServerError:
		//500
		errorMessage = "500 - Internal server error"
		http.ServeFile(w, r, "templates/error.html")
	case http.StatusMethodNotAllowed:
		//405
		errorMessage = "405 - Method not allowed"
		http.ServeFile(w, r, "templates/error.html")
	default:
		errorMessage = "Unexpected error"
		http.ServeFile(w, r, "templates/error.html")
	}
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, errorMessage)
}

