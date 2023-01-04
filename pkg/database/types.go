package database

import (
	"database/sql"
	"strings"
)

const DEFAULT = "__default__"

type TypeConverter func(any) any

var Types = map[string]func() any{
	"BOOL":      func() any { return new(sql.NullBool) },
	"INTEGER":   func() any { return new(sql.NullInt32) },
	"INT4":      func() any { return new(sql.NullInt64) },
	"TIMESTAMP": func() any { return new(sql.NullTime) },
	"NUMERIC":   func() any { return new(sql.NullFloat64) },
	DEFAULT:     func() any { return new(sql.NullString) },
}

var TypeConverters = map[string]TypeConverter{
	"BOOL":      func(i any) any { return i.(*sql.NullBool).Bool },
	"INTEGER":   func(i any) any { return i.(*sql.NullInt32).Int32 },
	"INT4":      func(i any) any { return i.(*sql.NullInt64).Int64 },
	"TIMESTAMP": func(i any) any { return i.(*sql.NullTime).Time },
	"NUMERIC":   func(i any) any { return i.(*sql.NullFloat64).Float64 },
	DEFAULT:     func(i any) any { return i.(*sql.NullString).String },
}

var Operators = map[string]string{
	"eq":    " = ",
	"ne":    " <> ",
	"gt":    " > ",
	"lt":    " < ",
	"gte":   " >= ",
	"lte":   " <= ",
	"like":  " like ",
	"ilike": " ilike ",
	"in":    " in ",
	"cs":    " @> ",
	"cd":    " <@ ",
}

var ReservedWords = map[string]struct{}{
	"select": {},
	"order":  {},
	"count":  {},
}

func GetTypeAndConverter(t string) (any, TypeConverter) {
	t = normalize(t)
	if f, ok := Types[t]; ok {
		return f(), TypeConverters[t]
	}

	return Types[DEFAULT](), TypeConverters[DEFAULT]
}

// normalize converts various type to standard type
// e.g. sqlite has NVARCHAR(70), NUMERIC(10,2) will be NVARCHAR and NEMERIC
func normalize(t string) string {
	i := strings.Index(t, "(")
	if i != -1 {
		return t[:i]
	}
	return t
}
