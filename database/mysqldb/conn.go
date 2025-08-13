package mysqldb

import (
	"database/sql"
	"log/slog"

	_ "github.com/go-sql-driver/mysql"
)

func InitMysql(dsn string) *sql.DB {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		slog.Error("FAILED TO OPEN DB", "ERROR", err.Error())
		panic(err)
	}

	if err = db.Ping(); err != nil {
		slog.Error("FAILED TO CONNECT TO DB", "ERROR", err.Error())
		panic(err)
	}

	return db
}
