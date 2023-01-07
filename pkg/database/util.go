package database

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

var tableNameReg = regexp.MustCompile("^[_a-zA-Z][_a-zA-Z0-9]*$")

func IsValidTableName(tableName string) bool {
	return tableNameReg.MatchString(tableName)
}

// Open connects to database by specify database url and ping it
func Open(url string) (*sql.DB, error) {
	parts := strings.SplitN(url, "://", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid db url: %s", url)
	}

	driver, dsn := parts[0], parts[1]
	if driver == "postgres" {
		driver = "pgx"
		dsn = url
	}
	db, err := sql.Open(driver, dsn)
	if err == nil {
		err = db.Ping()
	}
	return db, err
}

func placeholder(driver string, index uint) string {
	switch strings.ToLower(driver) {
	case "mysql":
		return "?"
	default:
		return fmt.Sprintf("$%d", index)
	}
}
