package server

import (
	"log"
	"net/http"
)

func errHandler(w http.ResponseWriter, _ *http.Request, errData *ErrorPageData) {
	err := templates.ExecuteTemplate(w, "error.html", errData)
	if err != nil {
		log.Println("Error rendering error page:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	// errCodeInt, _ := strconv.Atoi(errData.Code)

	// w.WriteHeader(errCodeInt)
}
