package server

import (
	"database/sql"
	"log"
	"net/http"
	"time"
)

func AddComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Println("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Error opening database:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	postID := r.FormValue("post_id")
	userID := r.FormValue("user_id")
	content := r.FormValue("content")

	log.Println("post_id:", postID)
	log.Println("user_id:", userID)
	log.Println("content:", content)

	if postID == "" || userID == "" || content == "" {
		log.Println("Missing form values")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO comment (content, comment_at, post_postid, user_userid) VALUES (?, ?, ?, ?)", content, time.Now(), postID, userID)
	if err != nil {
		log.Println("Error inserting comment:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/post?id="+postID+"&user="+userID, http.StatusSeeOther)
}
