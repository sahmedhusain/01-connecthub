package server

import (
	"database/sql"
	"fmt"
	"forum/database"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const maxPostLength = 500

func NewPostPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/newpost" {
		log.Println("Invalid URL path")
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		errHandler(w, r, &err)
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Database connection failed")
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
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
		errHandler(w, r, &err)
		return
	}

	switch r.Method {
	case "GET":
		categories, err := database.GetAllCategories(db)
		if err != nil {
			log.Println("Failed to fetch categories")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		user, err := database.GetUserByID(db, userID)
		if err != nil {
			log.Println("Failed to fetch user data")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		notifications, err := database.GetLastNotifications(db, userID)
		if err != nil {
			log.Println("Failed to fetch notifications")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		userAvatar := user.Avatar.String // Assuming Avatar is of type sql.NullString

		log.Printf("Fetched user data: %+v\n", user)

		if user.RoleID == 0 {
			user.RoleID = 3
		}

		roleName, err := database.GetRoleNameByID(db, user.RoleID)
		if err != nil {
			log.Println("Failed to fetch role name")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		log.Printf("Fetched role name: %s\n", roleName)

		totalLikes, err := database.GetTotalLikes(db, userID)
		if err != nil {
			log.Println("Failed to fetch total likes")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		totalPosts, err := database.GetTotalPosts(db, userID)
		if err != nil {
			log.Println("Failed to fetch total posts")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		data := struct {
			UserID        int
			Categories    []database.Category
			Notifications []database.Notification
			Avatar        string
			RoleName      string
			UserName      string
			TotalLikes    int
			TotalPosts    int
			SelectedTab   string
			RoleID        int
		}{
			UserID:        userID,
			Categories:    categories,
			Notifications: notifications,
			Avatar:        userAvatar,
			RoleName:      roleName,
			UserName:      userName,
			TotalLikes:    totalLikes,
			TotalPosts:    totalPosts,
			SelectedTab:   "posts",
			RoleID:        user.RoleID,
		}

		err = templates.ExecuteTemplate(w, "newpost.html", data)
		if err != nil {
			log.Println("Error rendering new post page:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

	case "POST":

		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			log.Println("Failed to parse form data")
			err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
			errHandler(w, r, &err)
			return
		}

		userID := r.FormValue("user")
		content := strings.TrimSpace(r.FormValue("content"))
		fmt.Println(userID, content)
		if userID == "" || content == "" {
			log.Println("Invalid form data")
			err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
			errHandler(w, r, &err)
			return
		}

		if len(content) > maxPostLength {
			http.Error(w, "Post content exceeds the character limit", http.StatusBadRequest)
			return
		}

		if len(content) > maxPostLength {
			http.Error(w, "Post content exceeds the character limit", http.StatusBadRequest)
			return
		}

		file, _, err := r.FormFile("image")
		var image sql.NullString
		if err == nil {
			defer file.Close()

			image.String = "forum/static/uploads"
			image.Valid = true
		} else {
			image.Valid = false
		}

		postID, err := database.InsertPost(db, content, image, userID)
		if err != nil {
			log.Println("Failed to insert new post")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		categories := r.Form["categories"]
		for _, categoryID := range categories {
			categoryIDInt, err := strconv.Atoi(categoryID)
			if err != nil {
				log.Println("Invalid category ID")
				continue
			}
			err = database.InsertPostCategory(db, postID, categoryIDInt)
			if err != nil {
				log.Println("Failed to insert post category")
			}
		}

		// Redirect to the home page after successful post
		http.Redirect(w, r, "/home?tab=posts&filter=all", http.StatusSeeOther)
	}
}
