package server

import (
	"database/sql"
	"forum/database"
	"log"
	"net/http"
)

func ProfilePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/profile" {
		log.Println("Redirecting to Home page")
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

	// Retrieve UserID from session
	session, _ := store.Get(r, "session-name")
	userID, ok := session.Values["userID"].(string)
	if !ok || userID == "" {
		log.Println("UserID not found in session, redirecting to login page")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Retrieve ProfileUserID from query parameters
	profileUserID := r.URL.Query().Get("user")
	if profileUserID == "" {
		log.Println("ProfileUserID not found in query parameters")
		err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
		errHandler(w, r, &err)
		return
	}

	var user database.User
	err = db.QueryRow("SELECT userid, F_name, L_name, Username, Avatar FROM user WHERE userid = ?", profileUserID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Avatar)
	if err != nil {
		log.Println("Failed to fetch user data")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &errData)
		return
	}

	posts, err := database.GetUserPosts(db, profileUserID)
	if err != nil {
		log.Println("Failed to fetch user posts")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &errData)
		return
	}

	followersCount, err := database.GetFollowersCount(db, profileUserID)
	if err != nil {
		log.Println("Failed to fetch followers count")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &errData)
		return
	}

	followingCount, err := database.GetFollowingCount(db, profileUserID)
	if err != nil {
		log.Println("Failed to fetch following count")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &errData)
		return
	}

	friendsCount, err := database.GetFriendsCount(db, profileUserID)
	if err != nil {
		log.Println("Failed to fetch friends count")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &errData)
		return
	}

	isFollowing, err := database.IsFollowing(db, userID, profileUserID)
	if err != nil {
		log.Println("Failed to check if user is following")
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &errData)
		return
	}

	view := r.URL.Query().Get("view")
	var followers, following []database.User

	if view == "followers" {
		followers, err = database.GetFollowers(db, profileUserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if view == "following" {
		following, err = database.GetFollowing(db, profileUserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	data := struct {
		UserID                string
		Avatar                string
		ProfileUserID         string
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
		UserID:                userID,
		Avatar:                session.Values["avatar"].(string),
		ProfileUserID:         profileUserID,
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
