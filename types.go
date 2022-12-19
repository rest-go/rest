package main

import "database/sql"

const DEFAULT = "__default__"

type TypeConverter func(interface{}) interface{}

var Types = map[string]func() interface{}{
	"BOOL":      func() interface{} { return new(sql.NullBool) },
	"INTEGER":   func() interface{} { return new(sql.NullInt32) },
	"INT4":      func() interface{} { return new(sql.NullInt64) },
	"TIMESTAMP": func() interface{} { return new(sql.NullTime) },
	DEFAULT:     func() interface{} { return new(sql.NullString) },
}

var TypeConverters = map[string]TypeConverter{
	"BOOL":      func(i interface{}) interface{} { return i.(*sql.NullBool).Bool },
	"INTEGER":   func(i interface{}) interface{} { return i.(*sql.NullInt32).Int32 },
	"INT4":      func(i interface{}) interface{} { return i.(*sql.NullInt64).Int64 },
	"TIMESTAMP": func(i interface{}) interface{} { return i.(*sql.NullTime).Time },
	DEFAULT:     func(i interface{}) interface{} { return i.(*sql.NullString).String },
}
