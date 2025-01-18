package server

import (
	"database/sql"
	"fmt"
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

		// Retrieve username cookie
		usrCok, err := r.Cookie("dotcom_user")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			fmt.Println("Error fetching username from cookie")
			return
		}

		//Set username from cookie value
		userName := usrCok.Value

		var userID int
		err = db.QueryRow("SELECT userid FROM user WHERE Username = ?", userName).Scan(&userID)
		if err != nil {
			log.Println("Error fetching session ID from user table:", err)
			http.Redirect(w, r, "/", http.StatusFound)
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

		log.Printf("Fetched user data: %+v\n", user) // Add this line for debugging

		// Handle case where roleID is 0
		if user.RoleID == 0 {
			user.RoleID = 3 // Assign default role (User)
		}

		roleName, err := database.GetRoleNameByID(db, user.RoleID)
		if err != nil {
			log.Println("Failed to fetch role name")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		log.Printf("Fetched role name: %s\n", roleName) // Add this line for debugging

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
			RoleID        int // Add this line
		}{
			UserID:        userID,
			Categories:    categories,
			Notifications: notifications,
			Avatar:        userAvatar,
			RoleName:      roleName,
			UserName:      user.Username,
			TotalLikes:    totalLikes,
			TotalPosts:    totalPosts,
			SelectedTab:   "posts",     // Default value or set based on your logic
			RoleID:        user.RoleID, // Add this line
		}

		err = templates.ExecuteTemplate(w, "newpost.html", data)
		if err != nil {
			log.Println("Error rendering new post page:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

	case "POST":
		// Parse the form data
		err := r.ParseMultipartForm(10 << 20) // 10 MB max memory
		if err != nil {
			log.Println("Failed to parse form data")
			err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
			errHandler(w, r, &err)
			return
		}

		// Get the post content
		userID := r.FormValue("user")
		content := r.FormValue("content")
		if userID == "" || content == "" {
			log.Println("Invalid form data")
			err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
			errHandler(w, r, &err)
			return
		}

		// Handle file upload
		file, _, err := r.FormFile("image")
		var image sql.NullString
		if err == nil {
			defer file.Close()
			// Process the file and save it, then set the image path
			image.String = "forum/static/uploads" // Update with actual path
			image.Valid = true
		} else {
			image.Valid = false
		}

		// Insert the new post into the database
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

		// Handle post categories
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
