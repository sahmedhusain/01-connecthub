package main

import (
	"fmt"
	"forum/database"
	"log"
	"net/http"
)

func main() {

	database.DataBase()
	fmt.Println("Tables created successfully")

	fmt.Print("Server running on http://localhost:8080 \nTo stop the server press Ctrl+C\n")

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
