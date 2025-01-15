package server

import (
    "database/sql"
    "forum/database"
    "log"
    "net/http"
)

func MyProfilePage(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/myprofile" {
        http.NotFound(w, r)
        return
    }

    db, err := sql.Open("sqlite3", "./database/main.db")
    if err != nil {
        log.Println("Error opening database:", err)
        err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
        errHandler(w, r, &err)
        return
    }
    defer db.Close()

    userID := r.URL.Query().Get("user")
    if userID == "" {
        http.Redirect(w, r, "/", http.StatusSeeOther)
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
        errHandler(w, r, &err)
        return
    }

    posts, err := database.GetUserPosts(db, userID, "newest")
    if err != nil {
        log.Println("Error fetching user posts:", err)
        err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
        errHandler(w, r, &err)
        return
    }

    followersCount, err := database.GetFollowersCount(db, userID)
    if err != nil {
        log.Println("Error fetching followers count:", err)
        err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
        errHandler(w, r, &err)
        return
    }

    followingCount, err := database.GetFollowingCount(db, userID)
    if err != nil {
        log.Println("Error fetching following count:", err)
        err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
        errHandler(w, r, &err)
        return
    }

    friendsCount, err := database.GetFriendsCount(db, userID)
    if err != nil {
        log.Println("Error fetching friends count:", err)
        err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
        errHandler(w, r, &err)
        return
    }

    notifications, err := database.GetLastNotifications(db, userID)
    if err != nil {
        log.Println("Error fetching notifications:", err)
        err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
        errHandler(w, r, &err)
        return
    }

    totalLikes, err := database.GetTotalLikes(db, userID)
    if err != nil {
        log.Println("Error fetching total likes:", err)
        err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
        errHandler(w, r, &err)
        return
    }

    totalPosts, err := database.GetTotalPosts(db, userID)
    if err != nil {
        log.Println("Error fetching total posts:", err)
        err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
        errHandler(w, r, &err)
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
            errHandler(w, r, &err)
            return
        }
    } else if view == "following" {
        following, err = database.GetFollowing(db, userID)
        if err != nil {
            log.Println("Error fetching following:", err)
            err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
            errHandler(w, r, &err)
            return
        }
    } else if view == "friends" {
        friends, err = database.GetFriends(db, userID)
        if err != nil {
            log.Println("Error fetching friends:", err)
            err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
            errHandler(w, r, &err)
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
        Notifications  []database.Notification
        RoleName       string
        TotalLikes     int
        TotalPosts     int
        SelectedTab    string
        SelectedFilter string
        RoleID         int
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
        SelectedTab:    "your+posts", // Set the default selected tab
        SelectedFilter: "newest",     // Set the default selected filter
        RoleID:         user.RoleID,
    }

    err = templates.ExecuteTemplate(w, "myprofile.html", data)
    if err != nil {
        log.Println("Error rendering my profile page:", err)
        err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
        errHandler(w, r, &err)
    }
}
