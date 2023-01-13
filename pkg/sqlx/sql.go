// package sqlx provides a is a wrap of golang database/sql package to hide
// logic for different drivers, the three main functions of this package is"
// 1. generate query from HTTP input
// 2. execute query against different SQL databases
// 3. provide helper functions to get meta information from database
package sqlx

import (
	"database/sql"
	"fmt"
	"strings"
)

type DB struct {
	*sql.DB
	DriverName string
}

// Open connects to database by specify database url and ping it
func Open(url string) (*DB, error) {
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
	return &DB{db, driver}, err
}
