package main

import (
	"database/sql"

	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
)

const (
	engine    = "InnoDB"
	encording = "UTF-8"
)

//conn DB接続を行います.
func conn(dbConfig *DBConfig) (*gorp.DbMap, error) {
	db, err := sql.Open(
		dbConfig.driver.String(),
		dbConfig.username+":"+dbConfig.password+
			"@tcp("+dbConfig.url+":"+dbConfig.port+")/"+dbConfig.database+"?parseTime=true")

	if err != nil {
		return nil, err
	}

	return &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{Engine: engine, Encoding: encording}}, nil
}
