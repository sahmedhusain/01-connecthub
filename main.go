package main

import (
	"fmt"
	db "forum/database"
	"forum/src/server"
	"log"
	"net/http"
)

func init() {
	db.DataBase()
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	http.HandleFunc("/", server.MainPage)
	http.HandleFunc("/login", server.LoginPage)
	http.HandleFunc("/signup", server.SignupPage)
	http.HandleFunc("/home", server.HomePage)

	fmt.Println("Server running on http://localhost:8080 \nTo stop the server press Ctrl+C")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
