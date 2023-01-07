package database

import (
	"database/sql"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type TypeConverter func(any) any

const (
	StrType    = "StringType"
	GoldenType = "GoldenType"
)

var (
	strTypeRegexp = regexp.MustCompile("CHAR|TEXT|UUID|ENUM|BINARY|CLOB|BLOB|JSON|XML|DATETIME|TIMESTAMP")

	// Various data types
	// PG: https://www.postgresql.org/docs/current/datatype.html
	// MY: https://dev.mysql.com/doc/refman/8.0/en/data-types.html
	// SQLITE: https://www.sqlite.org/datatype3.html
	Types = map[string]func() any{
		"BIT":         func() any { return new(sql.NullInt16) },
		"TINYINT":     func() any { return new(sql.NullInt16) },
		"SMALLINT":    func() any { return new(sql.NullInt16) },
		"SMALLSERIAL": func() any { return new(sql.NullInt16) },
		"SERIAL":      func() any { return new(sql.NullInt32) },
		"INT":         func() any { return new(sql.NullInt32) },
		"INTEGER":     func() any { return new(sql.NullInt32) },
		"BIGINT":      func() any { return new(sql.NullInt64) },
		"BIGSERIAL":   func() any { return new(sql.NullInt64) },

		"DECIMAL":          func() any { return new(sql.NullFloat64) },
		"NUMERIC":          func() any { return new(sql.NullFloat64) },
		"FLOAT":            func() any { return new(sql.NullFloat64) },
		"REAL":             func() any { return new(sql.NullFloat64) },
		"DOUBLE PRECISION": func() any { return new(sql.NullFloat64) },

		"BOOL":    func() any { return new(sql.NullBool) },
		"BOOLEAN": func() any { return new(sql.NullBool) },

		StrType:    func() any { return new(sql.NullString) },
		GoldenType: func() any { return new(sql.NullString) },
	}

	TypeConverters = map[string]TypeConverter{
		"BIT":         func(i any) any { return i.(*sql.NullInt16).Int16 },
		"TINYINT":     func(i any) any { return i.(*sql.NullInt16).Int16 },
		"SMALLINT":    func(i any) any { return i.(*sql.NullInt16).Int16 },
		"SMALLSERIAL": func(i any) any { return i.(*sql.NullInt16).Int16 },
		"SERIAL":      func(i any) any { return i.(*sql.NullInt32).Int32 },
		"INT":         func(i any) any { return i.(*sql.NullInt32).Int32 },
		"INTEGER":     func(i any) any { return i.(*sql.NullInt32).Int32 },
		"BIGINT":      func(i any) any { return i.(*sql.NullInt64).Int64 },
		"BIGSERIAL":   func(i any) any { return i.(*sql.NullInt64).Int64 },

		"DECIMAL":          func(i any) any { return i.(*sql.NullFloat64).Float64 },
		"NUMERIC":          func(i any) any { return i.(*sql.NullFloat64).Float64 },
		"FLOAT":            func(i any) any { return i.(*sql.NullFloat64).Float64 },
		"REAL":             func(i any) any { return i.(*sql.NullFloat64).Float64 },
		"DOUBLE PRECISION": func(i any) any { return i.(*sql.NullFloat64).Float64 },

		"BOOL":    func(i any) any { return i.(*sql.NullBool).Bool },
		"BOOLEAN": func(i any) any { return i.(*sql.NullBool).Bool },

		StrType: func(i any) any { return i.(*sql.NullString).String },
		GoldenType: func(i any) any {
			rawData := i.(*sql.NullString).String
			if s, err := strconv.ParseFloat(rawData, 64); err == nil {
				return s
			}
			return rawData
		},
	}

	Operators = map[string]string{
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

	ReservedWords = map[string]struct{}{
		"select": {},
		"order":  {},
		"count":  {},
	}
)

func getTypeAndConverter(t string) (any, TypeConverter) {
	t = normalize(t)
	if f, ok := Types[t]; ok {
		return f(), TypeConverters[t]
	}

	if strTypeRegexp.MatchString(t) {
		return Types[StrType](), TypeConverters[StrType]
	}

	log.Printf("unrecognized type: %s, using golden type", t)
	return Types[GoldenType](), TypeConverters[GoldenType]
}

// normalize converts various type to standard type
// e.g. sqlite has NVARCHAR(70), NUMERIC(10,2) will be NVARCHAR and NEMERIC
func normalize(t string) string {
	i := strings.Index(t, "(")
	if i != -1 {
		t = t[:i]
	}

	return strings.ToUpper(t)
}
