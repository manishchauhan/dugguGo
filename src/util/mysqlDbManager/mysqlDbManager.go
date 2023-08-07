package mysqlDbManager

import (
	"database/sql"
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
