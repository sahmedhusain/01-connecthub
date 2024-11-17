package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/mattn/go-sqlite3"
)

func main() {
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
            idcategories INTEGER NOT NULL,
            name TEXT NOT NULL,
            description TEXT NULL,
            PRIMARY KEY (idcategories)
        );`,
        `DROP TABLE IF EXISTS comment;`,
        `CREATE TABLE IF NOT EXISTS comment (
            commentid INTEGER NOT NULL,
            content TEXT NULL,
            comment_at DATETIME NULL,
            post_postid INTEGER NOT NULL,
            user_userid INTEGER NOT NULL,
            PRIMARY KEY (commentid, post_postid, user_userid)
        );`,
        `DROP TABLE IF EXISTS dislike;`,
        `CREATE TABLE IF NOT EXISTS dislike (
            dislikeid INTEGER NOT NULL,
            dislike_at DATE NULL,
            user_userid INTEGER NOT NULL,
            post_postid INTEGER NOT NULL,
            PRIMARY KEY (dislikeid, user_userid, post_postid)
        );`,
        `DROP TABLE IF EXISTS like;`,
        `CREATE TABLE IF NOT EXISTS like (
            likeid INTEGER NOT NULL,
            like_at DATETIME NULL,
            post_postid INTEGER NOT NULL,
            user_userid INTEGER NOT NULL,
            PRIMARY KEY (likeid, post_postid, user_userid)
        );`,
        `DROP TABLE IF EXISTS post;`,
        `CREATE TABLE IF NOT EXISTS post (
            postid INTEGER NOT NULL,
            image INTEGER NULL,
            contant INTEGER NULL,
            post_at DATETIME NOT NULL,
            user_userid INTEGER NOT NULL,
            PRIMARY KEY (postid, user_userid)
        );`,
        `DROP TABLE IF EXISTS post_has_categories;`,
        `CREATE TABLE IF NOT EXISTS post_has_categories (
            post_postid INTEGER NOT NULL,
            post_user_userid INTEGER NOT NULL,
            categories_idcategories INTEGER NOT NULL,
            PRIMARY KEY (post_postid, post_user_userid, categories_idcategories)
        );`,
        `DROP TABLE IF EXISTS session;`,
        `CREATE TABLE IF NOT EXISTS session (
            sessionid INTEGER NOT NULL,
            start DATETIME NOT NULL,
            end DATETIME NOT NULL,
            PRIMARY KEY (sessionid)
        );`,
        `DROP TABLE IF EXISTS user;`,
        `CREATE TABLE IF NOT EXISTS user (
            userid INTEGER NOT NULL,
            F_name TEXT NOT NULL,
            L_name TEXT NOT NULL,
            Username TEXT NOT NULL,
            Email TEXT NOT NULL,
            password TEXT NOT NULL,
            session_sessionid INTEGER NOT NULL,
            role_id INTEGER NOT NULL,
            PRIMARY KEY (userid, session_sessionid),
            FOREIGN KEY (role_id) REFERENCES user_roles(roleid)
        );`,
        `DROP TABLE IF EXISTS user_roles;`,
        `CREATE TABLE IF NOT EXISTS user_roles (
            roleid INTEGER NOT NULL,
            role_name TEXT NOT NULL,
            PRIMARY KEY (roleid)
        );`,
        `INSERT INTO user_roles (roleid, role_name) VALUES (1, 'admin'), (2, 'moderator'), (3, 'normal_user');`,
        `DROP TABLE IF EXISTS friends;`,
        `CREATE TABLE IF NOT EXISTS friends (
            user_userid INTEGER NOT NULL,
            friend_userid INTEGER NOT NULL,
            PRIMARY KEY (user_userid, friend_userid)
        );`,
        `DROP TABLE IF EXISTS followers;`,
        `CREATE TABLE IF NOT EXISTS followers (
            user_userid INTEGER NOT NULL,
            follower_userid INTEGER NOT NULL,
            PRIMARY KEY (user_userid, follower_userid)
        );`,
        `DROP TABLE IF EXISTS notifications;`,
        `CREATE TABLE IF NOT EXISTS notifications (
            notificationid INTEGER NOT NULL,
            user_userid INTEGER NOT NULL,
            message TEXT NOT NULL,
            created_at DATETIME NOT NULL,
            PRIMARY KEY (notificationid)
        );`,
    }

    // Execute each SQL statement
    for _, stmt := range sqlStatements {
        _, err := db.Exec(stmt)
        if err != nil {
            log.Fatal(err)
        }
    }

    fmt.Println("Tables created successfully")
}
