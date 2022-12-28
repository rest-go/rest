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
	var res Response
	switch r.Method {
	case "POST":
		res = s.create(w, r)
	case "DELETE":
		res = s.delete(w, r)
	case "PUT":
		res = s.update(w, r)
	case "GET":
		res = s.get(w, r)
	default:
		if _, err := w.Write([]byte(fmt.Sprintf("method not supported: %s", r.Method))); err != nil {
			log.Fatal(err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("failed to encode json data, err: %v", err)
	}
}

func (s *Service) create(w http.ResponseWriter, r *http.Request) Response {
	table := strings.Trim(r.URL.Path, "/")

	data := make(map[string]any)
	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("failed to parse http body as json data, err: %v", err),
		}
	}

	log.Printf("get json data: %+v", data)
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()
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

	result, err := s.db.ExecContext(ctx,
		fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(names, ","), strings.Join(placeholders, ",")),
		args...,
	)
	if err != nil {

		return Response{
			Code: http.StatusInternalServerError,
			Msg:  fmt.Sprintf("failed to insert data to database, err: %v", err),
		}
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return Response{
			Code: http.StatusInternalServerError,
			Msg:  fmt.Sprintf("failed to parse  database respons, err: %v", err),
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

func (s *Service) delete(w http.ResponseWriter, r *http.Request) Response {
	return Response{
		Code: http.StatusOK,
		Msg:  "success",
	}
}

func (s *Service) update(w http.ResponseWriter, r *http.Request) Response {
	return Response{
		Code: http.StatusOK,
		Msg:  "success",
	}
}

func (s *Service) get(w http.ResponseWriter, r *http.Request) Response {
	// Example
	// fetch all:
	//   http get "localhost:8080/artists"
	// fetch by query:
	//   http get "localhost:8080/artists?ArtistId=1"
	// query by array:
	//   http get "localhost:8080/artists?ArtistId=1&ArtistId=2"

	tableName := strings.Trim(r.URL.Path, "/")
	// TODO: validate table exists in database
	sql := fmt.Sprintf("SELECT * FROM %s", tableName)

	query := r.URL.Query()
	args := make([]any, 0, len(query))
	if len(query) > 0 {
		// build where clause
		var queryBuilder strings.Builder
		count := 0
		index := 1
		for k, v := range query {
			v := filterEmpty(v)
			if len(v) == 0 {
				log.Printf("empty query: %s\n", k)
				continue
			}
			if count > 0 {
				queryBuilder.WriteString(" AND ")
			}
			queryBuilder.WriteString(k)
			if len(v) == 1 {
				queryBuilder.WriteString(" = ")
				queryBuilder.WriteString(fmt.Sprintf("$%d", index))
				args = append(args, v[0])
				index++
			} else {
				queryBuilder.WriteString(" in (")
				for i, vv := range v {
					queryBuilder.WriteString(fmt.Sprintf("$%d", index))
					if i != len(v)-1 {
						queryBuilder.WriteString(",")
					}
					args = append(args, vv)
					index++
				}
				queryBuilder.WriteString(")")

			}
			count++
		}
		s := queryBuilder.String()
		if len(s) > 0 {
			sql += fmt.Sprintf(" WHERE %s", s)
		}
	}

	// TODO: handle page operations
	sql += " LIMIT 10"

	log.Printf("sql: %v, args: %v", sql, args)
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

func (s *Service) fetchData(ctx context.Context, sql string, args ...any) ([]any, error) {
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to run query in database, err: %w", err)
	}
	defer rows.Close()

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns from database, err: %v", err)
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
			return nil, fmt.Errorf("failed to scan data from database, err: %v", err)
		}

		object := make(map[string]any, count)
		for i, v := range columnTypes {
			object[v.Name()] = converters[i](scanArgs[i])
		}
		objects = append(objects, object)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to fetch rows from database, err: %v", err)

	}
	return objects, nil
}

func filterEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
