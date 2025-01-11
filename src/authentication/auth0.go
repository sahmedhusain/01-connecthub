package authentication

import (
	"context"
	"database/sql"
	"encoding/json"
	"forum/src/server"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))


var (
	auth0Domain       = os.Getenv("AUTH0_DOMAIN")
	auth0ClientID     = os.Getenv("AUTH0_CLIENT_ID")
	auth0ClientSecret = os.Getenv("AUTH0_CLIENT_SECRET")
	auth0RedirectURL  = os.Getenv("AUTH0_REDIRECT_URL")

	auth0Config = &oauth2.Config{
		ClientID:     auth0ClientID,
		ClientSecret: auth0ClientSecret,
		RedirectURL:  auth0RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://" + auth0Domain + "/authorize",
			TokenURL: "https://" + auth0Domain + "/oauth/token",
		},
		Scopes: []string{"openid", "profile", "email"},
	}
)

func HandleAuth0Login(w http.ResponseWriter, r *http.Request) {
	url := auth0Config.AuthCodeURL("state")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleAuth0Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	token, err := auth0Config.Exchange(context.Background(), code)
	if err != nil {
		err := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		server.AutherrHandler(w, r, &err)
		return
	}

	client := auth0Config.Client(context.Background(), token)
	resp, err := client.Get("https://" + auth0Domain + "/userinfo")
	if err != nil {
		err := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		server.AutherrHandler(w, r, &err)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		FirstName string `json:"given_name"`
		LastName  string `json:"family_name"`
		Email     string `json:"email"`
		Username  string `json:"nickname"`
		Avatar    string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		err := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		server.AutherrHandler(w, r, &err)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		err := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		server.AutherrHandler(w, r, &err)
		return
	}
	defer db.Close()

	var userID int
	err = db.QueryRow("SELECT userid FROM user WHERE email = ?", userInfo.Email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// User does not exist, create a new account
			stmt, err := db.Prepare("INSERT INTO user (F_name, L_name, Username, Email, Avatar) VALUES (?, ?, ?, ?, ?)")
			if err != nil {
				err := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				server.AutherrHandler(w, r, &err)
				return
			}
			defer stmt.Close()

			res, err := stmt.Exec(userInfo.FirstName, userInfo.LastName, userInfo.Username, userInfo.Email, userInfo.Avatar)
			if err != nil {
				err := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				server.AutherrHandler(w, r, &err)
				return
			}

			lastID, err := res.LastInsertId()
			if err != nil {
				err := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
				server.AutherrHandler(w, r, &err)
				return
			}
			userID = int(lastID)
		} else {
			err := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
			server.AutherrHandler(w, r, &err)
			return
		}
	}

	session, _ := store.Get(r, "session")
	session.Values["userID"] = userID
	err = session.Save(r, w)
	if err != nil {
		err := server.ErrorPageData{Code: "500", ErrorMsg: "INTERNAL SERVER ERROR"}
		server.AutherrHandler(w, r, &err)
		return
	}

	http.Redirect(w, r, "/home", http.StatusSeeOther)
}
