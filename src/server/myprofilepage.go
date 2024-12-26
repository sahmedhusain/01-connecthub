package server

import (
	"database/sql"
	"forum/database"
	"net/http"
)

func MyProfilePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/myprofile" {
		http.NotFound(w, r)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Retrieve UserID from session
	session, _ := store.Get(r, "session-name")
	userID, ok := session.Values["userID"].(string)
	if !ok || userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var user database.User
	err = db.QueryRow("SELECT userid, F_name, L_name, Username, Avatar FROM user WHERE userid = ?", userID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Avatar)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	posts, err := database.GetUserPosts(db, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	followersCount, err := database.GetFollowersCount(db, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	followingCount, err := database.GetFollowingCount(db, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	friendsCount, err := database.GetFriendsCount(db, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view := r.URL.Query().Get("view")
	var followers, following, friends []database.User

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
	} else if view == "friends" {
		friends, err = database.GetFriends(db, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	data := struct {
		UserID         string
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
	}

	err = templates.ExecuteTemplate(w, "myprofile.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
