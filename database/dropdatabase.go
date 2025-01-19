package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func DropDataBase() {
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	const DropCategoriesTable = `DROP TABLE IF EXISTS categories;`
	const DropCommentTable = `DROP TABLE IF EXISTS comment;`
	const DropDislikeTable = `DROP TABLE IF EXISTS dislikes;`
	const DropLikeTable = `DROP TABLE IF EXISTS likes;`
	const DropCommentDislikeTable = `DROP TABLE IF EXISTS comment_dislikes;`
	const DropCommentLikeTable = `DROP TABLE IF EXISTS comment_likes;`
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

	dropTableStatements := []string{
		DropCategoriesTable,
		DropCommentTable,
		DropDislikeTable,
		DropLikeTable,
		DropCommentDislikeTable,
		DropCommentLikeTable,
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
