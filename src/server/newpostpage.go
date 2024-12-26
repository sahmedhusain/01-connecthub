package server

import (
	"database/sql"
	"forum/database"
	"log"
	"net/http"
	"strconv"
)

func NewPostPage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Render the new post page template
		err := templates.ExecuteTemplate(w, "newpost.html", nil)
		if err != nil {
			log.Println("Error rendering new post page:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

	case "POST":
		// Parse the form data
		err := r.ParseForm()
		if err != nil {
			log.Println("Failed to parse form data")
			err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
			errHandler(w, r, &err)
			return
		}

		// Get the post content and user ID
		content := r.FormValue("content")
		userID := r.FormValue("user")

		if content == "" || userID == "" {
			log.Println("Invalid form data")
			err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
			errHandler(w, r, &err)
			return
		}

		// Insert the new post into the database
		db, err := sql.Open("sqlite3", "./database/main.db")
		if err != nil {
			log.Println("Database connection failed")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}
		defer db.Close()

		image := sql.NullString{String: r.FormValue("image"), Valid: r.FormValue("image") != ""}
		postID, err := database.InsertPost(db, content, image, userID)
		if err != nil {
			log.Println("Failed to insert post data")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		// Associate the post with categories
		categoryIDs := r.Form["categories"]
		for _, categoryID := range categoryIDs {
			categoryIDInt, err := strconv.Atoi(categoryID)
			if err != nil {
				log.Println("Failed to parse category ID")
				err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
				errHandler(w, r, &err)
				return
			}
			err = database.InsertPostCategory(db, postID, categoryIDInt)
			if err != nil {
				log.Println("Failed to insert post category data")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
		}
		log.Println("New post created with ID:", postID)
		http.Redirect(w, r, "/home?user="+userID, http.StatusSeeOther)

	default:
		log.Println("Method not allowed")
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		errHandler(w, r, &err)
		return
	}
}
