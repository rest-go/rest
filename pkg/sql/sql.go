// package sqlx provides a is a wrap of golang database/sql package to hide
// logic for different drivers, the three main functions of this package are:
// 1. generate query from HTTP input
// 2. execute query against different SQL databases
// 3. provide helper functions to get meta information from database
package sql

import (
	"context"
	stdSQL "database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rest-go/rest/pkg/log"
)

type DriverType int

const (
	Postgres DriverType = iota
	MySQL
	SQLite
)

// DefaultTimeout for all database operations
const DefaultTimeout = 2 * time.Minute

// DB is a wrapper of the golang database/sql DB struct with a DriverName to
// handle generic logic against different SQL database
type DB struct {
	*stdSQL.DB
	DriverName string
}

// Open connects to database by specify database url and ping it
func Open(url string) (*DB, error) {
	parts := strings.SplitN(url, "://", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid db url, no driver: %s", url)
	}

	driverName, dsn := parts[0], parts[1]
	driver := driverName
	if driverName == "postgres" {
		driver = "pgx"
		dsn = url
	}
	if driver == "sqlite" {
		// increase busy timeout to avoid database lock in concurrent goroutine
		// https://www.sqlite.org/c3ref/busy_timeout.html
		// https://github.com/mattn/go-sqlite3/issues/274#issuecomment-232942571
		if !strings.Contains(dsn, "busy_timeout") {
			dsn += "?_pragma=busy_timeout(5000)"
		}
	}
	db, err := stdSQL.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open db error: %v", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("ping db error: %v", err)
	}
	dbb := &DB{DB: db, DriverName: driverName}
	return dbb, err
}

// fetchColumns fetch columns for a table
// Note: it doesn't use `fetchData` method because we want to control return
// data type by ourself
func (db *DB) fetchColumns(tableName string) ([]*Column, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	helper := helpers[db.DriverName]
	columnsQuery := helper.GetColumnsSQL(tableName)
	rows, err := db.QueryContext(ctx, columnsQuery)
	if err != nil {
		return nil, "", err
	}

	columns := []*Column{}
	primaryKey := ""
	hasPK := false
	for rows.Next() {
		var column Column
		if err := rows.Scan(&column.ColumnName, &column.DataType, &column.NotNull, &column.Pk); err != nil {
			return nil, "", err
		}
		if column.Pk {
			if !hasPK {
				primaryKey = column.ColumnName
				hasPK = true
			} else {
				// primary key on multiple columns is not suppored for now, clear it
				primaryKey = ""
			}
		}
		columns = append(columns, &column)
	}
	return columns, primaryKey, nil
}

// FetchTables return all the tables in current database along with all the columns
// name and datatype
func (db *DB) FetchTables() map[string]*Table {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	helper := helpers[db.DriverName]
	query := helper.GetTablesSQL()
	rows, err := db.FetchData(ctx, query)
	if err != nil {
		log.Errorf("fetch tables error: %v", err)
	}
	tables := make(map[string]*Table, len(rows))
	for _, row := range rows {
		tableName := row["name"].(string)
		columns, pk, err := db.fetchColumns(tableName)
		if err != nil {
			log.Errorf("fetch columns error %v, skip table %s", err, tableName)
			continue
		}
		tables[tableName] = &Table{tableName, pk, columns}
	}
	return tables
}

// ExecQuery execute and query in database and return rows affected or an error
func (db *DB) ExecQuery(ctx context.Context, query string, args ...any) (int64, error) {
	query = Rebind(db.DriverName, query)
	log.Debugf("exec query, query: %v, args: %v", query, args)
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, convertError("failed to exec sql", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, convertError("rows affected error", err)
	}
	return rows, nil
}

// FetchData execute query and fetch data from database, it always return an array
// or error
func (db *DB) FetchData(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	query = Rebind(db.DriverName, query)
	log.Debugf("fetch data, query: %v, args: %v", query, args)
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, convertError("failed to run query", err)
	}
	defer rows.Close()

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, convertError("failed to get columns from database", err)
	}

	columnCount := len(columnTypes)
	objects := []map[string]any{}
	for rows.Next() {
		scanArgs := make([]any, columnCount)
		converters := make([]TypeConverter, columnCount)
		for i, v := range columnTypes {
			t, converter := getTypeAndConverter(v.DatabaseTypeName())
			scanArgs[i] = t
			converters[i] = converter
		}
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, convertError("failed to scan data from database", err)
		}

		object := make(map[string]any, columnCount)
		for i, v := range columnTypes {
			object[v.Name()] = converters[i](scanArgs[i])
		}
		objects = append(objects, object)
	}
	if err = rows.Err(); err != nil {
		return nil, convertError("failed to fetch rows from database", err)
	}
	return objects, nil
}

// FetchOne execute query and fetch data from database, it returns one row or error
func (db *DB) FetchOne(ctx context.Context, query string, args ...any) (map[string]any, error) {
	objects, err := db.FetchData(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	if len(objects) == 0 {
		return nil, NewError(http.StatusNotFound, "not found")
	} else if len(objects) > 1 {
		return nil, NewError(http.StatusBadRequest, "multiple rows found in database")
	}

	return objects[0], nil
}
