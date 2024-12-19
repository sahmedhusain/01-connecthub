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
    http.HandleFunc("/newpost", server.NewPostPage)
    http.HandleFunc("/settings", server.SettingsPage)
    http.HandleFunc("/notifications", server.NotificationsPage)
    http.HandleFunc("/myprofile", server.MyProfilePage)
    http.HandleFunc("/profile", server.ProfilePage)
    http.HandleFunc("/admin", server.AdminPage)
    http.HandleFunc("/moderator", server.ModeratorPage)
    http.HandleFunc("/post", server.PostPage)
    http.HandleFunc("/like", server.LikePost)
    http.HandleFunc("/dislike", server.DislikePost)

    fmt.Println("Server running on http://localhost:8080\nTo stop the server press Ctrl+C")

    log.Fatal(http.ListenAndServe(":8080", nil))
}
