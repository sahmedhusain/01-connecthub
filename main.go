package main

import (
	"github.com/gorilla/context"
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
	http.HandleFunc("/newpost", server.AuthMiddleware(server.NewPostPage))
	http.HandleFunc("/settings", server.AuthMiddleware(server.SettingsPage))
	http.HandleFunc("/notifications", server.AuthMiddleware(server.NotificationsPage))
	http.HandleFunc("/myprofile", server.AuthMiddleware(server.MyProfilePage))
	http.HandleFunc("/profile", server.AuthMiddleware(server.ProfilePage))
	http.HandleFunc("/admin", server.AuthMiddleware(server.AdminPage))
	http.HandleFunc("/moderator", server.AuthMiddleware(server.ModeratorPage))
	http.HandleFunc("/post", server.AuthMiddleware(server.PostPage))
	http.HandleFunc("/like", server.AuthMiddleware(server.LikePost))
	http.HandleFunc("/dislike", server.AuthMiddleware(server.DislikePost))
	http.HandleFunc("/deletepost", server.AuthMiddleware(server.DeletePost))
	http.HandleFunc("/reportpost", server.AuthMiddleware(server.ReportPost))
	http.HandleFunc("/changepassword", server.AuthMiddleware(server.ChangePassword))
	http.HandleFunc("/togglepassword", server.AuthMiddleware(server.TogglePassword))
	http.HandleFunc("/addcomment", server.AuthMiddleware(server.AddComment))
	http.HandleFunc("/logout", server.Logout)

    fmt.Println("Server running on http://localhost:8080\nTo stop the server press Ctrl+C")

    log.Fatal(http.ListenAndServe(":8080", context.ClearHandler(http.DefaultServeMux)))
}
