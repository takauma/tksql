package tksql

import (
	"database/sql"

	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
)

const (
	defaultEngine    = "InnoDB"
	defaultEncording = "UTF-8"
)

// conn DB接続を行います.
func conn(dbConfig *DBConfig) (*gorp.DbMap, error) {
	db, err := sql.Open(
		dbConfig.driver.String(),
		dbConfig.username+":"+dbConfig.password+
			"@tcp("+dbConfig.url+":"+dbConfig.port+")/"+dbConfig.database+"?parseTime=true")

	if err != nil {
		return nil, err
	}

	if len(dbConfig.engine) == 0 {
		dbConfig.engine = defaultEngine
	}
	if len(dbConfig.encoding) == 0 {
		dbConfig.encoding = defaultEngine
	}

	return &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{Engine: defaultEngine, Encoding: defaultEncording}}, nil
}
