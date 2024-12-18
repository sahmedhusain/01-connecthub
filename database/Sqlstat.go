package database

import (
	"database/sql"
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
	ID         int       // Comment ID
	Content    string    // Comment text
	CommentAt  time.Time // Timestamp of the comment
	PostPostID int       // Post ID the comment belongs to
	UserUserID int       // User ID who made the comment
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

// GetAllCategories retrieves all categories from the database
func GetAllCategories(db *sql.DB) ([]Category, error) {
	rows, err := db.Query("SELECT * FROM categories")
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

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func GetComment(db *sql.DB) ([]User, error) {
	rows, err := db.Query("SELECT *  FROM comment")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Email); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
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
		if err := rows.Scan(&comment.ID, &comment.Content, &commentAt, &comment.PostPostID, &comment.UserUserID); err != nil {
			return nil, err
		}

		// Parse comment_at into a time.Time object
		comment.CommentAt = commentAt

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
