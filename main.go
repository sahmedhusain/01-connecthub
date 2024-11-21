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
   http.Handle("/Static/", http.StripPrefix("/Static/", http.FileServer(http.Dir("Static"))))
   http.HandleFunc("/", server.MainPage)

	
   fmt.Println("Server is listening on port 8080")
   log.Fatal(http.ListenAndServe(":8080", nil))
}   
