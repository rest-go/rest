// package sqlx provides a is a wrap of golang database/sql package to hide
// logic for different drivers, the three main functions of this package are:
// 1. generate query from HTTP input
// 2. execute query against different SQL databases
// 3. provide helper functions to get meta information from database
package sqlx

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type DriverType int

const (
	Postgres DriverType = iota
	MySQL
	SQLite
)

// DefaultTimeout for all database operations
const DefaultTimeout = 2 * time.Minute

// Column represents a table column with name and type
type Column struct {
	ColumnName string `json:"column_name"`
	DataType   string `json:"data_type"`
}

// Table represents a table in database with name and columns
type Table struct {
	Name    string
	Columns []Column
}

func (t Table) String() string {
	var columnsBuilder strings.Builder
	for i, c := range t.Columns {
		columnsBuilder.WriteString(c.ColumnName)
		columnsBuilder.WriteString(" ")
		columnsBuilder.WriteString(c.DataType)
		if i < len(t.Columns)-1 {
			columnsBuilder.WriteString(",\n")
		}
	}
	return fmt.Sprintf("%s (%s)", t.Name, columnsBuilder.String())
}

// DB is a wrapper of the golang database/sql DB struct with a DriverName to
// handle generic logic against different SQL database
type DB struct {
	*sql.DB
	DriverName string
}

func New(db *sql.DB, driverName string) *DB {
	return &DB{db, driverName}
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
	db, err := sql.Open(driver, dsn)
	if err == nil {
		err = db.Ping()
	}
	return &DB{db, driverName}, err
}

// Tables return all the tables in current database along with all the columns
// name and datatype
func (db *DB) Tables() map[string]Table {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	helper := helpers[db.DriverName]
	query := helper.GetTablesSQL()
	rows, err := db.FetchData(ctx, query)
	if err != nil {
		log.Print("fetch tables error: ", err)
	}
	tables := make(map[string]Table, len(rows))
	for _, row := range rows {
		tableName := row["name"].(string)

		columnsQuery := helper.GetColumnsSQL(tableName)
		rows, err := db.FetchData(ctx, columnsQuery)
		if err != nil {
			log.Printf("fetch columns error %v, skip table %s", err, tableName)
			continue
		}

		columns := make([]Column, 0, len(rows))
		var columnErr error
		for _, row := range rows {
			data, err := json.Marshal(row)
			if err != nil {
				columnErr = err
				break
			}

			column := Column{}
			if err := json.Unmarshal(data, &column); err != nil {
				columnErr = err
				break
			}
			columns = append(columns, column)
		}
		if columnErr != nil {
			log.Printf("get columns error %v, skip table %s", columnErr, tableName)
			continue
		}

		tables[tableName] = Table{tableName, columns}
	}

	return tables
}

// ExecQuery execute and query in database and return rows affected or an error
func (db *DB) ExecQuery(ctx context.Context, query string, args ...any) (int64, error) {
	query = Rebind(db.DriverName, query)
	log.Printf("exec query, query: %v, args: %v", query, args)
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
	log.Printf("fetch data, query: %v, args: %v", query, args)
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

	count := len(columnTypes)
	objects := []map[string]any{}
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
			return nil, convertError("failed to scan data from database", err)
		}

		object := make(map[string]any, count)
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
		return nil, NewError(http.StatusNotFound, "data not found in database")
	} else if len(objects) > 1 {
		return nil, NewError(http.StatusBadRequest, "multiple rows found in database")
	}

	return objects[0], nil
}
