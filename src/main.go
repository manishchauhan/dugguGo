package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/manishchauhan/dugguGo/servers/mysqlhttpserver"
	"github.com/manishchauhan/dugguGo/util/auth/jwtAuth"
	"github.com/manishchauhan/dugguGo/util/mysqlDbManager"
	"github.com/manishchauhan/dugguGo/websocket"
)

var (
	db      *mysqlDbManager.DBManager
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
		fmt.Printf("Name %s %s %s\n", username, password, email)
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
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(1000)
	fmt.Println(rowSize)

	userData := &User{
		Id:       rowSize,
		Username: "akuma" + fmt.Sprintf("%d", randomNumber),
		Password: "akuma" + fmt.Sprintf("%d", randomNumber),
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
	// Example usage to update a single record based on a WHERE clause in "users" table
	conditions := "id IN (?, ?, ?)"
	args := []interface{}{1, 2, 3}
	columnsToUpdate := map[string]interface{}{
		"email": "manish1457@yahoo.com",
	}

	rowsAffected, err := db.ExecuteUpdateWithWhere("register", columnsToUpdate, conditions, args...)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%d rows updated.\n", rowsAffected)
}
func deleteData() {

	//query := "DELETE FROM register WHERE username = ?"
	rowsAffected, err := db.ExecuteDeleteAll("register")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%d rows deleted.\n", rowsAffected)
}
func insertMulti() {
	// Example usage to insert multiple records with varying column values
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(1000)

	inserts := []map[string]interface{}{
		{
			"id":       randomNumber + 1,
			"username": "John Doe" + fmt.Sprintf("%d", randomNumber),
			"password": "akuma" + fmt.Sprintf("%d", randomNumber),
			"email":    "john@example.com",
		},
		{
			"id":       randomNumber + 2,
			"username": "John Doe" + fmt.Sprintf("%d", randomNumber),
			"password": "akuma" + fmt.Sprintf("%d", randomNumber),
			"email":    "john@example.com",
		},
		{
			"id":       randomNumber + 3,
			"username": "John Doe" + fmt.Sprintf("%d", randomNumber),
			"password": "akuma" + fmt.Sprintf("%d", randomNumber),
			"email":    "john@example.com",
		},
	}

	rowsAffected, err := db.ExecuteMultiInsert("admin", inserts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%d rows inserted.\n", rowsAffected)
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
	db, err = mysqlDbManager.NewDBManager(dataSourceName)
	if err != nil {
		fmt.Println("Error creating DBManager:", err)
		return
	}
	defer db.Close()
	//	deleteData()
	//
	//writeData()
	//getData()
	//writeData()
	//updateData()
	//getData()
	//getData()
	//
	//deleteData()
	//getData()
	//
	port := os.Getenv("HTTP_SERVER_PORT")
	go func() {

		mysqlhttpserver.StartServer(port, db)

	}()

	//connect web socket
	serverAddr := os.Getenv("WEB_SOCKET_PORT") // or any other desired address

	server := websocket.NewWebSocketServer(serverAddr)

	go func() {
		fmt.Println("Starting")
		if err := server.Start(); err != nil {
			fmt.Printf("WebSocket server error: %v\n", err)
		}
	}()

	// Graceful shutdown handling
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("Shutting down...")
	// Perform cleanup and shutdown operations if needed

	os.Exit(0)
}
