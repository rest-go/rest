package sql

import (
	"database/sql"
	"regexp"
	"strconv"
	"strings"
)

type TypeConverter func(any) any

var (
	numericRegexp = regexp.MustCompile(`^(INT|FLOAT)\d+`)
	// Various data types
	// PG: https://www.postgresql.org/docs/current/datatype.html
	// MY: https://dev.mysql.com/doc/refman/8.0/en/data-types.html
	// SQLITE: https://www.sqlite.org/datatype3.html

	// TODO: benchmark performance of regexp match VS map access
	// the code below could be simplified by using regexp, but declare it in a
	// map should result in better performance in theory.
	Types = map[string]func() any{
		"TINYINT":     func() any { return new(sql.NullInt64) },
		"SMALLINT":    func() any { return new(sql.NullInt64) },
		"SMALLSERIAL": func() any { return new(sql.NullInt64) },
		"SERIAL":      func() any { return new(sql.NullInt64) },
		"INT":         func() any { return new(sql.NullInt64) },
		"INTEGER":     func() any { return new(sql.NullInt64) },
		"BIGINT":      func() any { return new(sql.NullInt64) },
		"BIGSERIAL":   func() any { return new(sql.NullInt64) },

		"DEC":              func() any { return new(sql.NullFloat64) },
		"DECIMAL":          func() any { return new(sql.NullFloat64) },
		"NUMERIC":          func() any { return new(sql.NullFloat64) },
		"FLOAT":            func() any { return new(sql.NullFloat64) },
		"REAL":             func() any { return new(sql.NullFloat64) },
		"DOUBLE":           func() any { return new(sql.NullFloat64) },
		"DOUBLE PRECISION": func() any { return new(sql.NullFloat64) },

		"BOOL":    func() any { return new(sql.NullBool) },
		"BOOLEAN": func() any { return new(sql.NullBool) },

		"JSON": func() any { return new(sql.NullString) },

		"CHAR":      func() any { return new(sql.NullString) },
		"VARCHAR":   func() any { return new(sql.NullString) },
		"NVARCHAR":  func() any { return new(sql.NullString) },
		"TEXT":      func() any { return new(sql.NullString) },
		"UUID":      func() any { return new(sql.NullString) },
		"ENUM":      func() any { return new(sql.NullString) },
		"BLOB":      func() any { return new(sql.NullString) },
		"BINARY":    func() any { return new(sql.NullString) },
		"XML":       func() any { return new(sql.NullString) },
		"DATE":      func() any { return new(sql.NullString) },
		"DATETIME":  func() any { return new(sql.NullString) },
		"TIMESTAMP": func() any { return new(sql.NullString) },
	}

	TypeConverters = map[string]TypeConverter{
		"TINYINT":     func(i any) any { return i.(*sql.NullInt64).Int64 },
		"SMALLINT":    func(i any) any { return i.(*sql.NullInt64).Int64 },
		"SMALLSERIAL": func(i any) any { return i.(*sql.NullInt64).Int64 },
		"SERIAL":      func(i any) any { return i.(*sql.NullInt64).Int64 },
		"INT":         func(i any) any { return i.(*sql.NullInt64).Int64 },
		"INTEGER":     func(i any) any { return i.(*sql.NullInt64).Int64 },
		"BIGINT":      func(i any) any { return i.(*sql.NullInt64).Int64 },
		"BIGSERIAL":   func(i any) any { return i.(*sql.NullInt64).Int64 },

		"DEC":              func(i any) any { return i.(*sql.NullFloat64).Float64 },
		"DECIMAL":          func(i any) any { return i.(*sql.NullFloat64).Float64 },
		"NUMERIC":          func(i any) any { return i.(*sql.NullFloat64).Float64 },
		"FLOAT":            func(i any) any { return i.(*sql.NullFloat64).Float64 },
		"REAL":             func(i any) any { return i.(*sql.NullFloat64).Float64 },
		"DOUBLE":           func(i any) any { return i.(*sql.NullFloat64).Float64 },
		"DOUBLE PRECISION": func(i any) any { return i.(*sql.NullFloat64).Float64 },

		"BOOL":    func(i any) any { return i.(*sql.NullBool).Bool },
		"BOOLEAN": func(i any) any { return i.(*sql.NullBool).Bool },

		"CHAR":      func(i any) any { return i.(*sql.NullString).String },
		"VARCHAR":   func(i any) any { return i.(*sql.NullString).String },
		"NVARCHAR":  func(i any) any { return i.(*sql.NullString).String },
		"TEXT":      func(i any) any { return i.(*sql.NullString).String },
		"UUID":      func(i any) any { return i.(*sql.NullString).String },
		"ENUM":      func(i any) any { return i.(*sql.NullString).String },
		"BLOB":      func(i any) any { return i.(*sql.NullString).String },
		"BINARY":    func(i any) any { return i.(*sql.NullString).String },
		"XML":       func(i any) any { return i.(*sql.NullString).String },
		"DATE":      func(i any) any { return i.(*sql.NullString).String },
		"DATETIME":  func(i any) any { return i.(*sql.NullString).String },
		"TIMESTAMP": func(i any) any { return i.(*sql.NullString).String },

		"JSON": func(i any) any {
			rawData := i.(*sql.NullString).String
			if s, err := strconv.ParseFloat(rawData, 64); err == nil {
				return s
			}
			if s, err := strconv.ParseBool(rawData); err == nil {
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
		"is":    " is ",
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
	} else {
		t = numericRegexp.ReplaceAllString(t, "${1}")
		if f, ok := Types[t]; ok {
			return f(), TypeConverters[t]
		}
	}

	return Types["JSON"](), TypeConverters["JSON"]
}

// normalize converts various type to standard type
// e.g. sqlite has NVARCHAR(70), NUMERIC(10,2) will be NVARCHAR and NEMERIC
// PG has INT4,INT8,FLOAT4 etc. will be INT, FLOAT
func normalize(t string) string {
	i := strings.Index(t, "(")
	if i != -1 {
		t = t[:i]
	}
	return strings.ToUpper(t)
}
