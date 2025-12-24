package server

import (
	"01connecthub/database"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const maxPostLength = 500
const maxFileSize = 20 << 20

func NewPostPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/newpost" {
		log.Println("Invalid URL path")
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		ErrHandler(w, r, &err)
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Database connection failed")
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &err)
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
	if err == sql.ErrNoRows {
		log.Println("No user found with the given user ID:", userID)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	} else if err != nil {
		log.Println("User is not an admin or error occurred")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if hasSession {
		var avatar sql.NullString
		err = db.QueryRow("SELECT avatar, role_id FROM user WHERE userid = ?", userID).Scan(&avatar, &roleID)
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

		switch r.Method {
		case "GET":
			categories, err := database.GetAllCategories(db)
			if err != nil {
				log.Println("Failed to fetch categories")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}

			user, err := database.GetUserByID(db, userID)
			if err != nil {
				log.Println("Failed to fetch user data")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}

			notifications, err := database.GetLastNotifications(db, userID)
			if err != nil {
				log.Println("Failed to fetch notifications")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
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
				ErrHandler(w, r, &err)
				return
			}

			log.Printf("Fetched role name: %s\n", roleName)

			totalLikes, err := database.GetTotalLikes(db, userID)
			if err != nil {
				log.Println("Failed to fetch total likes")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}

			totalPosts, err := database.GetTotalPosts(db, userID)
			if err != nil {
				log.Println("Failed to fetch total posts")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
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
				HasSession    bool
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
				HasSession:    hasSession,
			}

			err = templates.ExecuteTemplate(w, "newpost.html", data)
			if err != nil {
				log.Println("Error rendering new post page:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}

		case "POST":

			err := r.ParseMultipartForm(10 << 20)
			if err != nil {
				log.Println("Failed to parse form data")
				err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
				ErrHandler(w, r, &err)
				return
			}

			userID := r.FormValue("user")
			content := strings.TrimSpace(r.FormValue("content"))
			title := strings.TrimSpace(r.FormValue("title"))
			fmt.Println(userID, content)
			if userID == "" || content == "" {
				log.Println("Invalid form data")
				err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
				ErrHandler(w, r, &err)
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

			// Handle file upload
			file, fileHeader, err := r.FormFile("image")
			if err != nil && err != http.ErrMissingFile {
				log.Println("Error getting uploaded file")
				err := ErrorPageData{Code: "500", ErrorMsg: "Error getting uploaded file"}
				ErrHandler(w, r, &err)
				return
			}

			var imageData []byte
			if file != nil {
				defer file.Close()

				if fileHeader.Size > maxFileSize {
					err := ErrorPageData{Code: "500", ErrorMsg: "Image size exceeds 20 MB limit"}
					ErrHandler(w, r, &err)
					return
				}

				imageData, err = io.ReadAll(file)
				if err != nil {
					err := ErrorPageData{Code: "500", ErrorMsg: "Error reading uploaded file"}
					ErrHandler(w, r, &err)
					return
				}
			}

			postID, err := database.InsertPost(db, content, title, imageData, userID)
			if err != nil {
				log.Println("Failed to insert new post")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
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

			http.Redirect(w, r, "/home?tab=posts&filter=all", http.StatusSeeOther)
		}
	}
}
