package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Category struct {
	ID          int
	Name        string
	Description string
}

var categories []Category

func DataBase() {

	// Open a connection to the SQLite3 database
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// SQL statements to create tables
	sqlStatements := []string{
		`DROP TABLE IF EXISTS categories;`,
		`CREATE TABLE IF NOT EXISTS categories (
				idcategories INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				description TEXT NULL
			);`,

		`DROP TABLE IF EXISTS comment;`,
		`CREATE TABLE IF NOT EXISTS comment (
				commentid INTEGER PRIMARY KEY AUTOINCREMENT,
				content TEXT NULL,
				comment_at DATETIME NULL,
				post_postid INTEGER NOT NULL,
				user_userid INTEGER NOT NULL,
				FOREIGN KEY (post_postid) REFERENCES post(postid),
				FOREIGN KEY (user_userid) REFERENCES user(userid)
			);`,

		`DROP TABLE IF EXISTS dislike;`,
		`CREATE TABLE IF NOT EXISTS dislike (
				dislikeid INTEGER PRIMARY KEY AUTOINCREMENT,
				dislike_at DATE NULL,
				user_userid INTEGER NOT NULL,
				post_postid INTEGER NOT NULL,
				FOREIGN KEY (user_userid) REFERENCES user(userid),
				FOREIGN KEY (post_postid) REFERENCES post(postid)
			);`,

		`DROP TABLE IF EXISTS like;`,
		`CREATE TABLE IF NOT EXISTS like (
				likeid INTEGER PRIMARY KEY AUTOINCREMENT,
				like_at DATETIME NULL,
				post_postid INTEGER NOT NULL,
				user_userid INTEGER NOT NULL,
				FOREIGN KEY (post_postid) REFERENCES post(postid),
				FOREIGN KEY (user_userid) REFERENCES user(userid)
			);`,

		`DROP TABLE IF EXISTS post;`,
		`CREATE TABLE IF NOT EXISTS post (
				postid INTEGER PRIMARY KEY AUTOINCREMENT,
				image TEXT NULL,
				content TEXT NULL,
				post_at DATETIME NOT NULL,
				user_userid INTEGER NOT NULL,
				FOREIGN KEY (user_userid) REFERENCES user(userid)
			);`,

		`DROP TABLE IF EXISTS post_has_categories;`,
		`CREATE TABLE IF NOT EXISTS post_has_categories (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				post_postid INTEGER NOT NULL,
				post_user_userid INTEGER NOT NULL,
				categories_idcategories INTEGER NOT NULL,
				FOREIGN KEY (post_postid) REFERENCES post(postid),
				FOREIGN KEY (post_user_userid) REFERENCES user(userid),
				FOREIGN KEY (categories_idcategories) REFERENCES categories(idcategories)
			);`,

		`DROP TABLE IF EXISTS session;`,
		`CREATE TABLE IF NOT EXISTS session (
				sessionid INTEGER PRIMARY KEY AUTOINCREMENT,
				start DATETIME NOT NULL,
				end DATETIME NOT NULL
			);`,

		`DROP TABLE IF EXISTS user;`,
		`CREATE TABLE IF NOT EXISTS user (
				userid INTEGER PRIMARY KEY AUTOINCREMENT,
				F_name TEXT NOT NULL,
				L_name TEXT NOT NULL,
				Username TEXT NOT NULL,
				Email TEXT NOT NULL,
				password TEXT NOT NULL,
				session_sessionid INTEGER NOT NULL,
				role_id INTEGER NOT NULL,
				FOREIGN KEY (session_sessionid) REFERENCES session(sessionid),
				FOREIGN KEY (role_id) REFERENCES user_roles(roleid)
			);`,

		`DROP TABLE IF EXISTS user_roles;`,
		`CREATE TABLE IF NOT EXISTS user_roles (
				roleid INTEGER PRIMARY KEY AUTOINCREMENT,
				role_name TEXT NOT NULL
			);`,
		`INSERT INTO user_roles (role_name) VALUES ('administrator'), ('moderator'), ('user'), ('guest');`,

		`DROP TABLE IF EXISTS friends;`,
		`CREATE TABLE IF NOT EXISTS friends (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_userid INTEGER NOT NULL,
				friend_userid INTEGER NOT NULL,
				FOREIGN KEY (user_userid) REFERENCES user(userid),
				FOREIGN KEY (friend_userid) REFERENCES user(userid)
			);`,

		`DROP TABLE IF EXISTS followers;`,
		`CREATE TABLE IF NOT EXISTS followers (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_userid INTEGER NOT NULL,
				follower_userid INTEGER NOT NULL,
				FOREIGN KEY (user_userid) REFERENCES user(userid),
				FOREIGN KEY (follower_userid) REFERENCES user(userid)
			);`,

		`DROP TABLE IF EXISTS notifications;`,
		`CREATE TABLE IF NOT EXISTS notifications (
				notificationid INTEGER PRIMARY KEY AUTOINCREMENT,
				user_userid INTEGER NOT NULL,
				message TEXT NOT NULL,
				created_at DATETIME NOT NULL,
				FOREIGN KEY (user_userid) REFERENCES user(userid)
			);`,
	}

	// Execute each SQL statement
	for _, stmt := range sqlStatements {
		_, err := db.Exec(stmt)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Insert data into categories table
	insertStatements := []string{
		`INSERT INTO categories (name, description) VALUES ('Technology', 'All about tech');`,
		`INSERT INTO categories (name, description) VALUES ('Science', 'Scientific discoveries and research');`,
		`INSERT INTO categories (name, description) VALUES ('Art', 'Artistic expressions and creations');`,
	}

	for _, stmt := range insertStatements {
		_, err := db.Exec(stmt)
		if err != nil {
			log.Fatal(err)
		}
	}

	rows, err := db.Query("SELECT * FROM categories")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Store categories in an array

	for rows.Next() {
		var id int
		var name, description string
		if err := rows.Scan(&id, &name, &description); err != nil {
			log.Fatal(err)
		}
		categories = append(categories, Category{ID: id, Name: name, Description: description})
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	// Print categories
	for _, category := range categories {
		fmt.Println(category)
	}

}
