package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

const (
	Createusrquery = `CREATE TABLE IF NOT EXISTS user (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		email TEXT,
		password TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`
	meow = `SELECT * FROM user  WHERE id = ? AND name = ?;`
)

func main() {
	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		panic(err)
	}

	// Create table
	x := db.QueryRow(meow, 1, "meow")
	fmt.Println(x)

}
