package server

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"forum/database"
	"log"
	"net/http"
	"strconv"
	"time"
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

	var hasSession bool
	var userID int
	var userName string
	seshCok, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println("No cookie found, treated as guest")
	} else if seshCok.Value == "" {
		hasSession = false
	} else {
		hasSession = true

		seshVal := seshCok.Value

		err = db.QueryRow("SELECT userid, Username FROM user WHERE current_session = ?", seshVal).Scan(&userID, &userName)
		if err == sql.ErrNoRows {
			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    "",
				Expires:  time.Now().Add(-time.Hour),
				HttpOnly: true,
			})
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else if err != nil {
			log.Println("Error fetching userid ID from user table:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		} else {
			log.Println("User is logged in:", userName)
		}
	}

	var roleID int
	err = db.QueryRow("SELECT role_id FROM user WHERE userid = ?", userID).Scan(&roleID)
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &err)
		return
	}

	var roleName string
	if roleID == 1 {
		roleName = "Admin"
	} else if roleID == 2 {
		roleName = "Moderator"
	} else {
		roleName = "User"
	}

	if hasSession {
		var avatar sql.NullString
		err = db.QueryRow("SELECT avatar, role_id FROM user WHERE userID = ?", userID).Scan(&avatar, &roleID)
		if err == sql.ErrNoRows {
			log.Println("No user found with the given ID:", userID)
			err := ErrorPageData{Code: "404", ErrorMsg: "USER NOT FOUND"}
			ErrHandler(w, r, &err)
			return
		} else if err != nil {
			log.Println("Failed to fetch user data:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}

		var userName string
		seshVal := seshCok.Value
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

		categories, err := database.GetCategoriesForPost(db, post.PostID)
		if err != nil {
			log.Println("Error fetching categories for post:", err)
			return
		}

		log.Println("User ID:", userID)
		user, err := database.GetUserByID(db, userID)
		if err != nil {
			log.Println("Failed to fetch user data")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}

		userAvatar := user.Avatar.String

		data := PageData{
			RoleName:    roleName,
			HasSession:  hasSession,
			Post:        post,
			Comments:    comments,
			UserID:      userID,
			UserName:    userName,
			Categories:  categories,
			ImageBase64: base64Image,
			Avatar:      userAvatar,
		}

		err = templates.ExecuteTemplate(w, "post.html", data)
		if err != nil {
			log.Println("Error executing template:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}
