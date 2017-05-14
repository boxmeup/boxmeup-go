package main

import (
	"database/sql"

	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// GetDBResource will get a reference to a pre-configured sql DB
func GetDBResource() (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%v?parseTime=true", config.MysqlDSN))
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
