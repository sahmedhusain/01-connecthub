package server

import (
	"database/sql"
	"forum/database"
	"log"
	"net/http"
	"strconv"
)

func PostPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/post" {
		log.Println("Redirecting to Home page")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	if r.Method != "GET" {
		log.Println("Method not allowed")
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		errHandler(w, r, &err)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Error opening database:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}
	defer db.Close()

	postID := r.URL.Query().Get("id")
	if postID == "" {
		log.Println("Post ID not found in query parameters")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	var post database.Post
	err = db.QueryRow(`
        SELECT post.postid, post.image, post.content, post.post_at, post.user_userid, user.Username, user.F_name, user.L_name, user.Avatar,
               (SELECT COUNT(*) FROM likes WHERE likes.post_postid = post.postid) AS Likes,
               (SELECT COUNT(*) FROM dislikes WHERE dislikes.post_postid = post.postid) AS Dislikes,
               (SELECT COUNT(*) FROM comment WHERE comment.post_postid = post.postid) AS Comments
        FROM post
        JOIN user ON post.user_userid = user.userid
        WHERE post.postid = ?
    `, postID).Scan(&post.PostID, &post.Image, &post.Content, &post.PostAt, &post.UserUserID, &post.Username, &post.FirstName, &post.LastName, &post.Avatar, &post.Likes, &post.Dislikes, &post.Comments)
	if err != nil {
		log.Println("Error querying post:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}

	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		log.Println("Error converting post ID to integer:", err)
		err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
		errHandler(w, r, &err)
		return
	}
	comments, err := database.GetCommentsForPost(db, postIDInt)
	if err != nil {
		log.Println("Error getting comments for post:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}

	data := PageData{
		Post:     post,
		Comments: comments,
	}

	err = templates.ExecuteTemplate(w, "post.html", data)
	if err != nil {
		log.Println("Error executing template:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}
}
