package main

import (
	"fmt"
	// db "forum/database"
	"forum/src/server"
	"log"
	"net/http"
	authentication "forum/src/authentication"

	"github.com/gorilla/context"
)

func init() {
	// db.DataBase()
	// db.DropDataBase()
	// db.CommentReactions()
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// http.HandleFunc("/", server.MainPage)
	http.HandleFunc("/", server.ReverseMiddleware(server.LoginPage))
	http.HandleFunc("/logout", server.AuthMiddleware(server.Logout))
	http.HandleFunc("/signup", server.SignupPage)
	http.HandleFunc("/home", server.HomePage)
	http.HandleFunc("/newpost", server.AuthMiddleware(server.NewPostPage))
	http.HandleFunc("/settings", server.AuthMiddleware(server.SettingsPage))
	http.HandleFunc("/notifications", server.AuthMiddleware(server.NotificationsPage))
	http.HandleFunc("/myprofile", server.AuthMiddleware(server.MyProfilePage))
	http.HandleFunc("/profile", server.AuthMiddleware(server.ProfilePage))
	http.HandleFunc("/admin", server.AdminPage)
	http.HandleFunc("/moderator", server.ModeratorPage)
	http.HandleFunc("/post", server.AuthMiddleware(server.PostPage))
	http.HandleFunc("/like", server.AuthMiddleware(server.LikePost))
	http.HandleFunc("/dislike", server.AuthMiddleware(server.DislikePost))
	http.HandleFunc("/commentlike", server.AuthMiddleware(server.LikeComment))
	http.HandleFunc("/commentdislike", server.AuthMiddleware(server.DislikeComment))
	http.HandleFunc("/deletepost", server.AuthMiddleware(server.DeletePost))
	http.HandleFunc("/reportpost", server.AuthMiddleware(server.ReportPost))
	http.HandleFunc("/changepassword", server.AuthMiddleware(server.ChangePassword))
	// http.HandleFunc("/togglepassword", server.AuthMiddleware(server.TogglePassword))
	http.HandleFunc("/addcomment", server.AuthMiddleware(server.AddComment))

	fmt.Println("Server running on http://localhost:8080\nTo stop the server press Ctrl+C")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
