package server

import (
	"database/sql"
	"fmt"
	"forum/database"
	"log"
	"net/http"
)

func HomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/home" {
		log.Println("Incorrect path")
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		errHandler(w, r, &err)
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
		log.Println("Database connection failed:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
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
		if err != nil {
			log.Println("Error fetching userid ID from user table:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}
	}

	users, err := database.GetAllUsers(db)
	if err != nil {
		log.Println("Failed to fetch users:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}

	categories, err := database.GetAllCategories(db)
	if err != nil {
		log.Println("Failed to fetch categories:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}

	categoryNames := make([]string, len(categories))
	for i, category := range categories {
		categoryNames[i] = category.Name
	}

	var posts []database.Post

	allPosts, err := database.GetAllPosts(db)
	if err != nil {
		log.Println("Failed to fetch posts:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}

	filter := r.URL.Query().Get("filter")
	selectedTab := r.URL.Query().Get("tab")

	if selectedTab == "" {
		selectedTab = "posts"
	}
	if selectedTab == "" {
		selectedTab = "posts"
	}

	if filter == "" {
		if selectedTab == "your+posts" {
			filter = "newest"
		} else if selectedTab == "your+replies" {
			filter = "newest"
		} else if selectedTab == "your+reactions" {
			filter = "likes"
		} else {
			filter = "all"
		}
	}
	if filter == "" {
		if selectedTab == "your+posts" {
			filter = "newest"
		} else if selectedTab == "your+replies" {
			filter = "newest"
		} else if selectedTab == "your+reactions" {
			filter = "likes"
		} else {
			filter = "all"
		}
	}

	switch selectedTab {
	case "posts":
		switch filter {
		case "all":
			posts = allPosts
		case "top-rated":
			posts, err = database.GetFilteredPosts(db, filter)
			if err != nil {
				log.Println("Failed to fetch posts:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
		case "oldest":
			posts, err = database.GetFilteredPosts(db, filter)
			if err != nil {
				log.Println("Failed to fetch posts:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
		default:
			log.Println("Invalid filter selected", err)
			err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
			errHandler(w, r, &err)
			return
		}
	case "tags":

		if filter == "all" {
			posts = allPosts
		} else if CheckFilter(filter, categoryNames) {
			posts, err = database.GetFilteredPosts(db, filter)
			if err != nil {
				log.Println("Failed to fetch posts:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
		} else {
			log.Println("Invalid filter selected", err)
			err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
			errHandler(w, r, &err)
			return
		}

	case "your+posts":

		switch filter {

		case "newest":
			posts, err = database.GetUserPosts(db, userID, filter)
			if err != nil {
				log.Println("Failed to fetch posts:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
		case "oldest":
			posts, err = database.GetUserPosts(db, userID, filter)
			if err != nil {
				log.Println("Failed to fetch posts:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
		default:
			log.Println("Invalid filter selected", err)
			err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
			errHandler(w, r, &err)
			return
		}

	case "your+replies":

		posts, err = database.GetUserCommentedPosts(db, userID, filter)
		if err != nil {
			log.Println("Failed to fetch posts:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

	case "your+reactions":

		switch filter {

		case "likes":
			//
			posts, err = database.GetUserReaction(db, userID, filter)
			if err != nil {
				log.Println("Failed to fetch posts:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
		case "dislikes":
			//
			posts, err = database.GetUserReaction(db, userID, filter)
			if err != nil {
				log.Println("Failed to fetch posts:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
		default:
			log.Println("Invalid filter selected", err)
			err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
			errHandler(w, r, &err)
			return
		}

	default:
		log.Println("Invalid tab selected", err)
		err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
		errHandler(w, r, &err)
		return
	}

	if hasSession {
		var avatar sql.NullString
		var roleID int
		err = db.QueryRow("SELECT avatar, role_id FROM user WHERE userID = ?", userID).Scan(&avatar, &roleID)
		if err == sql.ErrNoRows {
			log.Println("No user found with the given ID:", userID)
			err := ErrorPageData{Code: "404", ErrorMsg: "USER NOT FOUND"}
			errHandler(w, r, &err)
			return
		} else if err != nil {
			log.Println("Failed to fetch user data:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
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

		var totalLikes, totalPosts int
		err = db.QueryRow("SELECT COUNT(*) FROM likes WHERE user_userID = ?", userID).Scan(&totalLikes)
		if err != nil {
			log.Println("Failed to fetch total likes:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		err = db.QueryRow("SELECT COUNT(*) FROM post WHERE user_userID = ?", userID).Scan(&totalPosts)
		if err != nil {
			log.Println("Failed to fetch total posts:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		notifications, err := database.GetLastNotifications(db, userID)
		if err != nil {
			log.Println("Failed to fetch notifications:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		data := PageData{
			HasSession:     hasSession,
			UserID:         userID,
			UserName:       userName,
			Avatar:         avatar.String,
			RoleName:       roleName,
			TotalLikes:     totalLikes,
			TotalPosts:     totalPosts,
			Categories:     categories,
			Users:          users,
			Posts:          posts,
			SelectedTab:    selectedTab,
			SelectedFilter: filter,
			Notifications:  notifications,
			RoleID:         roleID,
		}

		err = templates.ExecuteTemplate(w, "home.html", data)
		if err != nil {
			log.Println("Error rendering home page:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}
	} else {
		if selectedTab != "posts" && selectedTab != "tags" {
			err = templates.ExecuteTemplate(w, "index.html", nil)
			if err != nil {
				log.Println("Error rendering home page:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
		}

		data := PageData{
			HasSession:     hasSession,
			Categories:     categories,
			Users:          users,
			Posts:          posts,
			SelectedTab:    selectedTab,
			SelectedFilter: filter,
		}

		err = templates.ExecuteTemplate(w, "home.html", data)
		if err != nil {
			log.Println("Error rendering home page:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}
	}
}
