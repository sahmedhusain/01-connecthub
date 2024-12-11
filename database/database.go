package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func DataBase() {
	// Open a connection to the SQLite3 database
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Named constants for CREATE TABLE statements
	const CreateCategoriesTable = `
		CREATE TABLE IF NOT EXISTS categories (
			idcategories INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT NULL
		);`

	const CreateCommentTable = `
		CREATE TABLE IF NOT EXISTS comment (
			commentid INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT NULL,
			comment_at DATETIME NULL,
			post_postid INTEGER NOT NULL,
			user_userid INTEGER NOT NULL,
			FOREIGN KEY (post_postid) REFERENCES post(postid),
			FOREIGN KEY (user_userid) REFERENCES user(userid)
		);`

	const CreateDislikeTable = `
		CREATE TABLE IF NOT EXISTS dislike (
			dislikeid INTEGER PRIMARY KEY AUTOINCREMENT,
			dislike_at DATE NULL,
			user_userid INTEGER NOT NULL,
			post_postid INTEGER NOT NULL,
			FOREIGN KEY (user_userid) REFERENCES user(userid),
			FOREIGN KEY (post_postid) REFERENCES post(postid)
		);`

	const CreateLikeTable = `
		CREATE TABLE IF NOT EXISTS like (
			likeid INTEGER PRIMARY KEY AUTOINCREMENT,
			like_at DATETIME NULL,
			post_postid INTEGER NOT NULL,
			user_userid INTEGER NOT NULL,
			FOREIGN KEY (post_postid) REFERENCES post(postid),
			FOREIGN KEY (user_userid) REFERENCES user(userid)
		);`

	const CreatePostTable = `
		CREATE TABLE IF NOT EXISTS post (
			postid INTEGER PRIMARY KEY AUTOINCREMENT,
			image TEXT NULL,
			content TEXT NULL,
			post_at DATETIME NOT NULL,
			user_userid INTEGER NOT NULL,
			FOREIGN KEY (user_userid) REFERENCES user(userid)
		);`

	const CreatePostHasCategoriesTable = `
		CREATE TABLE IF NOT EXISTS post_has_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			post_postid INTEGER NOT NULL,
			post_user_userid INTEGER NOT NULL,
			categories_idcategories INTEGER NOT NULL,
			FOREIGN KEY (post_postid) REFERENCES post(postid),
			FOREIGN KEY (post_user_userid) REFERENCES user(userid),
			FOREIGN KEY (categories_idcategories) REFERENCES categories(idcategories)
		);`

	const CreateSessionTable = `
		CREATE TABLE IF NOT EXISTS session (
			sessionid INTEGER PRIMARY KEY AUTOINCREMENT,
			start DATETIME NOT NULL,
			end DATETIME NOT NULL
		);`

	const CreateUserTable = `
		CREATE TABLE IF NOT EXISTS user (
			userid INTEGER PRIMARY KEY AUTOINCREMENT,
			F_name TEXT NOT NULL,
			L_name TEXT NOT NULL,
			Username TEXT NOT NULL,
			Email TEXT NOT NULL,
			password TEXT NOT NULL,
			session_sessionid INTEGER NOT NULL,
			role_id INTEGER NOT NULL,
			Avatar TEXT ,
			FOREIGN KEY (session_sessionid) REFERENCES session(sessionid),
			FOREIGN KEY (role_id) REFERENCES user_roles(roleid)
		);`

	const CreateUserRolesTable = `
		CREATE TABLE IF NOT EXISTS user_roles (
			roleid INTEGER PRIMARY KEY AUTOINCREMENT,
			role_name TEXT NOT NULL
		);`

	const CreateFriendsTable = `
		CREATE TABLE IF NOT EXISTS friends (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_userid INTEGER NOT NULL,
			friend_userid INTEGER NOT NULL,
			FOREIGN KEY (user_userid) REFERENCES user(userid),
			FOREIGN KEY (friend_userid) REFERENCES user(userid)
		);`

	const CreateFollowersTable = `
		CREATE TABLE IF NOT EXISTS followers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_userid INTEGER NOT NULL,
			follower_userid INTEGER NOT NULL,
			FOREIGN KEY (user_userid) REFERENCES user(userid),
			FOREIGN KEY (follower_userid) REFERENCES user(userid)
		);`

	const CreateNotificationsTable = `
		CREATE TABLE IF NOT EXISTS notifications (
			notificationid INTEGER PRIMARY KEY AUTOINCREMENT,
			user_userid INTEGER NOT NULL,
			message TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (user_userid) REFERENCES user(userid)
		);`

	// Drop and create tables
	sqlStatements := []string{
		`DROP TABLE IF EXISTS categories;`, CreateCategoriesTable,
		`DROP TABLE IF EXISTS comment;`, CreateCommentTable,
		`DROP TABLE IF EXISTS dislike;`, CreateDislikeTable,
		`DROP TABLE IF EXISTS like;`, CreateLikeTable,
		`DROP TABLE IF EXISTS post;`, CreatePostTable,
		`DROP TABLE IF EXISTS post_has_categories;`, CreatePostHasCategoriesTable,
		`DROP TABLE IF EXISTS session;`, CreateSessionTable,
		`DROP TABLE IF EXISTS user;`, CreateUserTable,
		`DROP TABLE IF EXISTS user_roles;`, CreateUserRolesTable,
		`DROP TABLE IF EXISTS friends;`, CreateFriendsTable,
		`DROP TABLE IF EXISTS followers;`, CreateFollowersTable,
		`DROP TABLE IF EXISTS notifications;`, CreateNotificationsTable,
	}

	// Execute each SQL statement
	for _, stmt := range sqlStatements {
		_, err := db.Exec(stmt)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Insert sample data
	insertStatements := []string{
		`INSERT INTO categories (name, description) VALUES ('Technology', 'All about tech');`,
		`INSERT INTO categories (name, description) VALUES ('Science', 'Scientific discoveries and research');`,
		`INSERT INTO categories (name, description) VALUES ('Art', 'Artistic expressions and creations');`,

		`INSERT INTO user_roles (role_name) VALUES ('administrator'), ('moderator'), ('user'), ('guest');`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id ) 
		 VALUES ('John', 'Doe', 'johndoe', 'johndoe@example.com', 'password123', 1, 3);`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id) 
		 VALUES ('Jane', 'Smith', 'janesmith', 'janesmith@example.com', 'securepass', 2, 3);`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id) 
		 VALUES ('Alice', 'Brown', 'alicebrown', 'alicebrown@example.com', 'alice12345', 3, 3);`,

		`INSERT INTO comment (content, comment_at, post_postid, user_userid) 
		 VALUES ('This is an interesting post about tech!', '2024-12-05 10:00:00', 1, 1);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) 
		 VALUES ('I love the insights in this article, very helpful!', '2024-12-05 11:00:00', 2, 2);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) 
		 VALUES ('Great perspective on modern science!', '2024-12-05 12:00:00', 3, 3);`,

		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/images/tech.jpg', 'Tech news', '2024-12-05 10:00:00', 1);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/images/science.jpg', 'Science news', '2024-12-05 11:00:00', 2);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/images/art.jpg', 'Art news', '2024-12-05 12:00:00', 3);`,

		// New random users
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id) 
		 VALUES ('Michael', 'Johnson', 'mjohnson', 'mjohnson@example.com', 'michaelpass', 4, 3);`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id) 
		 VALUES ('Emily', 'Davis', 'edavis', 'edavis@example.com', 'emilypass', 5, 3);`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id) 
		 VALUES ('David', 'Wilson', 'dwilson', 'dwilson@example.com', 'davidpass', 6, 3);`,
	}

	for _, stmt := range insertStatements {
		_, err := db.Exec(stmt)
		if err != nil {
			log.Fatal(err)
		}
	}
}
