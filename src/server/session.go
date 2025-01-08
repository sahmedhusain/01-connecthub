package server

import "github.com/gorilla/sessions"

var store = sessions.NewCookieStore([]byte("iufheuiifhugdbgfghbgidbgbghfdgbgfbig"))

func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   5, //* 60, 1 hour in seconds
		HttpOnly: true,
	}
}
