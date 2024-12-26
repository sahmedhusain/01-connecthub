package server

import (
	"database/sql"
	"forum/database"
	"log"
	"net/http"
)

func ModeratorPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/moderator" {
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
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check if the user is a moderator
	var roleID int
	err = db.QueryRow("SELECT role_id FROM user WHERE id = ?", userID).Scan(&roleID)
	if err != nil || roleID != 2 {
		log.Println("User is not a moderator, redirecting to Home page")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	switch r.Method {
	case "GET":
		posts, err := database.GetAllPosts(db)
		if err != nil {
			log.Println("Failed to fetch posts")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}

		data := struct {
			UserID string
			Avatar string
			Posts  []database.Post
		}{
			UserID: userID,
			Avatar: session.Values["avatar"].(string),
			Posts:  posts,
		}

		err = templates.ExecuteTemplate(w, "home.html", data) // Changed from "moderator.html" to "home.html"
		if err != nil {
			log.Println("Error rendering home page:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
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
				errHandler(w, r, &err)
				return
			}
		} else if r.FormValue("report_post") != "" {
			postID := r.FormValue("report_post")
			reportReason := r.FormValue("report_reason")
			_, err := db.Exec("INSERT INTO reports (post_id, reported_by, report_reason) VALUES (?, ?, ?)", postID, userID, reportReason)
			if err != nil {
				log.Println("Failed to report post")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
		} else if r.FormValue("delete_comment") != "" {
			commentID := r.FormValue("delete_comment")
			_, err := db.Exec("DELETE FROM comment WHERE commentid = ?", commentID)
			if err != nil {
				log.Println("Failed to delete comment")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
		} else if r.FormValue("report_comment") != "" {
			commentID := r.FormValue("report_comment")
			_, err := db.Exec("INSERT INTO reports (comment_id, reported_by, report_reason) VALUES (?, ?, ?)", commentID, userID, "Reported by moderator")
			if err != nil {
				log.Println("Failed to report comment")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
		}
		log.Println("Moderator action completed")
		http.Redirect(w, r, "/moderator", http.StatusSeeOther)
	default:
		log.Println("Method not allowed")
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		errHandler(w, r, &err)
		return
	}
}
