package server

import "github.com/gorilla/sessions"

var store = sessions.NewCookieStore([]byte("iufheuiifhugdbgfghbgidbgbghfdgbgfbig"))

func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   30 * 60,
		HttpOnly: true,
	}
}
