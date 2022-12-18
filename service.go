package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

type Service struct {
	db *sql.DB
}

func NewService() *Service {
	// Opening a driver typically will not attempt to connect to the database.
	db, err := sql.Open("sqlite", "/Users/hao/Downloads/chinook.db")
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)

	return &Service{
		db: db,
	}
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db := s.db
	table := strings.Trim(r.URL.Path, "/")
	log.Println("get table: ", db, table)

	// This is a long SELECT. Use the request context as the base of
	// the context timeout, but give it some time to finish. If
	// the client cancels before the query is done the query will also
	// be canceled.
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, fmt.Sprintf("select * from %s as t;", table))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		log.Fatal(err)
	}

	count := len(columnTypes)
	finalRows := []interface{}{}

	for rows.Next() {
		scanArgs := make([]interface{}, count)
		for i, v := range columnTypes {
			switch v.DatabaseTypeName() {
			case "TIMESTAMP":
				scanArgs[i] = new(sql.NullTime)
			case "BOOL":
				scanArgs[i] = new(sql.NullBool)
			case "INT4":
				scanArgs[i] = new(sql.NullInt64)
			default:
				scanArgs[i] = new(sql.NullString)
			}
		}

		err := rows.Scan(scanArgs...)

		if err != nil {
			log.Fatal(err)
		}

		masterData := map[string]interface{}{}

		for i, v := range columnTypes {

			if z, ok := (scanArgs[i]).(*sql.NullBool); ok {
				masterData[v.Name()] = z.Bool
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullString); ok {
				masterData[v.Name()] = z.String
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt64); ok {
				masterData[v.Name()] = z.Int64
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullFloat64); ok {
				masterData[v.Name()] = z.Float64
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt32); ok {
				masterData[v.Name()] = z.Int32
				continue
			}

			masterData[v.Name()] = scanArgs[i]
		}

		finalRows = append(finalRows, masterData)
	}
	if err := json.NewEncoder(w).Encode(finalRows); err != nil {
		log.Fatal(err)
	}
}
