package main

import (
	"fmt"
	"forum/database"
)

func main() {

	database.DataBase()
	fmt.Println("Tables created successfully")
}
