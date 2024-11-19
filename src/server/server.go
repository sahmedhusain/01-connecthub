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
	
		return
	}
	tmpl, err := template.ParseFiles("Templates/index.html")
	if err != nil {
		// error 500
		ErrorHandler(w, r, http.StatusInternalServerError, "")
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
	case http.StatusBadRequest:
		//400
		errorMessage = "400 - Bad request"
		if errM != "" {
			//400 with extra message
			errorMessage += ": " + errM
		}
	case http.StatusInternalServerError:
		//500
		errorMessage = "500 - Internal server error"
	case http.StatusMethodNotAllowed:
		//405
		errorMessage = "405 - Method not allowed"
	default:
		errorMessage = "Unexpected error"
	}
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, errorMessage)
}

