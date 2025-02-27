package main

import (
	"fmt"
	db "forum/database"
	auth "forum/src/authentication"
	"forum/src/server"
	"log"
	"net/http"
)

func init() {
	db.DataBase()
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	http.HandleFunc("/", server.ReverseMiddleware(server.LoginPage))
	http.HandleFunc("/logout", server.AuthMiddleware(server.Logout))
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
	http.HandleFunc("/commentlike", server.AuthMiddleware(server.LikeComment))
	http.HandleFunc("/commentdislike", server.AuthMiddleware(server.DislikeComment))
	http.HandleFunc("/deletepost", server.AuthMiddleware(server.DeletePost))
	http.HandleFunc("/reportpost", server.AuthMiddleware(server.ReportPost))
	http.HandleFunc("/changepassword", server.AuthMiddleware(server.ChangePassword))
	// http.HandleFunc("/togglepassword", server.AuthMiddleware(server.TogglePassword))
	http.HandleFunc("/addcomment", server.AuthMiddleware(server.AddComment))
	http.HandleFunc("/callbackGoogle", auth.CallbackGoogle)
	http.HandleFunc("/auth/google", auth.LoginPageGoogle)
	http.HandleFunc("/callback", auth.Callback)
	http.HandleFunc("/auth/github", auth.LoginPageGit)
	http.HandleFunc("/search", server.AuthMiddleware(server.SearchHandler))
	http.HandleFunc("/searchpage", server.AuthMiddleware(server.SearchPageHandler))

	fmt.Println("Server running on http://localhost:8080\nTo stop the server press Ctrl+C")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
