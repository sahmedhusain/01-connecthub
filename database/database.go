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
		CREATE TABLE IF NOT EXISTS dislikes (
			dislikeid INTEGER PRIMARY KEY AUTOINCREMENT,
			dislike_at DATETIME NULL,
			post_postid INTEGER NOT NULL,
			user_userid INTEGER NOT NULL,
			FOREIGN KEY (post_postid) REFERENCES post(postid),
			FOREIGN KEY (user_userid) REFERENCES user(userid)
		);`

	const CreateLikeTable = `
		CREATE TABLE IF NOT EXISTS likes (
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
		`DROP TABLE IF EXISTS dislikes;`, CreateDislikeTable,
		`DROP TABLE IF EXISTS likes;`, CreateLikeTable,
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

	// Insert sampledata
	insertStatements := []string{
		`INSERT INTO categories (name, description) VALUES ('Web Development', 'All about web development');`,
		`INSERT INTO categories (name, description) VALUES ('Mobile Development', 'All about mobile development');`,
		`INSERT INTO categories (name, description) VALUES ('Data Science', 'All about data science');`,
		`INSERT INTO categories (name, description) VALUES ('Machine Learning', 'All about machine learning');`,
		`INSERT INTO categories (name, description) VALUES ('Cybersecurity', 'All about cybersecurity');`,
		`INSERT INTO categories (name, description) VALUES ('Cloud Computing', 'All about cloud computing');`,
		`INSERT INTO categories (name, description) VALUES ('DevOps', 'All about DevOps practices');`,
		`INSERT INTO categories (name, description) VALUES ('Blockchain', 'All about blockchain technology');`,
		`INSERT INTO categories (name, description) VALUES ('Game Development', 'All about game development');`,
		`INSERT INTO categories (name, description) VALUES ('UI/UX Design', 'All about UI/UX design');`,

		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id) 
     VALUES ('John', 'Doe', 'johndoe', 'johndoe@example.com', 'password123', 1, 3);`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id) 
     VALUES ('Jane', 'Smith', 'janesmith', 'janesmith@example.com', 'securepass', 2, 3);`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id) 
     VALUES ('Alice', 'Brown', 'alicebrown', 'alicebrown@example.com', 'alice12345', 3, 3);`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id) 
     VALUES ('Michael', 'Johnson', 'mjohnson', 'mjohnson@example.com', 'michaelpass', 4, 3);`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id) 
     VALUES ('Emily', 'Davis', 'edavis', 'edavis@example.com', 'emilypass', 5, 3);`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id) 
     VALUES ('David', 'Wilson', 'dwilson', 'dwilson@example.com', 'davidpass', 6, 3);`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id) 
     VALUES ('Chris', 'Evans', 'cevans', 'cevans@example.com', 'chrispass', 7, 3);`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id) 
     VALUES ('Natalie', 'Portman', 'nportman', 'nportman@example.com', 'nataliepass', 8, 3);`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id) 
     VALUES ('Robert', 'Downey', 'rdowney', 'rdowney@example.com', 'robertpass', 9, 3);`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id) 
     VALUES ('Scarlett', 'Johansson', 'sjohansson', 'sjohansson@example.com', 'scarlettpass', 10, 3);`,

		`INSERT INTO comment (content, comment_at, post_postid, user_userid) 
     VALUES ('This is an interesting post about web development!', '2024-12-05 10:00:00', 1, 1);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) 
     VALUES ('I love the insights in this article, very helpful!', '2024-12-05 11:00:00', 2, 2);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) 
     VALUES ('Great perspective on data science!', '2024-12-05 12:00:00', 3, 3);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) 
     VALUES ('Amazing tips on mobile development!', '2024-12-06 09:00:00', 4, 4);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) 
     VALUES ('Very informative machine learning article.', '2024-12-06 10:00:00', 5, 5);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) 
     VALUES ('Loved the cybersecurity recommendations!', '2024-12-06 11:00:00', 6, 6);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) 
     VALUES ('Great cloud computing analysis!', '2024-12-06 12:00:00', 7, 7);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) 
     VALUES ('DevOps practices are very useful.', '2024-12-06 13:00:00', 8, 8);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) 
     VALUES ('Educational content on blockchain is top-notch.', '2024-12-06 14:00:00', 9, 9);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) 
     VALUES ('Game development tips are very helpful.', '2024-12-06 15:00:00', 10, 10);`,

		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/images/webdev.jpg', 'Exploring the latest trends in web development.', '2024-12-05 10:00:00', 1);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/images/mobiledev.jpg', 'Mobile development: Best practices and tools.', '2024-12-05 11:00:00', 2);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/images/datascience.jpg', 'Data science techniques for beginners.', '2024-12-05 12:00:00', 3);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/images/machinelearning.jpg', 'Machine learning algorithms explained.', '2024-12-06 09:00:00', 4);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/images/cybersecurity.jpg', 'Top cybersecurity threats in 2024.', '2024-12-06 10:00:00', 5);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/images/cloudcomputing.jpg', 'Cloud computing: Benefits and challenges.', '2024-12-06 11:00:00', 6);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/images/devops.jpg', 'DevOps practices for efficient workflows.', '2024-12-06 12:00:00', 7);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/images/blockchain.jpg', 'Blockchain technology: Use cases and future.', '2024-12-06 13:00:00', 8);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/images/gamedev.jpg', 'Game development: Tips for beginners.', '2024-12-06 14:00:00', 9);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/images/uiux.jpg', 'UI/UX design principles for better user experience.', '2024-12-06 15:00:00', 10);`,

		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-05 10:00:00', 1, 1);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-05 11:00:00', 2, 2);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-05 12:00:00', 3, 3);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-06 09:00:00', 4, 4);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-06 10:00:00', 5, 5);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-06 11:00:00', 6, 6);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-06 12:00:00', 7, 7);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-06 13:00:00', 8, 8);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-06 14:00:00', 9, 9);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-07 11:00:00', 2, 3);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-07 12:00:00', 3, 4);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-08 09:00:00', 4, 5);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-08 10:00:00', 5, 6);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-08 11:00:00', 6, 7);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-08 14:00:00', 9, 10);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-08 15:00:00', 10, 1);`,

		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-05 10:00:00', 1, 2);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-05 11:00:00', 2, 3);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-05 12:00:00', 3, 4);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-06 09:00:00', 4, 5);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-06 10:00:00', 5, 6);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-06 11:00:00', 6, 7);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-06 12:00:00', 7, 8);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-06 13:00:00', 8, 9);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-06 14:00:00', 9, 10);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-06 15:00:00', 10, 1);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-07 12:00:00', 3, 5);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-08 09:00:00', 4, 6);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-08 10:00:00', 5, 7);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-08 11:00:00', 6, 8);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-08 12:00:00', 7, 9);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-08 13:00:00', 8, 10);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-08 14:00:00', 9, 1);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-08 15:00:00', 10, 2);`,
	}

	for _, stmt := range insertStatements {
		_, err := db.Exec(stmt)
		if err != nil {
			log.Fatal(err)
		}
	}
}
