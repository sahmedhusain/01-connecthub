package main

import (
	"fmt"
	"forum/src/server"
	"log"
	"net/http"
)

func main() {
	http.NewServeMux()
	http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("templates"))))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", server.MainPage)

	fmt.Println("Server running on http://localhost:8080 \nTo stop the server press Ctrl+C")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
