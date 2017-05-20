package database

import (
	"database/sql"

	"fmt"

	"github.com/cjsaylor/boxmeup-go/modules/config"
	_ "github.com/go-sql-driver/mysql"
)

// GetDBResource will get a reference to a pre-configured sql DB
func GetDBResource() (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%v?parseTime=true", config.Config.MysqlDSN))
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return db, nil
}
