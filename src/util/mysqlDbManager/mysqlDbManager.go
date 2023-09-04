package mysqlDbManager

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type DBManager struct {
	dbPool *sql.DB
}

func NewDBManager(dataSourceName string) (*DBManager, error) {
	db := &DBManager{}
	var err error

	db.dbPool, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}

	// Set the maximum number of connections to 100
	db.dbPool.SetMaxOpenConns(100)

	// Check the connection
	if err = db.dbPool.Ping(); err != nil {
		db.dbPool.Close()
		return nil, err
	}

	return db, nil
}

func (dm *DBManager) Close() error {
	return dm.dbPool.Close()
}

func (dm *DBManager) GetConnection() (*sql.DB, error) {
	return dm.dbPool, nil
}

func (dm *DBManager) Execute(query string, args ...interface{}) (sql.Result, error) {
	db, err := dm.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.Exec(query, args...)
}

func (dm *DBManager) QueryRow(query string, args ...interface{}) (*sql.Row, error) {
	db, err := dm.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.QueryRow(query, args...), nil
}

func (dm *DBManager) Query(query string, args ...interface{}) (*sql.Rows, error) {
	db, err := dm.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.Query(query, args...)
}

func (dm *DBManager) Begin() (*sql.Tx, error) {
	db, err := dm.GetConnection()
	if err != nil {
		return nil, err
	}
	return db.Begin()
}

func (dm *DBManager) ExecuteDeleteAll(tableName string) (int64, error) {
	query := fmt.Sprintf("DELETE FROM `%s`", tableName)
	return dm.ExecuteDelete(query)
}

func (dm *DBManager) ExecuteDeleteWithWhere(tableName string, conditions string, args ...interface{}) (int64, error) {
	if conditions == "" {
		return 0, fmt.Errorf("conditions cannot be empty")
	}

	query := fmt.Sprintf("DELETE FROM `%s` WHERE %s", tableName, conditions)
	return dm.ExecuteDelete(query, args...)
}

func (dm *DBManager) ExecuteDelete(query string, args ...interface{}) (int64, error) {
	return dm.ExecuterowsAffected(query, args...)
}

func (dm *DBManager) ExecuterowsAffected(query string, args ...interface{}) (int64, error) {
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

func (dm *DBManager) ExecuteUpdateWithWhere(tableName string, columns map[string]interface{}, conditions string, args ...interface{}) (int64, error) {
	if conditions == "" {
		return 0, fmt.Errorf("conditions cannot be empty")
	}

	setClauses := make([]string, 0, len(columns))
	setArgs := make([]interface{}, 0, len(columns))

	for column, value := range columns {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", column))
		setArgs = append(setArgs, value)
	}

	setClause := strings.Join(setClauses, ", ")
	whereClause := conditions
	allArgs := append(setArgs, args...)

	query := fmt.Sprintf("UPDATE `%s` SET %s WHERE %s", tableName, setClause, whereClause)
	return dm.ExecuterowsAffected(query, allArgs...)
}

func placeholderForValue(value interface{}) string {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return "?"
	default:
		return "?"
	}
}
func (dm *DBManager) ExecuteInsertOrUpdate(tableName string, data interface{}, uniqueKeyColumns []string) (int64, error) {
	// Build the INSERT INTO ... ON DUPLICATE KEY UPDATE ... SQL query
	query, values := buildInsertOrUpdateQuery(tableName, data, uniqueKeyColumns)

	db, err := dm.GetConnection()
	if err != nil {
		return 0, err
	}

	stmt, err := db.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare SQL statement: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(values...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute SQL statement: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve rows affected: %w", err)
	}

	return rowsAffected, nil
}

func buildInsertOrUpdateQuery(tableName string, data interface{}, uniqueKeyColumns []string) (string, []interface{}) {
	var valueFields, valuePlaceholders, updateFields []string
	values := reflect.ValueOf(data).Elem()
	var valuesArgs []interface{}

	for i := 0; i < values.NumField(); i++ {
		fieldName := values.Type().Field(i).Name
		valueFields = append(valueFields, fieldName)
		valuePlaceholders = append(valuePlaceholders, "?")
		updateFields = append(updateFields, fmt.Sprintf("%s=VALUES(%s)", fieldName, fieldName))
		valuesArgs = append(valuesArgs, values.Field(i).Interface())
	}

	// Build the SQL query
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON DUPLICATE KEY UPDATE %s",
		tableName,
		strings.Join(valueFields, ", "),
		strings.Join(valuePlaceholders, ", "),
		strings.Join(updateFields, ", "))

	return query, valuesArgs
}

func (dm *DBManager) ExecuteInsert(tableName string, data interface{}) (int64, error) {
	query := buildInsertQuery(tableName, data)

	db, err := dm.GetConnection()
	if err != nil {
		return 0, err
	}
	stmt, err := db.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare SQL statement: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(extractFieldValues(data)...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute SQL statement: %w", err)
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}

	return lastInsertID, nil
}

func buildInsertQuery(tableName string, data interface{}) string {
	var valueFields, valuePlaceholders []string
	values := reflect.ValueOf(data).Elem()

	for i := 0; i < values.NumField(); i++ {
		fieldName := values.Type().Field(i).Name
		valueFields = append(valueFields, fieldName)
		valuePlaceholders = append(valuePlaceholders, "?")
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(valueFields, ", "), strings.Join(valuePlaceholders, ", "))
}

func extractFieldValues(data interface{}) []interface{} {
	var fieldValues []interface{}
	values := reflect.ValueOf(data).Elem()

	for i := 0; i < values.NumField(); i++ {
		fieldValue := values.Field(i).Interface()
		fieldValues = append(fieldValues, fieldValue)
	}

	return fieldValues
}

func (dm *DBManager) ExecuteMultiInsert(tableName string, inserts []map[string]interface{}) (int64, error) {
	if len(inserts) == 0 {
		return 0, fmt.Errorf("no records to insert")
	}

	columns := make([]string, 0, len(inserts[0]))
	valuesPlaceholders := make([]string, 0)
	valuesArgs := make([]interface{}, 0, len(inserts)*len(columns))

	for column := range inserts[0] {
		columns = append(columns, column)
	}

	for _, record := range inserts {
		recordPlaceholders := make([]string, 0, len(columns))
		for _, columnName := range columns {
			value, ok := record[columnName]
			if !ok {
				return 0, fmt.Errorf("missing value for column %s", columnName)
			}
			recordPlaceholders = append(recordPlaceholders, "?")
			valuesArgs = append(valuesArgs, value)
		}
		valuesPlaceholders = append(valuesPlaceholders, "("+strings.Join(recordPlaceholders, ", ")+")")
	}

	query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s", tableName, strings.Join(columns, ", "), strings.Join(valuesPlaceholders, ", "))
	return dm.ExecuterowsAffected(query, valuesArgs...)
}
