package sqlx

import (
	"regexp"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

var tableNameReg = regexp.MustCompile("^[_a-zA-Z][_a-zA-Z0-9]*$")

func IsValidTableName(tableName string) bool {
	return tableNameReg.MatchString(tableName)
}
