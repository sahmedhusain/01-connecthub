package server

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"forum/database"
	"log"
	"net/http"
	"strconv"
)

func PostPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/post" {
		log.Println("Invalid URL path")
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		ErrHandler(w, r, &err)
		return
	}

	if r.Method != "GET" {
		log.Println("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Error opening database:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Fetch session cookie
	seshCok, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusBadRequest)
		fmt.Println("Error fetching session cookie")
		return
	}

	// Set session token from cookie value
	seshVal := seshCok.Value

	var userID int
	var userName string
	err = db.QueryRow("SELECT userid, Username FROM user WHERE current_session = ?", seshVal).Scan(&userID, &userName)
	if err != nil {
		log.Println("Error fetching userid and username from user table:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &err)
		return
	}

	postID := r.URL.Query().Get("id")
	if postID == "" {
		log.Println("Post ID not found in query parameters")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	var post database.Post
	err = db.QueryRow(`
        SELECT post.postid, post.image, post.title, post.content, post.post_at, post.user_userid, user.Username, user.F_name, user.L_name, user.Avatar,
               (SELECT COUNT(*) FROM likes WHERE likes.post_postid = post.postid) AS Likes,
               (SELECT COUNT(*) FROM dislikes WHERE dislikes.post_postid = post.postid) AS Dislikes,
               (SELECT COUNT(*) FROM comment WHERE comment.post_postid = post.postid) AS Comments
        FROM post
        JOIN user ON post.user_userid = user.userid
        WHERE post.postid = ?
		`, postID).Scan(&post.PostID, &post.Image, &post.Title, &post.Content, &post.PostAt, &post.UserUserID, &post.Username, &post.FirstName, &post.LastName, &post.Avatar, &post.Likes, &post.Dislikes, &post.Comments)
	if err != nil {
		log.Println("Failed to fetch posts")
		errData := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
		ErrHandler(w, r, &errData)
		return
	}

	// Encode the image to base64
	var base64Image string
	if post.Image.Valid {
		base64Image = base64.StdEncoding.EncodeToString([]byte(post.Image.String))
	}

	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		log.Println("Error converting post ID to integer:", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	comments, err := database.GetCommentsForPost(db, postIDInt)
	if err != nil {
		log.Println("Error getting comments for post:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if userID < 0 {
		log.Println("User ID not found in query parameters")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	// Fetch categories for the post
	categories, err := database.GetCategoriesForPost(db, post.PostID)
	if err != nil {
		log.Println("Error fetching categories for post:", err)
		return
	}

	log.Println("User ID:", userID)

	data := PageData{
		Post:        post,
		Comments:    comments,
		UserID:      userID,
		UserName:    userName,
		Categories:  categories,
		ImageBase64: base64Image,
	}

	err = templates.ExecuteTemplate(w, "post.html", data)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
