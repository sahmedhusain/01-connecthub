package database

import (
	"database/sql"
	"log"
)

func CommentReactions() {
	// Open a connection to the SQLite3 database
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	const DropCommentDislikeTable = `DROP TABLE IF EXISTS comment_dislikes;`
	const DropCommentLikeTable = `DROP TABLE IF EXISTS comment_likes;`

	dropTableStatements := []string{
		DropCommentDislikeTable,
		DropCommentLikeTable,
	}

	for _, stmt := range dropTableStatements {
		_, err = db.Exec(stmt)
		if err != nil {
			log.Fatal(err)
		}
	}

	const CreateCommentDislikeTable = `
	CREATE TABLE IF NOT EXISTS comment_dislikes (
		dislikeid INTEGER PRIMARY KEY AUTOINCREMENT,
		dislike_at DATETIME NULL,
		commentid INTEGER NOT NULL,
		userid INTEGER NOT NULL,
		FOREIGN KEY (commentid) REFERENCES comment(commentid)
		FOREIGN KEY (userid) REFERENCES user(userid)
	);`

	const CreateCommentLikeTable = `
	CREATE TABLE IF NOT EXISTS comment_likes (
		likeid INTEGER PRIMARY KEY AUTOINCREMENT,
		like_at DATETIME NULL,
		commentid INTEGER NOT NULL,
		userid INTEGER NOT NULL,
		FOREIGN KEY (commentid) REFERENCES comment(commentid)
		FOREIGN KEY (userid) REFERENCES user(userid)
	);`

	createTableStatements := []string{
		CreateCommentDislikeTable,
		CreateCommentLikeTable,
	}

	for _, stmt := range createTableStatements {
		_, err = db.Exec(stmt)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Insert likes
	// Insert likes
	insertLikes := []string{
		`INSERT INTO comment_likes (like_at, commentid, userid) VALUES ('2024-12-21 08:00:00', 1, 1);`,
		`INSERT INTO comment_likes (like_at, commentid, userid) VALUES ('2024-12-21 08:30:00', 2, 2);`,
		`INSERT INTO comment_likes (like_at, commentid, userid) VALUES ('2024-12-21 09:00:00', 3, 3);`,
		`INSERT INTO comment_likes (like_at, commentid, userid) VALUES ('2024-12-21 09:30:00', 4, 4);`,
		`INSERT INTO comment_likes (like_at, commentid, userid) VALUES ('2024-12-21 10:00:00', 5, 5);`,
		`INSERT INTO comment_likes (like_at, commentid, userid) VALUES ('2024-12-21 10:30:00', 6, 6);`,
		`INSERT INTO comment_likes (like_at, commentid, userid) VALUES ('2024-12-21 11:00:00', 7, 7);`,
		`INSERT INTO comment_likes (like_at, commentid, userid) VALUES ('2024-12-21 11:30:00', 8, 8);`,
		`INSERT INTO comment_likes (like_at, commentid, userid) VALUES ('2024-12-21 12:00:00', 9, 9);`,
		`INSERT INTO comment_likes (like_at, commentid, userid) VALUES ('2024-12-21 12:30:00', 10, 10);`,
	}

	// Insert dislikes
	insertDislikes := []string{
		`INSERT INTO comment_dislikes (dislike_at, commentid, userid) VALUES ('2024-12-22 08:00:00', 1, 1);`,
		`INSERT INTO comment_dislikes (dislike_at, commentid, userid) VALUES ('2024-12-22 08:30:00', 2, 2);`,
		`INSERT INTO comment_dislikes (dislike_at, commentid, userid) VALUES ('2024-12-22 09:00:00', 3, 3);`,
		`INSERT INTO comment_dislikes (dislike_at, commentid, userid) VALUES ('2024-12-22 09:30:00', 4, 4);`,
		`INSERT INTO comment_dislikes (dislike_at, commentid, userid) VALUES ('2024-12-22 10:00:00', 5, 5);`,
		`INSERT INTO comment_dislikes (dislike_at, commentid, userid) VALUES ('2024-12-22 10:30:00', 6, 6);`,
		`INSERT INTO comment_dislikes (dislike_at, commentid, userid) VALUES ('2024-12-22 11:00:00', 7, 7);`,
		`INSERT INTO comment_dislikes (dislike_at, commentid, userid) VALUES ('2024-12-22 11:30:00', 8, 8);`,
		`INSERT INTO comment_dislikes (dislike_at, commentid, userid) VALUES ('2024-12-22 12:00:00', 9, 9);`,
		`INSERT INTO comment_dislikes (dislike_at, commentid, userid) VALUES ('2024-12-22 12:30:00', 10, 10);`,
	}

	for _, query := range insertLikes {
		_, err := db.Exec(query)
		if err != nil {
			log.Println(err)
		}
	}

	for _, query := range insertDislikes {
		_, err := db.Exec(query)
		if err != nil {
			log.Println(err)
		}
	}

}
