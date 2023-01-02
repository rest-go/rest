package database

import (
	"regexp"
)

var tableNameReg = regexp.MustCompile("^[_a-zA-Z][_a-zA-Z0-9]*$")

func IsValidTableName(tableName string) bool {
	return tableNameReg.MatchString(tableName)
}
