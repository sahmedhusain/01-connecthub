package database

import (
	"database/sql"
	"fmt"
	"log"
)

func Select(colToReturn string, table string, where string, input string) (string, error) {
	// Open a connection to the SQLite3 database
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	statement := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", colToReturn, table, where)

	var returnedValue string
	err = db.QueryRow(statement, input).Scan(&returnedValue)
	if err != nil {
		return "", err
	}
	return returnedValue, nil
}
