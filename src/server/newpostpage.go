package server

import (
	"database/sql"
	"forum/database"
	"log"
	"net/http"
	"strconv"
)

func NewPostPage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		db, err := sql.Open("sqlite3", "./database/main.db")
		if err != nil {
			log.Println("Database connection failed")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}
		defer db.Close()

		categories, err := database.GetAllCategories(db)
		if err != nil {
			log.Println("Failed to fetch categories")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		notifications, err := database.GetLastNotifications(db, r.FormValue("user"))
		if err != nil {
			log.Println("Failed to fetch notifications")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		userID := r.FormValue("user")
		user, err := database.GetUserByID(db, userID)
		if err != nil {
			log.Println("Failed to fetch user data")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}
		userAvatar := user.Avatar.String

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
			UserID        string
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
			UserName:      user.Username,
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
		content := r.FormValue("content")
		if userID == "" || content == "" {
			log.Println("Invalid form data")
			err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
			errHandler(w, r, &err)
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

		db, err := sql.Open("sqlite3", "./database/main.db")
		if err != nil {
			log.Println("Database connection failed")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}
		defer db.Close()

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

		http.Redirect(w, r, "/home?user="+userID+"&tab=posts&filter=all", http.StatusSeeOther)
	}
}
