package server

import (
	"01connecthub/database"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		json.NewEncoder(w).Encode([]SearchResult{})
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Database connection failed:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var results []SearchResult

	userRows, err := db.Query(`
        SELECT userid, Username, F_name, L_name, Avatar 
        FROM user 
        WHERE Username LIKE ? OR F_name LIKE ? OR L_name LIKE ? 
        ORDER BY Username
    `, "%"+query+"%", "%"+query+"%", "%"+query+"%")
	if err != nil {
		log.Println("Error searching users:", err)
	} else {
		defer userRows.Close()
		for userRows.Next() {
			var result SearchResult
			var avatar sql.NullString
			var fname, lname string
			err := userRows.Scan(&result.ID, &result.Username, &fname, &lname, &avatar)
			if err != nil {
				continue
			}
			result.Type = "user"
			result.Name = fname + " " + lname
			if avatar.Valid {
				result.Avatar = avatar.String
			}
			results = append(results, result)
		}
	}

	categoryRows, err := db.Query(`
        SELECT idcategories, name 
        FROM categories 
        WHERE name LIKE ? 
        ORDER BY name
    `, "%"+query+"%")
	if err != nil {
		log.Println("Error searching categories:", err)
	} else {
		defer categoryRows.Close()
		for categoryRows.Next() {
			var result SearchResult
			err := categoryRows.Scan(&result.CategoryID, &result.Name)
			if err != nil {
				continue
			}
			result.Type = "category"
			results = append(results, result)
		}
	}

	postRows, err := db.Query(`
        SELECT postid, title, content 
        FROM post 
        WHERE title LIKE ? OR content LIKE ? 
        ORDER BY post_at DESC
    `, "%"+query+"%", "%"+query+"%")
	if err != nil {
		log.Println("Error searching posts:", err)
	} else {
		defer postRows.Close()
		for postRows.Next() {
			var result SearchResult
			err := postRows.Scan(&result.ID, &result.Title, &result.Content)
			if err != nil {
				continue
			}
			result.Type = "post"
			if len(result.Content) > 100 {
				result.Content = result.Content[:100] + "..."
			}
			results = append(results, result)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func SearchPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Println("Database connection failed:", err)
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
			return
		} else if err != nil {
			log.Println("Error fetching userid ID from user table:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}
	}

	data := PageData{
		HasSession:  hasSession,
		SearchQuery: query,
		UserID:      userID,
		UserName:    userName,
	}

	if hasSession {
		var avatar sql.NullString
		var roleID int
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

		var roleName string
		if roleID == 1 {
			roleName = "Admin"
		} else if roleID == 2 {
			roleName = "Moderator"
		} else {
			roleName = "User"
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

		notifications, err := database.GetLastNotifications(db, userID)
		if err != nil {
			log.Println("Failed to fetch notifications:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			ErrHandler(w, r, &err)
			return
		}

		data.Avatar = avatar.String
		data.RoleName = roleName
		data.TotalLikes = totalLikes
		data.TotalPosts = totalPosts
		data.RoleID = roleID
		data.Notifications = notifications
	}

	err = templates.ExecuteTemplate(w, "search.html", data)
	if err != nil {
		log.Println("Error rendering search page:", err)
		err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		ErrHandler(w, r, &err)
		return
	}
}
