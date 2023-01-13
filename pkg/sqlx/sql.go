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
	"strings"
)

type Column struct {
	ColumnName string `json:"column_name"`
	DataType   string `json:"data_type"`
}
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

type DB struct {
	*sql.DB
	DriverName string
}

// Open connects to database by specify database url and ping it
func Open(url string) (*DB, error) {
	parts := strings.SplitN(url, "://", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid db url: %s", url)
	}

	driver, dsn := parts[0], parts[1]
	if driver == "postgres" {
		driver = "pgx"
		dsn = url
	}
	db, err := sql.Open(driver, dsn)
	if err == nil {
		err = db.Ping()
	}
	return &DB{db, driver}, err
}

func (db *DB) Tables() []Table {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	helper := helpers[db.DriverName]
	query := helper.GetTablesSQL()
	rows, err := FetchData(ctx, db, query)
	if err != nil {
		log.Print("fetch tables error: ", err)
	}
	tables := make([]Table, 0, len(rows))
	for _, row := range rows {
		tableName := row["name"].(string)

		columnsQuery := helper.GetColumnsSQL(tableName)
		rows, err := FetchData(ctx, db, columnsQuery)
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

		tables = append(tables, Table{tableName, columns})
	}

	return tables
}
