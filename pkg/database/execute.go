package database

import (
	"context"
	"database/sql"
	"log"
	"time"
)

const DefaultTimeout = 2 * time.Minute

// ExecQuery execute a sql query and return rows affected or an error
func ExecQuery(ctx context.Context, db *sql.DB, query string, args ...any) (int64, *Error) {
	log.Printf("exec query, query: %v, args: %v", query, args)
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, NewError("failed to exec sql", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, NewError("failed to get affected rows", err)
	}
	return rows, nil
}

// FetchData execute a sql and return matched rows or an error
func FetchData(ctx context.Context, db *sql.DB, query string, args ...any) ([]any, *Error) {
	log.Printf("fetch data, query: %v, args: %v", query, args)
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, NewError("failed to run query", err)
	}
	defer rows.Close()

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, NewError("failed to get columns from database", err)
	}

	count := len(columnTypes)
	objects := []any{}
	for rows.Next() {
		scanArgs := make([]any, count)
		converters := make([]TypeConverter, count)
		for i, v := range columnTypes {
			t, converter := getTypeAndConverter(v.DatabaseTypeName())
			scanArgs[i] = t
			converters[i] = converter
		}
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, NewError("failed to scan data from database", err)
		}

		object := make(map[string]any, count)
		for i, v := range columnTypes {
			object[v.Name()] = converters[i](scanArgs[i])
		}
		objects = append(objects, object)
	}
	if err = rows.Err(); err != nil {
		return nil, NewError("failed to fetch rows from database", err)
	}
	return objects, nil
}
