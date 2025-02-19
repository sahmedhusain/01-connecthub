package server

import (
	"database/sql"
	"fmt"
	"forum/database"
	"log"
	"net/http"
	"strconv"
	"time"
)

func ProfilePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/profile" {
		log.Println("Invalid URL path")
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		ErrHandler(w, r, &err)
		return
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

	//must check if user is a moderator!
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

	// Retrieve userID from url parameters
	userID, err := strconv.Atoi(r.FormValue("user"))
	if err != nil {
		log.Println("Error converting userID to int:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &err)
		return
	}
	var user database.User
	err = db.QueryRow("SELECT userid, F_name, L_name, Username, Avatar FROM user WHERE userid = ?", userID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Avatar)
	if err != nil {
		log.Println("Failed to fetch user data")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &errData)
		return
	}

	posts, err := database.GetUserPosts(db, userID, "newest")
	if err != nil {
		log.Println("Failed to fetch user posts")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &errData)
		return
	}

	followersCount, err := database.GetFollowersCount(db, userID)
	if err != nil {
		log.Println("Failed to fetch followers count")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &errData)
		return
	}

	followingCount, err := database.GetFollowingCount(db, userID)
	if err != nil {
		log.Println("Failed to fetch following count")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &errData)
		return
	}

	friendsCount, err := database.GetFriendsCount(db, userID)
	if err != nil {
		log.Println("Failed to fetch friends count")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &errData)
		return
	}

	isFollowing, err := database.IsFollowing(db, userID, userID)
	if err != nil {
		log.Println("Failed to check if user is following")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &errData)
		return
	}

	view := r.URL.Query().Get("view")
	var followers, following []database.User

	if view == "followers" {
		followers, err = database.GetFollowers(db, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if view == "following" {
		following, err = database.GetFollowing(db, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	data := struct {
		UserID int
		Hassession bool
		RoleName string
		// Avatar                string
		// userID                int
		ProfileFirstName      string
		ProfileLastName       string
		ProfileUsername       string
		ProfileAvatar         string
		ProfilePostsCount     int
		ProfileFollowersCount int
		ProfileFollowingCount int
		ProfileFriendsCount   int
		ProfilePosts          []database.Post
		IsFollowing           bool
		View                  string
		Followers             []database.User
		Following             []database.User
	}{
		UserID: userID,
		Hassession: hasSession,
		RoleName: roleName,
		// Avatar:                session.Values["avatar"].(string),
		// userID:                userID,
		ProfileFirstName:      user.FirstName,
		ProfileLastName:       user.LastName,
		ProfileUsername:       user.Username,
		ProfileAvatar:         user.Avatar.String,
		ProfilePostsCount:     len(posts),
		ProfileFollowersCount: followersCount,
		ProfileFollowingCount: followingCount,
		ProfileFriendsCount:   friendsCount,
		ProfilePosts:          posts,
		IsFollowing:           isFollowing,
		View:                  view,
		Followers:             followers,
		Following:             following,
	}

	err = templates.ExecuteTemplate(w, "profile.html", data)
	if err != nil {
		log.Println("Error rendering profile page:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &err)
		return
	}
}
}