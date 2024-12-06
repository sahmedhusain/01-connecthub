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
    http.HandleFunc("/about", server.AboutPage)
    http.HandleFunc("/help", server.HelpPage)
    http.HandleFunc("/privacy_policy", server.PrivacyPolicyPage)
    http.HandleFunc("/activity_centre", server.ActivityCentrePage)
    http.HandleFunc("/connections", server.ConnectionsPage)
    http.HandleFunc("/content_policy", server.ContentPolicyPage)
    http.HandleFunc("/login", server.LoginPage)
    http.HandleFunc("/notifications", server.NotificationsPage)
    http.HandleFunc("/profile", server.ProfilePage)
    http.HandleFunc("/signup", server.SignupPage)
    http.HandleFunc("/user_agreement", server.UserAgreementPage)
    http.HandleFunc("/indexs", server.IndexsPage)

	fmt.Println("Server running on http://localhost:8080 \nTo stop the server press Ctrl+C")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
