package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/manishchauhan/dugguGo/util/auth/jwtAuth"
	"github.com/manishchauhan/dugguGo/util/mysqlDbManager"
)

var (
	db      *mysqlDbManager.STDBManager
	rowSize int = 0
	err     error
)

// testing only get data----------------------------------------------------------------
func getData() {

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
		rowSize += 1
		fmt.Printf("Name %s\n", username)
	}

}

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func writeData() {

	// Define the data you want to insert
	userData := &User{
		Id:       rowSize,
		Username: "akuma",
		Password: "akumakapassword",
		Email:    "new@example.com",
	}

	// Specify the table name
	tableName := "register"

	// Insert the data into the table
	lastInsertID, err := db.ExecuteInsert(tableName, userData)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Last Inserted ID:", lastInsertID)
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

	dataSourceName := os.Getenv("DATABASE_KEY")

	// Get the DBManager instance
	db, err = mysqlDbManager.GetInstance(dataSourceName)
	if err != nil {
		fmt.Println("Error creating DBManager:", err)
		return
	}
	getData()
	writeData()

	updateData()
	deleteData()
	defer db.Close()
}
