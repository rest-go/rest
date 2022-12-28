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

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

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
	table := strings.Trim(r.URL.Path, "/")

	var res Response
	switch r.Method {
	case "POST":
		res = s.create(r, table)
	case "DELETE":
		res = s.delete(r, table)
	case "PUT":
		res = s.update(r, table)
	case "GET":
		res = s.get(r, table)
	default:
		if _, err := w.Write([]byte(fmt.Sprintf("method not supported: %s", r.Method))); err != nil {
			log.Fatal(err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("failed to encode json data, %v", err)
	}
}

func (s *Service) create(r *http.Request, tableName string) Response {
	data := make(map[string]any)
	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("failed to parse http body as json data, %v", err),
		}
	}
	if len(data) == 0 {
		return Response{
			Code: http.StatusBadRequest,
			Msg:  "empty data",
		}
	}

	log.Printf("get json data: %+v", data)
	placeholders := make([]string, 0, len(data))
	names := make([]string, 0, len(data))
	args := make([]any, 0, len(data))
	index := 1
	for k, v := range data {
		names = append(names, k)
		args = append(args, v)
		placeholders = append(placeholders, fmt.Sprintf("$%d", index))
		index++
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(names, ","), strings.Join(placeholders, ","))
	rows, err := s.execQuery(r.Context(), sql, args)
	if err != nil {
		return Response{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		}
	}
	if rows != 1 {
		return Response{
			Code: http.StatusInternalServerError,
			Msg:  fmt.Sprintf("expected to affect 1 row, affected %d", rows),
		}
	}

	return Response{
		Code: http.StatusOK,
		Msg:  "success",
	}
}

func (s *Service) delete(r *http.Request, tableName string) Response {
	sql := fmt.Sprintf("DELETE FROM %s", tableName)
	_, query, args := buildWhereQuery(1, r.URL.Query())
	if len(query) > 0 {
		sql += fmt.Sprintf(" WHERE %s", query)
	}

	_, err := s.fetchData(r.Context(), sql, args...)
	if err != nil {
		return Response{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		}
	}

	return Response{
		Code: http.StatusOK,
		Msg:  "success",
	}
}

func (s *Service) update(r *http.Request, tableName string) Response {
	var sqlBuilder strings.Builder
	sqlBuilder.WriteString("UPDATE ")
	sqlBuilder.WriteString(tableName)

	data := make(map[string]any)
	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("failed to parse http body as json data, %v", err),
		}
	}
	if len(data) == 0 {
		return Response{
			Code: http.StatusBadRequest,
			Msg:  "empty data",
		}
	}
	log.Printf("get json data: %+v", data)

	index, setQuery, args := buildSetQuery(1, data)
	sqlBuilder.WriteString(" SET ")
	sqlBuilder.WriteString(setQuery)

	_, query, args2 := buildWhereQuery(index, r.URL.Query())
	if len(query) > 0 {
		sqlBuilder.WriteString(" WHERE ")
		sqlBuilder.WriteString(query)
		args = append(args, args2...)
	}

	rows, err := s.execQuery(r.Context(), sqlBuilder.String(), args)
	if err != nil {
		return Response{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		}
	}
	return Response{
		Code: http.StatusOK,
		Msg:  fmt.Sprintf("successfully updated %d rows", rows),
	}
}

func (s *Service) get(r *http.Request, tableName string) Response {
	// Example
	// fetch all:
	//   http get "localhost:8080/artists"
	// fetch by query:
	//   http get "localhost:8080/artists?ArtistId=1"
	// query by array:
	//   http get "localhost:8080/artists?ArtistId=1&ArtistId=2"

	sql := fmt.Sprintf("SELECT * FROM %s", tableName)
	_, query, args := buildWhereQuery(1, r.URL.Query())
	if len(query) > 0 {
		sql += fmt.Sprintf(" WHERE %s", query)
	}

	// TODO: handle page operations
	sql += " LIMIT 10"

	objects, err := s.fetchData(r.Context(), sql, args...)
	if err != nil {
		return Response{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		}
	}

	return Response{
		Code: http.StatusOK,
		Msg:  "success",
		Data: objects,
	}
}

// execQuery ...
func (s *Service) execQuery(ctx context.Context, sql string, args ...any) (int64, error) {
	log.Printf("exec query in database, sql: %v, args: %v", sql, args)
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to insert data to database, %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to parse database rows, %w", err)
	}
	return rows, nil
}

// fetchData ...
func (s *Service) fetchData(ctx context.Context, sql string, args ...any) ([]any, error) {
	log.Printf("fetch data in database, sql: %v, args: %v", sql, args)
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to run query in database, %w", err)
	}
	defer rows.Close()

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns from database, %v", err)
	}

	count := len(columnTypes)
	objects := []any{}
	for rows.Next() {
		scanArgs := make([]any, count)
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
			return nil, fmt.Errorf("failed to scan data from database, %v", err)
		}

		object := make(map[string]any, count)
		for i, v := range columnTypes {
			object[v.Name()] = converters[i](scanArgs[i])
		}
		objects = append(objects, object)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to fetch rows from database, %v", err)

	}
	return objects, nil
}
