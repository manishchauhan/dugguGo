package mysqlDbManager

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

var (
	once     sync.Once
	instance *STDBManager
)

type STDBManager struct {
	dbPool *sql.DB
	mu     sync.Mutex
}

/*thread safe only once instance*/
func GetInstance(dataSourceName string) (*STDBManager, error) {
	var err error
	once.Do(func() {
		dbPool, err := sql.Open("mysql", dataSourceName)
		if err != nil {
			return
		}

		// Set the maximum number of connections to 100
		dbPool.SetMaxOpenConns(100)

		// check the connection
		if err = dbPool.Ping(); err != nil {
			dbPool.Close()
			return
		}

		instance = &STDBManager{dbPool: dbPool}
	})

	return instance, err
}

/*close db*/
func (dm *STDBManager) Close() error {
	return dm.dbPool.Close()
}

/*get connection*/
func (dm *STDBManager) GetConnection() (*sql.DB, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	return dm.dbPool, nil
}

func (dm *STDBManager) Execute(query string, args ...interface{}) (sql.Result, error) {
	db, err := dm.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.Exec(query, args...)
}

func (dm *STDBManager) QueryRow(query string, args ...interface{}) (*sql.Row, error) {
	db, err := dm.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.QueryRow(query, args...), nil
}

func (dm *STDBManager) Query(query string, args ...interface{}) (*sql.Rows, error) {
	db, err := dm.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.Query(query, args...)
}

// Begin starts a transaction. The default isolation level is dependent on the driver.
func (dm *STDBManager) Begin() (*sql.Tx, error) {
	db, err := dm.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.Begin()
}

// note by manish chauhan :- delete all record from a table
func (dbManager *STDBManager) ExecuteDeleteAll(tableName string) (int64, error) {
	query := fmt.Sprintf("DELETE FROM `%s`", tableName) // Properly escape table name
	return dbManager.ExecuteDelete(query)
}

/*
	Example usage to delete records based on flexible WHERE clauses from "any" table
	conditions := "id IN (?, ?, ?) AND name LIKE ?"
	idsToDelete := []interface{}{1, 2, 3}
	likePattern := "John%"
	args := append(idsToDelete, likePattern)
	rowsAffected, err := dm.ExecuteDeleteWithWhere("users", conditions, args...)
*/

func (dm *STDBManager) ExecuteDeleteWithWhere(tableName string, conditions string, args ...interface{}) (int64, error) {
	if conditions == "" {
		return 0, fmt.Errorf("conditions cannot be empty")
	}

	query := fmt.Sprintf("DELETE FROM `%s` WHERE %s", tableName, conditions)
	return dm.ExecuteDelete(query, args...)
}

// note by manish chauhan :- use this function to update data (single record or multi-record)
func (dm *STDBManager) ExecuteDelete(query string, args ...interface{}) (int64, error) {

	result, err := dm.Execute(query, args...)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil

}

// note by manish chauhan :- use this function to update data (single record or multi-record)
func (dm *STDBManager) ExecuteUpdate(tableName string, args ...interface{}) {

}

// note by manish chauhan :- use this function to insert data
// ExecuteInsert inserts data into the specified table and returns the last inserted ID.
func (dm *STDBManager) ExecuteInsert(tableName string, data interface{}) (int64, error) {
	// Build the SQL query for the INSERT statement
	query := buildInsertQuery(tableName, data)

	// Prepare the statement
	db, err := dm.GetConnection()
	if err != nil {
		return 0, err
	}
	stmt, err := db.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare SQL statement: %w", err)
	}
	defer stmt.Close()

	// Execute the statement with the provided data
	result, err := stmt.Exec(extractFieldValues(data)...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute SQL statement: %w", err)
	}

	// Get the last inserted ID
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}

	return lastInsertID, nil
}

// buildInsertQuery constructs the SQL query for the INSERT statement.
func buildInsertQuery(tableName string, data interface{}) string {
	var valueFields, valuePlaceholders []string
	values := reflect.ValueOf(data).Elem()
	// Iterate over struct fields to build the list of fields and placeholders
	for i := 0; i < values.NumField(); i++ {
		fieldName := values.Type().Field(i).Name
		valueFields = append(valueFields, fieldName)
		valuePlaceholders = append(valuePlaceholders, "?")
	}

	// Construct and return the complete SQL query
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(valueFields, ", "), strings.Join(valuePlaceholders, ", "))
}

// extractFieldValues extracts field values from the struct.
func extractFieldValues(data interface{}) []interface{} {
	var fieldValues []interface{}
	values := reflect.ValueOf(data).Elem()

	// Iterate over struct fields to extract values
	for i := 0; i < values.NumField(); i++ {
		fieldValue := values.Field(i).Interface()
		fieldValues = append(fieldValues, fieldValue)
	}

	return fieldValues
}
