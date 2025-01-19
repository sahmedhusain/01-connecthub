package server

import (
	"database/sql"
	"forum/database"
	"log"
	"net/http"
	"strconv"
)

func ProfilePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/profile" {
		log.Println("Invalid URL path")
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		errHandler(w, r, &err)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Database connection failed")
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}
	defer db.Close()

	// usrCok, err := r.Cookie("dotcom_user")
	// if err != nil {
	// 	http.Redirect(w, r, "/", http.StatusFound)
	// 	fmt.Println("Error fetching username from cookie")
	// 	return
	// }

	// var userID string
	// err = db.QueryRow("SELECT userid FROM user WHERE Username = ?", usrCok.Value).Scan(&userID)
	// if err != nil {
	// 	log.Println("Error fetching session ID:", err)
	// 	http.Redirect(w, r, "/", http.StatusFound)
	// 	return
	// }

	// Retrieve userID from url parameters
	userID, err := strconv.Atoi(r.FormValue("user"))
	if err != nil {
		log.Println("Error converting userID to int:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &err)
		return
	}
	var user database.User
	err = db.QueryRow("SELECT userid, F_name, L_name, Username, Avatar FROM user WHERE userid = ?", userID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Avatar)
	if err != nil {
		log.Println("Failed to fetch user data")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &errData)
		return
	}

	posts, err := database.GetUserPosts(db, userID, "newest")
	if err != nil {
		log.Println("Failed to fetch user posts")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &errData)
		return
	}

	followersCount, err := database.GetFollowersCount(db, userID)
	if err != nil {
		log.Println("Failed to fetch followers count")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &errData)
		return
	}

	followingCount, err := database.GetFollowingCount(db, userID)
	if err != nil {
		log.Println("Failed to fetch following count")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &errData)
		return
	}

	friendsCount, err := database.GetFriendsCount(db, userID)
	if err != nil {
		log.Println("Failed to fetch friends count")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &errData)
		return
	}

	isFollowing, err := database.IsFollowing(db, userID, userID)
	if err != nil {
		log.Println("Failed to check if user is following")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &errData)
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
		errHandler(w, r, &err)
		return
	}
}
