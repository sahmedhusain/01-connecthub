package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func DataBase() {
	// Open a connection to the SQLite3 database
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// CREATE TABLE statements
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
			categories_idcategories INTEGER NOT NULL,
			FOREIGN KEY (post_postid) REFERENCES post(postid),
			FOREIGN KEY (categories_idcategories) REFERENCES categories(idcategories)
		);`

	const CreateSessionsTable = `
		CREATE TABLE IF NOT EXISTS session (
			sessionid INTEGER PRIMARY KEY AUTOINCREMENT,
			userid INTEGER NOT NULL UNIQUE,
			start DATETIME NOT NULL,
			FOREIGN KEY (userid) REFERENCES user(userid)
		);
	`

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
			Avatar TEXT,
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
			post_id INTEGER NOT NULL,
			message TEXT NOT NULL,
			created_at DATETIME default CURRENT_TIMESTAMP,
			FOREIGN KEY (user_userid) REFERENCES user(userid),
			FOREIGN KEY (post_id) REFERENCES post(postid)
		);`

	const CreateFollowingTable = `
		CREATE TABLE IF NOT EXISTS following (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_userid INTEGER NOT NULL,
			following_userid INTEGER NOT NULL,
			FOREIGN KEY (user_userid) REFERENCES user(userid),
			FOREIGN KEY (following_userid) REFERENCES user(userid)
		);`

	const CreateReportsTable = `
		CREATE TABLE IF NOT EXISTS reports (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			post_id INTEGER NOT NULL,
			reported_by INTEGER NOT NULL,
			report_reason TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (post_id) REFERENCES post(postid),
			FOREIGN KEY (reported_by) REFERENCES user(userid)
		);`

	// Execute CREATE TABLE statements
	createTableStatements := []string{
		CreateCategoriesTable,
		CreateCommentTable,
		CreateDislikeTable,
		CreateLikeTable,
		CreatePostTable,
		CreatePostHasCategoriesTable,
		CreateSessionsTable,
		CreateUserTable,
		CreateUserRolesTable,
		CreateFriendsTable,
		CreateFollowersTable,
		CreateNotificationsTable,
		CreateFollowingTable,
		CreateReportsTable,
	}

	for _, stmt := range createTableStatements {
		_, err = db.Exec(stmt)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Insert sample data
	// New categories: Each category focuses on a specific modern tech domain
	insertCategories := []string{
		`INSERT INTO categories (name, description) VALUES ('AI & ML', 'All about Artificial Intelligence and Machine Learning');`,
		`INSERT INTO categories (name, description) VALUES ('Cloud & DevOps', 'Cloud infrastructure and DevOps best practices');`,
		`INSERT INTO categories (name, description) VALUES ('Cybersecurity', 'Guides and insights on staying secure online');`,
		`INSERT INTO categories (name, description) VALUES ('Blockchain & Web3', 'Decentralized networks and blockchain technologies');`,
		`INSERT INTO categories (name, description) VALUES ('AR/VR & Gaming', 'Immersive technologies and game development');`,
		`INSERT INTO categories (name, description) VALUES ('UI/UX Design', 'Improving digital experiences through design');`,
		`INSERT INTO categories (name, description) VALUES ('IoT & Edge Computing', 'Internet of Things and edge computing trends');`,
		`INSERT INTO categories (name, description) VALUES ('Data Analytics', 'Extracting insights from big data');`,
		`INSERT INTO categories (name, description) VALUES ('Quantum Computing', 'Next-gen computing paradigms and qubits');`,
		`INSERT INTO categories (name, description) VALUES ('SRE & Observability', 'Site Reliability Engineering and system observability');`,
	}

	// Insert user roles and sessions
	insertUserRoles := []string{
		`INSERT INTO user_roles (role_name) VALUES ('Admin');`,
		`INSERT INTO user_roles (role_name) VALUES ('Moderator');`,
		`INSERT INTO user_roles (role_name) VALUES ('User');`,
		`INSERT INTO user_roles (role_name) VALUES ('Guest');`,
	}

	// Insert users
	// Assign all users to role_id=3 (User) for simplicity, except first user as Admin (role_id=1) and second as Moderator (role_id=2)
	insertUsers := []string{
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id, Avatar) VALUES ('Alicia', 'Nguyen', 'aliceN', 'aliceN@example.com', 'alicePass', 1, 1, 'https://randomuser.me/api/portraits/women/1.jpg');`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id, Avatar) VALUES ('Brian', 'Lee', 'brianL', 'brianL@example.com', 'brianPass', 2, 2, 'https://randomuser.me/api/portraits/men/1.jpg');`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id, Avatar) VALUES ('Caroline', 'Smith', 'caroS', 'carolineS@example.com', 'carolinePass', 3, 3, 'https://randomuser.me/api/portraits/women/2.jpg');`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id, Avatar) VALUES ('Daniel', 'Foster', 'danF', 'danF@example.com', 'danielPass', 4, 3, 'https://randomuser.me/api/portraits/men/2.jpg');`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id, Avatar) VALUES ('Elena', 'Garcia', 'elenaG', 'elenaG@example.com', 'elenaPass', 5, 3, 'https://randomuser.me/api/portraits/women/3.jpg');`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id, Avatar) VALUES ('Farhan', 'Khan', 'farhanK', 'farhanK@example.com', 'farhanPass', 6, 3, 'https://randomuser.me/api/portraits/men/3.jpg');`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id, Avatar) VALUES ('Grace', 'Li', 'graceL', 'graceL@example.com', 'gracePass', 7, 3, 'https://randomuser.me/api/portraits/women/4.jpg');`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id, Avatar) VALUES ('Hiroshi', 'Tanaka', 'hiroshiT', 'hiroshiT@example.com', 'hiroshiPass', 8, 3, 'https://randomuser.me/api/portraits/men/4.jpg');`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id, Avatar) VALUES ('Irene', 'Santos', 'ireneS', 'ireneS@example.com', 'irenePass', 9, 3, 'https://randomuser.me/api/portraits/women/5.jpg');`,
		`INSERT INTO user (F_name, L_name, Username, Email, password, session_sessionid, role_id, Avatar) VALUES ('Jamal', 'Roberts', 'jamalR', 'jamalR@example.com', 'jamalPass', 10, 3, 'https://randomuser.me/api/portraits/men/5.jpg');`,
	}

	// Insert posts (each linked to a user and posted at a unique time)
	// Each post content reflects the categories it will be associated with
	insertPosts := []string{
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/database/images/ai_ml.jpg', 'Exploring cutting-edge transformer models for NLP tasks.', '2024-12-19 08:30:00', 1);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/database/images/cloud_devops.jpg', 'Implementing CI/CD pipelines on AWS for seamless deployments.', '2024-12-19 09:45:00', 2);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/database/images/cybersec.jpg', 'Top strategies for ransomware protection in modern enterprises.', '2024-12-19 10:15:00', 3);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/database/images/blockchain_web3.jpg', 'Understanding Ethereum Layer-2 scaling solutions.', '2024-12-19 11:00:00', 4);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/database/images/ar_vr_gaming.jpg', 'Building immersive AR experiences with Unity.', '2024-12-19 11:45:00', 5);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/database/images/ui_ux.jpg', 'Enhancing user engagement through micro-interactions.', '2024-12-19 12:30:00', 6);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/database/images/iot_edge.jpg', 'Optimizing sensor networks with edge computing analytics.', '2024-12-19 13:15:00', 7);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/database/images/data_analytics.jpg', 'Leveraging big data frameworks for predictive analytics.', '2024-12-19 14:00:00', 8);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/database/images/quantum.jpg', 'Quantum error correction: The next frontier.', '2024-12-19 14:45:00', 9);`,
		`INSERT INTO post (image, content, post_at, user_userid) VALUES ('/database/images/sre_observability.jpg', 'Implementing distributed tracing for improved observability.', '2024-12-19 15:30:00', 10);`,
	}

	// Insert comments (related to posts and users)
	insertComments := []string{
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) VALUES ('Fascinating look into NLP!', '2024-12-20 09:00:00', 1, 2);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) VALUES ('I love the CI/CD pipeline tips.', '2024-12-20 09:30:00', 2, 1);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) VALUES ('Great cyber defense strategies.', '2024-12-20 10:00:00', 3, 4);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) VALUES ('Layer-2 solutions are game-changers!', '2024-12-20 10:30:00', 4, 3);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) VALUES ('Unity AR examples are spot-on.', '2024-12-20 11:00:00', 5, 6);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) VALUES ('Micro-interactions improve UX a lot.', '2024-12-20 11:30:00', 6, 7);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) VALUES ('Edge analytics are the future.', '2024-12-20 12:00:00', 7, 8);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) VALUES ('Big data insights FTW!', '2024-12-20 12:30:00', 8, 9);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) VALUES ('Quantum is mind-blowing!', '2024-12-20 13:00:00', 9, 10);`,
		`INSERT INTO comment (content, comment_at, post_postid, user_userid) VALUES ('Traces are essential for debugging.', '2024-12-20 13:30:00', 10, 1);`,
	}

	// Insert likes
	insertLikes := []string{
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-21 08:00:00', 1, 3);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-21 08:30:00', 2, 4);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-21 09:00:00', 3, 5);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-21 09:30:00', 4, 6);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-21 10:00:00', 5, 7);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-21 10:30:00', 6, 8);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-21 11:00:00', 7, 9);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-21 11:30:00', 8, 10);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-21 12:00:00', 9, 1);`,
		`INSERT INTO likes (like_at, post_postid, user_userid) VALUES ('2024-12-21 12:30:00', 10, 2);`,
	}

	// Insert dislikes
	insertDislikes := []string{
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-22 08:00:00', 1, 4);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-22 08:30:00', 2, 5);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-22 09:00:00', 3, 6);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-22 09:30:00', 4, 7);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-22 10:00:00', 5, 8);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-22 10:30:00', 6, 9);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-22 11:00:00', 7, 10);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-22 11:30:00', 8, 1);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-22 12:00:00', 9, 2);`,
		`INSERT INTO dislikes (dislike_at, post_postid, user_userid) VALUES ('2024-12-22 12:30:00', 10, 3);`,
	}

	// Posts has categories (linking each post to relevant categories)
	insertPostHasCategories := []string{
		// Post 1: AI & ML
		`INSERT INTO post_has_categories (post_postid, categories_idcategories) VALUES (1, 1);`,
		// Post 2: Cloud & DevOps
		`INSERT INTO post_has_categories (post_postid, categories_idcategories) VALUES (2, 2);`,
		// Post 3: Cybersecurity
		`INSERT INTO post_has_categories (post_postid, categories_idcategories) VALUES (3, 3);`,
		// Post 4: Blockchain & Web3
		`INSERT INTO post_has_categories (post_postid, categories_idcategories) VALUES (4, 4);`,
		// Post 5: AR/VR & Gaming
		`INSERT INTO post_has_categories (post_postid, categories_idcategories) VALUES (5, 5);`,
		// Post 6: UI/UX Design
		`INSERT INTO post_has_categories (post_postid, categories_idcategories) VALUES (6, 6);`,
		// Post 7: IoT & Edge Computing
		`INSERT INTO post_has_categories (post_postid, categories_idcategories) VALUES (7, 7);`,
		// Post 8: Data Analytics
		`INSERT INTO post_has_categories (post_postid, categories_idcategories) VALUES (8, 8);`,
		// Post 9: Quantum Computing
		`INSERT INTO post_has_categories (post_postid, categories_idcategories) VALUES (9, 9);`,
		// Post 10: SRE & Observability
		`INSERT INTO post_has_categories (post_postid, categories_idcategories) VALUES (10, 10);`,
	}

	// Insert notifications as examples
	insertNotifications := []string{
		`INSERT INTO notifications (user_userid, post_id, message, created_at) VALUES (1, 1, 'Your post on AI & ML just received a new comment!', '2024-12-20 14:00:00');`,
		`INSERT INTO notifications (user_userid, post_id, message, created_at) VALUES (2, 2, 'Your Cloud & DevOps post was liked by a user!', '2024-12-20 14:30:00');`,
	}

	// Insert friends
	insertFriends := []string{
		`INSERT INTO friends (user_userid, friend_userid) VALUES (1, 3);`,
		`INSERT INTO friends (user_userid, friend_userid) VALUES (2, 4);`,
		`INSERT INTO friends (user_userid, friend_userid) VALUES (1, 2);`,
		`INSERT INTO friends (user_userid, friend_userid) VALUES (1, 3);`,
		`INSERT INTO friends (user_userid, friend_userid) VALUES (2, 1);`,
		`INSERT INTO friends (user_userid, friend_userid) VALUES (2, 3);`,
		`INSERT INTO friends (user_userid, friend_userid) VALUES (3, 1);`,
		`INSERT INTO friends (user_userid, friend_userid) VALUES (3, 2);`,
	}

	// Insert followers
	insertFollowers := []string{
		`INSERT INTO followers (user_userid, follower_userid) VALUES (1, 5);`,
		`INSERT INTO followers (user_userid, follower_userid) VALUES (2, 6);`,
		`INSERT INTO followers (user_userid, follower_userid) VALUES (1, 2);`,
		`INSERT INTO followers (user_userid, follower_userid) VALUES (1, 3);`,
		`INSERT INTO followers (user_userid, follower_userid) VALUES (2, 1);`,
		`INSERT INTO followers (user_userid, follower_userid) VALUES (2, 3);`,
		`INSERT INTO followers (user_userid, follower_userid) VALUES (3, 1);`,
		`INSERT INTO followers (user_userid, follower_userid) VALUES (3, 2);`,
	}

	// Insert following
	insertFollowing := []string{
		`INSERT INTO following (user_userid, following_userid) VALUES (1, 2);`,
		`INSERT INTO following (user_userid, following_userid) VALUES (1, 3);`,
		`INSERT INTO following (user_userid, following_userid) VALUES (2, 1);`,
		`INSERT INTO following (user_userid, following_userid) VALUES (2, 3);`,
		`INSERT INTO following (user_userid, following_userid) VALUES (3, 1);`,
		`INSERT INTO following (user_userid, following_userid) VALUES (3, 2);`,
	}

	// Combine all insert statements
	allInserts := [][]string{
		insertCategories,
		insertUserRoles,
		insertUsers,
		insertPosts,
		insertComments,
		insertLikes,
		insertDislikes,
		insertPostHasCategories,
		insertNotifications,
		insertFriends,
		insertFollowers,
		insertFollowing,
	}

	for _, group := range allInserts {
		for _, stmt := range group {
			_, err := db.Exec(stmt)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
