package server

import (
	"database/sql"
	"fmt"
	"forum/database"
	"log"
	"net/http"
	"time"
)

func ModeratorPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/moderator" {
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

		switch r.Method {
		case "GET":

			posts, err := database.GetAllPosts(db)
			if err != nil {
				log.Println("Failed to fetch posts:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}

			comments, err := database.GetComments(db)
			if err != nil {
				log.Println("Failed to fetch comments:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}

			var totalLikes, totalPosts int
			err = db.QueryRow("SELECT COUNT(*) FROM likes WHERE user_userID = ?", userID).Scan(&totalLikes)
			if err != nil {
				log.Println("Failed to fetch total likes:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}

			err = db.QueryRow("SELECT COUNT(*) FROM post WHERE user_userID = ?", userID).Scan(&totalPosts)
			if err != nil {
				log.Println("Failed to fetch total posts:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}

			data := PageData{
				HasSession:    hasSession,
				UserID:        userID,
				UserName:      userName,
				RoleName:      roleName,
				RoleID:        roleID,
				Posts:         posts,
				Comments:      comments,
				TotalPosts:    totalPosts,
				TotalLikes:    totalLikes,
				SelectedTab:   "moderator",
				Notifications: []database.Notification{},
			}

			err = templates.ExecuteTemplate(w, "moderator.html", data)
			if err != nil {
				log.Println("Error rendering moderator page:", err)
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				ErrHandler(w, r, &err)
				return
			}
		case "POST":
			r.ParseForm()
			if r.FormValue("delete_post") != "" {
				postID := r.FormValue("delete_post")
				_, err := db.Exec("DELETE FROM post WHERE postid = ?", postID)
				if err != nil {
					log.Println("Failed to delete post")
					err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
					ErrHandler(w, r, &err)
					return
				}
			} else if r.FormValue("report_post") != "" {
				postID := r.FormValue("report_post")
				reportReason := r.FormValue("report_reason")
				_, err := db.Exec("INSERT INTO reports (post_id, reported_by, report_reason) VALUES (?, ?, ?)", postID, userID, reportReason)
				if err != nil {
					log.Println("Failed to report post")
					err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
					ErrHandler(w, r, &err)
					return
				}
			} else if r.FormValue("delete_comment") != "" {
				commentID := r.FormValue("delete_comment")
				_, err := db.Exec("DELETE FROM comment WHERE commentid = ?", commentID)
				if err != nil {
					log.Println("Failed to delete comment")
					err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
					ErrHandler(w, r, &err)
					return
				}
			} else if r.FormValue("report_comment") != "" {
				commentID := r.FormValue("report_comment")
				_, err := db.Exec("INSERT INTO reports (comment_id, reported_by, report_reason) VALUES (?, ?, ?)", commentID, userID, "Reported by moderator")
				if err != nil {
					log.Println("Failed to report comment")
					err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
					ErrHandler(w, r, &err)
					return
				}
			}
			log.Println("Moderator action completed")
			http.Redirect(w, r, "/moderator", http.StatusSeeOther)
		default:
			log.Println("Method not allowed")
			err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
			ErrHandler(w, r, &err)
			return
		}
	}
}
