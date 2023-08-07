package main

import (
	"fmt"
	"log"

	"github.com/manishchauhan/dugguGo/util/mysqlDbManager"
)

func normalDb() {

	dataSourceName := "root:manish@tcp(127.0.0.1:3306)/dugguGo"

	// Get the DBManager instance
	db, err := mysqlDbManager.GetInstance(dataSourceName)
	if err != nil {
		fmt.Println("Error creating DBManager:", err)
		return
	}
	defer db.Close()

	// Example query
	rows, err := db.Query("SELECT * FROM register")
	if err != nil {
		fmt.Println("Error executing query:", err)
		return
	}
	defer rows.Close()

	// Process the rows...

	var id int
	var username string
	var password string
	var email string
	for rows.Next() {
		if err := rows.Scan(&id, &username, &password, &email); err != nil {
			log.Println("Error scanning row:", err)
			continue
		}
		fmt.Printf("Name %s\n", username)
	}

}
func main() {
	normalDb()
}
