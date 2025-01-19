package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func DropDataBase() {
	// Open a connection to the SQLite3 database
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// DROP TABLE statements
	const DropCategoriesTable = `DROP TABLE IF EXISTS categories;`
	const DropCommentTable = `DROP TABLE IF EXISTS comment;`
	const DropDislikeTable = `DROP TABLE IF EXISTS dislikes;`
	const DropLikeTable = `DROP TABLE IF EXISTS likes;`
	const DropPostTable = `DROP TABLE IF EXISTS post;`
	const DropPostHasCategoriesTable = `DROP TABLE IF EXISTS post_has_categories;`
	const DropSessionsTable = `DROP TABLE IF EXISTS session;`
	const DropUserTable = `DROP TABLE IF EXISTS user;`
	const DropUserRolesTable = `DROP TABLE IF EXISTS user_roles;`
	const DropFriendsTable = `DROP TABLE IF EXISTS friends;`
	const DropFollowersTable = `DROP TABLE IF EXISTS followers;`
	const DropNotificationsTable = `DROP TABLE IF EXISTS notifications;`
	const DropFollowingTable = `DROP TABLE IF EXISTS following;`
	const DropReportsTable = `DROP TABLE IF EXISTS reports;`

	// Execute DROP TABLE statements
	dropTableStatements := []string{
		DropCategoriesTable,
		DropCommentTable,
		DropDislikeTable,
		DropLikeTable,
		DropPostTable,
		DropPostHasCategoriesTable,
		DropSessionsTable,
		DropUserTable,
		DropUserRolesTable,
		DropFriendsTable,
		DropFollowersTable,
		DropNotificationsTable,
		DropFollowingTable,
		DropReportsTable,
	}

	for _, stmt := range dropTableStatements {
		_, err = db.Exec(stmt)
		if err != nil {
			log.Fatal(err)
		}
	}

}
