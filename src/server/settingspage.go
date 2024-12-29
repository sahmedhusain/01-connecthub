package server

import (
	"database/sql"
	"fmt"
	"forum/database"
	"io"
	"log"
	"net/http"
	"os"
)

func SettingsPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/settings" {
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

	log.Println("UserID retrieved from session:", userID)

	switch r.Method {
	case "GET":
		var user database.User
		err := db.QueryRow("SELECT id, first_name, last_name, username, email, avatar FROM user WHERE id = ?", userID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Email, &user.Avatar)
		if err != nil {
			log.Println("Failed to fetch user data")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}

		data := struct {
			UserID    string
			FirstName string
			LastName  string
			Username  string
			Email     string
			Avatar    string
		}{
			UserID:    userID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Username:  user.Username,
			Email:     user.Email,
			Avatar:    user.Avatar.String,
		}

		err = templates.ExecuteTemplate(w, "settings.html", data)
		if err != nil {
			log.Println("Error rendering settings page:", err)
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}
	case "POST":
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			log.Println("Failed to parse form data")
			err := ErrorPageData{Code: "400", ErrorMsg: "BAD REQUEST"}
			errHandler(w, r, &err)
			return
		}

		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		var avatarPath sql.NullString
		file, handler, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()
			avatarPath.String = fmt.Sprintf("static/uploads/%s", handler.Filename)
			avatarPath.Valid = true

			// Ensure the directory exists
			os.MkdirAll("static/uploads", os.ModePerm)

			f, err := os.OpenFile(avatarPath.String, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				log.Println("Failed to open file for writing")
				err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &err)
				return
			}
			defer f.Close()
			io.Copy(f, file)
		} else if err != http.ErrMissingFile {
			log.Println("Failed to upload avatar")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}

		if password != "" {
			_, err = db.Exec("UPDATE user SET first_name = ?, last_name = ?, username = ?, email = ?, password = ?, avatar = ? WHERE id = ?", firstName, lastName, username, email, password, avatarPath, userID)
		} else {
			_, err = db.Exec("UPDATE user SET first_name = ?, last_name = ?, username = ?, email = ?, avatar = ? WHERE id = ?", firstName, lastName, username, email, avatarPath, userID)
		}

		if err != nil {
			log.Println("Failed to update user data")
			err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &err)
			return
		}
		log.Println("User data updated with ID:", userID)
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
	default:
		log.Println("Method not allowed")
		err := ErrorPageData{Code: "405", ErrorMsg: "METHOD NOT ALLOWED"}
		errHandler(w, r, &err)
		return
	}
}

func ShowPassword(w http.ResponseWriter, r *http.Request) {
    // Retrieve UserID from session
    session, _ := store.Get(r, "session-name")
    userID, ok := session.Values["userID"].(string)
    if !ok || userID == "" {
        log.Println("UserID not found in session, redirecting to login page")
        http.Redirect(w, r, "/login", http.StatusSeeOther)
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

    var password string
    err = db.QueryRow("SELECT password FROM user WHERE id = ?", userID).Scan(&password)
    if err != nil {
        log.Println("Failed to fetch password")
        errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
        errHandler(w, r, &errData)
        return
    }

    data := struct {
        UserID   string
        Password string
    }{
        UserID:   userID,
        Password: password,
    }

    err = templates.ExecuteTemplate(w, "settings.html", data)
    if err != nil {
        log.Println("Error rendering settings page:", err)
        err := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
        errHandler(w, r, &err)
        return
    }
}
