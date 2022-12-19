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
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	table := strings.Trim(r.URL.Path, "/")
	// TODO: validate table name, check table exists in database
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf("select * from %s as t;", table))
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
		converters := make([]TypeConverter, count)
		for i, v := range columnTypes {
			t := v.DatabaseTypeName()
			if f, ok := Types[t]; ok {
				scanArgs[i] = f()
				converters[i] = TypeConverters[t]
			} else {
				scanArgs[i] = Types[DEFAULT]()
				converters[i] = TypeConverters[DEFAULT]
			}
		}

		err := rows.Scan(scanArgs...)
		if err != nil {
			log.Fatal(err)
		}

		data := make(map[string]interface{}, count)
		for i, v := range columnTypes {
			data[v.Name()] = converters[i](scanArgs[i])
		}
		finalRows = append(finalRows, data)
	}
	if err := json.NewEncoder(w).Encode(finalRows); err != nil {
		log.Fatal(err)
	}
}
