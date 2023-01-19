package sql

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

type Helper interface {
	GetTablesSQL() string
	GetColumnsSQL(string) string
}

var helpers = map[string]Helper{
	"postgres": PGHelper{},
	"pgx":      PGHelper{},
	"mysql":    MyHelper{},
	"sqlite":   SQLiteHelper{},
}
