package mysqldb

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func InitMysql(dsn string) *sql.DB {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		slog.Error("FAILED TO OPEN DB", "ERROR", err.Error())
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		slog.Error("FAILED TO CONNECT TO DB", "ERROR", err.Error())
		panic(err)
	}

	return db
}
