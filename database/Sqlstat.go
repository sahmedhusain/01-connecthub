package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type User struct {
	ID               int            `json:"id"`
	FirstName        string         `json:"first_name"`
	LastName         string         `json:"last_name"`
	Username         string         `json:"username"`
	Email            string         `json:"email"`
	Password         string         `json:"password"`
	SessionSessionID int            `json:"session_sessionid"`
	RoleID           int            `json:"role_id"`
	Avatar           sql.NullString `json:"avatar"` // Use sql.NullString to handle NULL values
}

type Category struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Comment struct {
	ID        int
	PostID    int
	UserID    int
	Content   string
	CreatedAt time.Time
}

type Post struct {
	PostID     int
	Image      sql.NullString
	Content    string
	PostAt     time.Time
	UserUserID int
	Username   string
	FirstName  string
	LastName   string
	Avatar     sql.NullString
	Likes      int
	Dislikes   int
	Comments   int
	Categories []Category
}

type Notification struct {
	ID        int
	UserID    int
	PostID    int
	Message   string
	CreatedAt time.Time
	UserImage string
	UserName  string // Correct field name
}

type Report struct {
	ID           int
	PostID       int
	ReportedBy   int
	ReportReason string
	CreatedAt    time.Time
}

type UserLog struct {
    ID        int
    UserID    int
    Action    string
    Timestamp time.Time
}

type UserSession struct {
    ID     int
    UserID int
    Start  time.Time
    End    time.Time
}

// GetAllCategories retrieves all categories from the database
func GetAllCategories(db *sql.DB) ([]Category, error) {
	rows, err := db.Query("SELECT idcategories, name, description FROM categories")
	if (err != nil) {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Description); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func GetComments(db *sql.DB) ([]Comment, error) {
	// Query to retrieve all comments
	rows, err := db.Query("SELECT commentid, content, comment_at, post_postid, user_userid FROM comment")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		var commentAt time.Time // SQLite DATETIME is fetched as a string

		// Scan each row into the Comment struct
		if err := rows.Scan(&comment.ID, &comment.Content, &commentAt, &comment.PostID, &comment.UserID); err != nil {
			return nil, err
		}

		// Parse comment_at into a time.Time object
		comment.CreatedAt = commentAt

		comments = append(comments, comment)
	}

	// Check for errors after iterating
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func GetAllPosts(db *sql.DB) ([]Post, error) {
	rows, err := db.Query(`
        SELECT post.postid, post.image, post.content, post.post_at, post.user_userid, user.Username, user.F_name, user.L_name, user.Avatar,
               (SELECT COUNT(*) FROM likes WHERE likes.post_postid = post.postid) AS Likes,
               (SELECT COUNT(*) FROM dislikes WHERE dislikes.post_postid = post.postid) AS Dislikes,
               (SELECT COUNT(*) FROM comment WHERE comment.post_postid = post.postid) AS Comments
        FROM post
        JOIN user ON post.user_userid = user.userid
        ORDER BY post.post_at DESC
    `)
	if err != nil {
		log.Println("Error executing query:", err)
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var postAt string
		if err := rows.Scan(&post.PostID, &post.Image, &post.Content, &postAt, &post.UserUserID, &post.Username, &post.FirstName, &post.LastName, &post.Avatar, &post.Likes, &post.Dislikes, &post.Comments); err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}

		// Parse the postAt string into a time.Time object
		post.PostAt, err = time.Parse(time.RFC3339, postAt)
		if err != nil {
			log.Println("Error parsing post_at:", err)
			return nil, err
		}

		// Fetch categories for the post
		categories, err := getCategoriesForPost(db, post.PostID)
		if err != nil {
			log.Println("Error fetching categories for post:", err)
			return nil, err
		}
		post.Categories = categories

		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error in rows:", err)
		return nil, err
	}

	return posts, nil
}

func getCategoriesForPost(db *sql.DB, postID int) ([]Category, error) {
	rows, err := db.Query(`
        SELECT c.idcategories, c.name, c.description
        FROM categories c
        JOIN post_has_categories phc ON c.idcategories = phc.categories_idcategories
        WHERE phc.post_postid = ?
    `, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Description); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func GetAllUsers(db *sql.DB) ([]User, error) {
	rows, err := db.Query("SELECT userid, F_name, L_name, Username, Email, Avatar FROM user")
	if err != nil {
		log.Println("Error executing query:", err)
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var avatar sql.NullString
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Email, &avatar); err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}
		user.Avatar = avatar
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error in rows:", err)
		return nil, err
	}

	return users, nil
}

func GetFilteredPosts(db *sql.DB, filter string) ([]Post, error) {
	var rows *sql.Rows
	var err error

	switch filter {
	case "following":
		rows, err = db.Query(`
            SELECT post.postid, post.image, post.content, post.post_at, post.user_userid, user.Username, user.F_name, user.L_name, user.Avatar,
                   (SELECT COUNT(*) FROM likes WHERE likes.post_postid = post.postid) AS Likes,
                   (SELECT COUNT(*) FROM dislikes WHERE dislikes.post_postid = post.postid) AS Dislikes,
                   (SELECT COUNT(*) FROM comment WHERE comment.post_postid = post.postid) AS Comments
            FROM post
            JOIN user ON post.user_userid = user.userid
            ORDER BY post.post_at DESC
        `)
	case "friends":
		rows, err = db.Query(`
            SELECT post.postid, post.image, post.content, post.post_at, post.user_userid, user.Username, user.F_name, user.L_name, user.Avatar,
                   (SELECT COUNT(*) FROM likes WHERE likes.post_postid = post.postid) AS Likes,
                   (SELECT COUNT(*) FROM dislikes WHERE dislikes.post_postid = post.postid) AS Dislikes,
                   (SELECT COUNT(*) FROM comment WHERE comment.post_postid = post.postid) AS Comments
            FROM post
            JOIN user ON post.user_userid = user.userid
            ORDER BY post.post_at DESC
        `)
	case "top-rated":
		rows, err = db.Query(`
            SELECT post.postid, post.image, post.content, post.post_at, post.user_userid, user.Username, user.F_name, user.L_name, user.Avatar,
                   (SELECT COUNT(*) FROM likes WHERE likes.post_postid = post.postid) AS Likes,
                   (SELECT COUNT(*) FROM dislikes WHERE dislikes.post_postid = post.postid) AS Dislikes,
                   (SELECT COUNT(*) FROM comment WHERE comment.post_postid = post.postid) AS Comments
            FROM post
            JOIN user ON post.user_userid = user.userid
            ORDER BY Likes DESC, post.post_at DESC
        `)
	case "oldest":
		rows, err = db.Query(`
            SELECT post.postid, post.image, post.content, post.post_at, post.user_userid, user.Username, user.F_name, user.L_name, user.Avatar,
                   (SELECT COUNT(*) FROM likes WHERE likes.post_postid = post.postid) AS Likes,
                   (SELECT COUNT(*) FROM dislikes WHERE dislikes.post_postid = post.postid) AS Dislikes,
                   (SELECT COUNT(*) FROM comment WHERE comment.post_postid = post.postid) AS Comments
            FROM post
            JOIN user ON post.user_userid = user.userid
            ORDER BY post.post_at ASC
        `)
	default:
		rows, err = db.Query(`
            SELECT post.postid, post.image, post.content, post.post_at, post.user_userid, user.Username, user.F_name, user.L_name, user.Avatar,
                   (SELECT COUNT(*) FROM likes WHERE likes.post_postid = post.postid) AS Likes,
                   (SELECT COUNT(*) FROM dislikes WHERE dislikes.post_postid = post.postid) AS Dislikes,
                   (SELECT COUNT(*) FROM comment WHERE comment.post_postid = post.postid) AS Comments
            FROM post
            JOIN user ON post.user_userid = user.userid
			ORDER BY post.post_at DESC
        `, filter)
	}

	if err != nil {
		log.Println("Error executing query:", err)
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var postAt string
		if err := rows.Scan(&post.PostID, &post.Image, &post.Content, &postAt, &post.UserUserID, &post.Username, &post.FirstName, &post.LastName, &post.Avatar, &post.Likes, &post.Dislikes, &post.Comments); err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}

		// Parse the postAt string into a time.Time object
		post.PostAt, err = time.Parse(time.RFC3339, postAt)
		if err != nil {
			log.Println("Error parsing post_at:", err)
			return nil, err
		}

		// Fetch categories for the post
		categories, err := getCategoriesForPost(db, post.PostID)
		if err != nil {
			log.Println("Error fetching categories for post:", err)
			return nil, err
		}
		post.Categories = categories

		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error in rows:", err)
		return nil, err
	}

	return posts, nil
}

func GetPostsByCategory(db *sql.DB, categoryName string) ([]Post, error) {
	rows, err := db.Query(`
        SELECT post.postid, post.image, post.content, post.post_at, post.user_userid, user.Username, user.F_name, user.L_name, user.Avatar,
               (SELECT COUNT(*) FROM likes WHERE likes.post_postid = post.postid) AS Likes,
               (SELECT COUNT(*) FROM dislikes WHERE dislikes.post_postid = post.postid) AS Dislikes,
               (SELECT COUNT(*) FROM comment WHERE comment.post_postid = post.postid) AS Comments
        FROM post
        JOIN user ON post.user_userid = user.userid
        JOIN post_has_categories phc ON post.postid = phc.post_postid
        JOIN categories c ON phc.categories_idcategories = c.idcategories
        WHERE c.name = ?
        ORDER BY post.post_at DESC
    `, categoryName)
	if err != nil {
		log.Println("Error executing query:", err)
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var postAt string
		if err := rows.Scan(&post.PostID, &post.Image, &post.Content, &postAt, &post.UserUserID, &post.Username, &post.FirstName, &post.LastName, &post.Avatar, &post.Likes, &post.Dislikes, &post.Comments); err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}

		// Parse the postAt string into a time.Time object
		post.PostAt, err = time.Parse(time.RFC3339, postAt)
		if err != nil {
			log.Println("Error parsing post_at:", err)
			return nil, err
		}

		// Fetch categories for the post
		categories, err := getCategoriesForPost(db, post.PostID)
		if err != nil {
			log.Println("Error fetching categories for post:", err)
			return nil, err
		}
		post.Categories = categories

		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error in rows:", err)
		return nil, err
	}

	return posts, nil
}

func GetLastNotifications(db *sql.DB, userID string) ([]Notification, error) {
    rows, err := db.Query(`
        SELECT n.notificationid, n.user_userid, n.post_id, n.message, n.created_at, u.Avatar, u.Username
        FROM notifications n
        JOIN user u ON n.user_userid = u.userid
        WHERE n.user_userid = ?
        ORDER BY n.created_at DESC
    `, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var notifications []Notification
    for rows.Next() {
        var notification Notification
        var avatar sql.NullString

        err := rows.Scan(&notification.ID, &notification.UserID, &notification.PostID, &notification.Message, &notification.CreatedAt, &avatar, &notification.UserName)
        if err != nil {
            return nil, err
        }

        notification.UserImage = avatar.String
        notifications = append(notifications, notification)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return notifications, nil
}

func InsertPost(db *sql.DB, content string, image sql.NullString, userID string) (int, error) {
	stmt, err := db.Prepare("INSERT INTO post (image, content, post_at, user_userid) VALUES (?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(image, content, time.Now(), userID)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(lastID), nil
}

func InsertPostCategory(db *sql.DB, postID int, categoryID int) error {
	stmt, err := db.Prepare("INSERT INTO post_has_categories (post_postid, categories_idcategories) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(postID, categoryID)
	return err
}

func GetUserPosts(db *sql.DB, userID string) ([]Post, error) {
	rows, err := db.Query("SELECT postid, image, content, post_at FROM post WHERE user_userid = ? ORDER BY post_at DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.PostID, &post.Image, &post.Content, &post.PostAt); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func GetFollowersCount(db *sql.DB, userID string) (int, error) {
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM followers WHERE user_userid = ?", userID).Scan(&count)
    return count, err
}

func GetFollowingCount(db *sql.DB, userID string) (int, error) {
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM following WHERE user_userid = ?", userID).Scan(&count)
    return count, err
}

func GetFriendsCount(db *sql.DB, userID string) (int, error) {
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM friends WHERE user_userid = ?", userID).Scan(&count)
    return count, err
}

func IsFollowing(db *sql.DB, userID string, profileUserID string) (bool, error) {
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM followers WHERE user_userid = ? AND follower_userid = ?", profileUserID, userID).Scan(&count)
    if err != nil {
        return false, err
    }
    return count > 0, nil
}

func GetTotalUsersCount(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM user").Scan(&count)
	return count, err
}

func GetTotalPostsCount(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM post").Scan(&count)
	return count, err
}

func GetTotalCategoriesCount(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	return count, err
}

func GetAllReports(db *sql.DB) ([]Report, error) {
	rows, err := db.Query("SELECT id, post_id, reported_by, report_reason, created_at FROM reports ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []Report
	for rows.Next() {
		var report Report
		if err := rows.Scan(&report.ID, &report.PostID, &report.ReportedBy, &report.ReportReason, &report.CreatedAt); err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reports, nil
}

func GetCommentsForPost(db *sql.DB, postID int) ([]Comment, error) {
    var comments []Comment

    query := `SELECT commentid, post_postid, user_userid, content, comment_at FROM comment WHERE post_postid = ?`
    rows, err := db.Query(query, postID)
    if err != nil {
        return nil, fmt.Errorf("GetCommentsForPost: %v", err)
    }
    defer rows.Close()

    for rows.Next() {
        var comment Comment
        if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt); err != nil {
            return nil, fmt.Errorf("GetCommentsForPost: %v", err)
        }
        comments = append(comments, comment)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("GetCommentsForPost: %v", err)
    }

    return comments, nil
}

func ToggleLike(db *sql.DB, postID int, userID string) error {
    var exists bool
    err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM likes WHERE post_postid = ? AND user_userid = ?)", postID, userID).Scan(&exists)
    if err != nil {
        return fmt.Errorf("ToggleLike: %v", err)
    }

    if exists {
        _, err = db.Exec("DELETE FROM likes WHERE post_postid = ? AND user_userid = ?", postID, userID)
    } else {
        // Remove dislike if it exists
        _, err = db.Exec("DELETE FROM dislikes WHERE post_postid = ? AND user_userid = ?", postID, userID)
        if err != nil {
            return fmt.Errorf("ToggleLike: %v", err)
        }
        _, err = db.Exec("INSERT INTO likes (post_postid, user_userid) VALUES (?, ?)", postID, userID)
    }
    return err
}

func ToggleDislike(db *sql.DB, postID int, userID string) error {
    var exists bool
    err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM dislikes WHERE post_postid = ? AND user_userid = ?)", postID, userID).Scan(&exists)
    if err != nil {
        return fmt.Errorf("ToggleDislike: %v", err)
    }

    if exists {
        _, err = db.Exec("DELETE FROM dislikes WHERE post_postid = ? AND user_userid = ?", postID, userID)
    } else {
        // Remove like if it exists
        _, err = db.Exec("DELETE FROM likes WHERE post_postid = ? AND user_userid = ?", postID, userID)
        if err != nil {
            return fmt.Errorf("ToggleDislike: %v", err)
        }
        _, err = db.Exec("INSERT INTO dislikes (post_postid, user_userid) VALUES (?, ?)", postID, userID)
    }
    return err
}

func GetUserLogs(db *sql.DB, userID int) ([]UserLog, error) {
    rows, err := db.Query("SELECT id, user_id, action, timestamp FROM user_logs WHERE user_id = ?", userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var logs []UserLog
    for rows.Next() {
        var log UserLog
        if err := rows.Scan(&log.ID, &log.UserID, &log.Action, &log.Timestamp); err != nil {
            return nil, err
        }
        logs = append(logs, log)
    }
    return logs, nil
}

func GetUserSessions(db *sql.DB, userID int) ([]UserSession, error) {
    rows, err := db.Query("SELECT sessionid, userid, start, end FROM sessions WHERE userid = ?", userID)
    if err != nil {
        log.Println("Failed to fetch user sessions:", err)
        return nil, err
    }
    defer rows.Close()

    var sessions []UserSession
    for rows.Next() {
        var session UserSession
        err := rows.Scan(&session.ID, &session.UserID, &session.Start, &session.End)
        if err != nil {
            log.Println("Failed to scan user session:", err)
            return nil, err
        }
        sessions = append(sessions, session)
    }
    return sessions, nil
}

func GetFollowers(db *sql.DB, userID string) ([]User, error) {
    rows, err := db.Query(`
        SELECT user.userid, user.F_name, user.L_name, user.Username, user.Avatar
        FROM followers
        JOIN user ON followers.follower_userid = user.userid
        WHERE followers.user_userid = ?
    `, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var followers []User
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Avatar); err != nil {
            return nil, err
        }
        followers = append(followers, user)
    }
    return followers, nil
}

func GetFollowing(db *sql.DB, userID string) ([]User, error) {
    rows, err := db.Query(`
        SELECT user.userid, user.F_name, user.L_name, user.Username, user.Avatar
        FROM following
        JOIN user ON following.following_userid = user.userid
        WHERE following.user_userid = ?
    `, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var following []User
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Avatar); err != nil {
            return nil, err
        }
        following = append(following, user)
    }
    return following, nil
}

func GetFriends(db *sql.DB, userID string) ([]User, error) {
    rows, err := db.Query(`
        SELECT user.userid, user.F_name, user.L_name, user.Username, user.Avatar
        FROM friends
        JOIN user ON friends.friend_userid = user.userid
        WHERE friends.user_userid = ?
    `, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var friends []User
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Avatar); err != nil {
            return nil, err
        }
        friends = append(friends, user)
    }
    return friends, nil
}

func GetTotalLikes(db *sql.DB, userID string) (int, error) {
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM likes WHERE user_userid = ?", userID).Scan(&count)
    return count, err
}

func GetTotalPosts(db *sql.DB, userID string) (int, error) {
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM post WHERE user_userid = ?", userID).Scan(&count)
    return count, err
}

func GetUserByID(db *sql.DB, userID string) (User, error) {
    var user User
    err := db.QueryRow("SELECT userid, F_name, L_name, Username, Email, Avatar, role_id FROM user WHERE userid = ?", userID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Email, &user.Avatar, &user.RoleID)
    if err != nil {
        return user, err
    }
    return user, nil
}

func GetRoleNameByID(db *sql.DB, roleID int) (string, error) {
    var roleName string
    err := db.QueryRow("SELECT role_name FROM user_roles WHERE roleid = ?", roleID).Scan(&roleName)
    if err != nil {
        log.Printf("Error fetching role name for roleID %d: %v\n", roleID, err) // Add this line for debugging
        return "", err
    }
    return roleName, nil
}
