package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	db "forum/database"
	"forum/src/server"
)

var templates *template.Template

func init() {
	db.DataBase()
	// Parse all templates
	templates = template.Must(template.ParseGlob(filepath.Join("templates", "*.html")))
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	http.HandleFunc("/", server.MainPage)
	http.HandleFunc("/login", server.LoginPage)
	http.HandleFunc("/signup", server.SignupPage)
	http.HandleFunc("/indexs", server.IndexsPage)

	fmt.Println("Server running on http://localhost:8080 \nTo stop the server press Ctrl+C")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
