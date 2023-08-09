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

func (dm *STDBManager) Begin() (*sql.Tx, error) {
	db, err := dm.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.Begin()
}

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
	fmt.Println("----->", values.NumField())
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
