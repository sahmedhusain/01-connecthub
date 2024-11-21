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
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		w.WriteHeader(http.StatusNotFound)
		errHandler(w, r, &err)
		return
	}
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		w.WriteHeader(http.StatusInternalServerError)
		errHandler(w, r, &err)
		return
	}
	tmpl.Execute(w, nil)
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
