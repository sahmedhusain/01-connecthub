package server

import (
	"database/sql"
	"log"
	"net/http"
)

func SignupPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/signup" {
		log.Println("Redirecting to Home page")
		err := ErrorPageData{Code: "404", ErrorMsg: "PAGE NOT FOUND"}
		errHandler(w, r, &err)
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm-password")

		if password != confirmPassword {
			err := templates.ExecuteTemplate(w, "signup.html", map[string]string{
				"ErrorMessage": "Passwords do not match",
			})
			if err != nil {
				log.Println("Error rendering signup page:", err)
				errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &errData)
			}
			return
		}

		db, err := sql.Open("sqlite3", "./database/main.db")
		if err != nil {
			log.Println("Database connection failed")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}
		defer db.Close()

		var usernameExists, emailExists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM user WHERE username = ?)", username).Scan(&usernameExists)
		if err != nil {
			log.Println("Failed to check if username exists")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}

		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM user WHERE email = ?)", email).Scan(&emailExists)
		if err != nil {
			log.Println("Failed to check if email exists")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}

		if usernameExists {
			err := templates.ExecuteTemplate(w, "signup.html", map[string]string{
				"ErrorMessage": "Username already exists",
			})
			if err != nil {
				log.Println("Error rendering signup page:", err)
				errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &errData)
			}
			return
		}

		if emailExists {
			err := templates.ExecuteTemplate(w, "signup.html", map[string]string{
				"ErrorMessage": "Email already exists",
			})
			if err != nil {
				log.Println("Error rendering signup page:", err)
				errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				errHandler(w, r, &errData)
			}
			return
		}

		_, err = db.Exec("INSERT INTO user (username, email, password) VALUES (?, ?, ?)", username, email, password)
		if err != nil {
			log.Println("Failed to insert user data")
			errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			errHandler(w, r, &errData)
			return
		}
		log.Println("User registered with username:", username)
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	err := templates.ExecuteTemplate(w, "signup.html", nil)
	if err != nil {
		log.Println("Error rendering signup page:", err)
		errData := ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		errHandler(w, r, &errData)
	}
}
