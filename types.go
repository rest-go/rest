package main

import "database/sql"

const DEFAULT = "__default__"

type TypeConverter func(any) any

var Types = map[string]func() any{
	"BOOL":      func() any { return new(sql.NullBool) },
	"INTEGER":   func() any { return new(sql.NullInt32) },
	"INT4":      func() any { return new(sql.NullInt64) },
	"TIMESTAMP": func() any { return new(sql.NullTime) },
	DEFAULT:     func() any { return new(sql.NullString) },
}

var TypeConverters = map[string]TypeConverter{
	"BOOL":      func(i any) any { return i.(*sql.NullBool).Bool },
	"INTEGER":   func(i any) any { return i.(*sql.NullInt32).Int32 },
	"INT4":      func(i any) any { return i.(*sql.NullInt64).Int64 },
	"TIMESTAMP": func(i any) any { return i.(*sql.NullTime).Time },
	DEFAULT:     func(i any) any { return i.(*sql.NullString).String },
}
