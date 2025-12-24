package server

import (
	"01connecthub/database"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"
)

func MyProfilePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/myprofile" {
		log.Println("Invalid URL path")
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		ErrHandler(w, r, &err)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Error opening database:", err)
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
	if err != nil {
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &err)
		return
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

		var user database.User
		err = db.QueryRow("SELECT userid, F_name, L_name, Username, Avatar, role_id FROM user WHERE userid = ?", userID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Avatar, &user.RoleID)
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		} else if err != nil {
			log.Println("Error fetching user details:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}

		posts, err := database.GetUserPosts(db, userID, "newest")
		if err != nil {
			log.Println("Error fetching user posts:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}

		followersCount, err := database.GetFollowersCount(db, userID)
		if err != nil {
			log.Println("Error fetching followers count:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}

		followingCount, err := database.GetFollowingCount(db, userID)
		if err != nil {
			log.Println("Error fetching following count:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}

		friendsCount, err := database.GetFriendsCount(db, userID)
		if err != nil {
			log.Println("Error fetching friends count:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}

		notifications, err := database.GetLastNotifications(db, userID)
		if err != nil {
			log.Println("Error fetching notifications:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}

		totalLikes, err := database.GetTotalLikes(db, userID)
		if err != nil {
			log.Println("Error fetching total likes:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}

		totalPosts, err := database.GetTotalPosts(db, userID)
		if err != nil {
			log.Println("Error fetching total posts:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}

		var roleName string
		if user.RoleID == 1 {
			roleName = "Admin"
		} else if user.RoleID == 2 {
			roleName = "Moderator"
		} else {
			roleName = "User"
		}

		view := r.URL.Query().Get("view")
		var followers, following, friends []database.User

		if view == "followers" {
			followers, err = database.GetFollowers(db, userID)
			if err != nil {
				log.Println("Error fetching followers:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}
		} else if view == "following" {
			following, err = database.GetFollowing(db, userID)
			if err != nil {
				log.Println("Error fetching following:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}
		} else if view == "friends" {
			friends, err = database.GetFriends(db, userID)
			if err != nil {
				log.Println("Error fetching friends:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}
		}
		if view == "followers" {
			followers, err = database.GetFollowers(db, userID)
			if err != nil {
				log.Println("Error fetching followers:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}
		} else if view == "following" {
			following, err = database.GetFollowing(db, userID)
			if err != nil {
				log.Println("Error fetching following:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}
		} else if view == "friends" {
			friends, err = database.GetFriends(db, userID)
			if err != nil {
				log.Println("Error fetching friends:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}
		}

		data := struct {
			UserID         int
			FirstName      string
			LastName       string
			Username       string
			Avatar         string
			PostsCount     int
			FollowersCount int
			FollowingCount int
			FriendsCount   int
			Posts          []database.Post
			View           string
			Followers      []database.User
			Following      []database.User
			Friends        []database.User
			Notifications  []database.Notification
			RoleName       string
			TotalLikes     int
			TotalPosts     int
			SelectedTab    string
			SelectedFilter string
			RoleID         int
			HasSession     bool
		}{
			UserID:         userID,
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			Username:       user.Username,
			Avatar:         user.Avatar.String,
			PostsCount:     len(posts),
			FollowersCount: followersCount,
			FollowingCount: followingCount,
			FriendsCount:   friendsCount,
			Posts:          posts,
			View:           view,
			Followers:      followers,
			Following:      following,
			Friends:        friends,
			Notifications:  notifications,
			RoleName:       roleName,
			TotalLikes:     totalLikes,
			TotalPosts:     totalPosts,
			SelectedTab:    "your+posts",
			RoleID:         user.RoleID,
			HasSession:     hasSession,
		}

		err = templates.ExecuteTemplate(w, "myprofile.html", data)
		if err != nil {
			log.Println("Error rendering my profile page:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
		}
	}
}
