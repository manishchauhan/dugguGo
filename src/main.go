package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/manishchauhan/dugguGo/util/auth/jwtAuth"
	"github.com/manishchauhan/dugguGo/util/mysqlDbManager"
)

// testing only get data----------------------------------------------------------------
func getData() {

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
func writeData() {

}
func updateData() {

}
func deleteData() {

}

// testing only get data----------------------------------------------------------------
func main() {
	//load env variables first
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	jwtAuth.SetEnvData()
	getData()
	writeData()
	updateData()
	deleteData()
}
