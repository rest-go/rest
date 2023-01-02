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
	db     *sql.DB
	tables map[string]struct{}
}

// TODO: accept db config
func NewService(url string, limitedTables ...string) *Service {
	// Opening a driver typically will not attempt to connect to the database.
	parts := strings.SplitN(url, "://", 2)
	if len(parts) != 2 {
		log.Fatal("invalid db url: ", url)
	}
	log.Print("open db url: ", url)
	driver, dsn := parts[0], parts[1]
	if driver == "postgres" {
		driver = "pgx"
		dsn = url
	}
	db, err := sql.Open(driver, dsn)
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)

	var tables map[string]struct{}
	if len(limitedTables) > 0 {
		tables = make(map[string]struct{}, len(limitedTables))
		for _, t := range limitedTables {
			tables[t] = struct{}{}
		}
	}

	return &Service{
		db:     db,
		tables: tables,
	}
}

func (s *Service) debug(sql string, args ...any) *Response {
	return &Response{
		Code: http.StatusOK,
		Msg:  "success",
		Data: struct {
			Sql  string `json:"sql"`
			Args []any  `json:"args"`
		}{
			sql, args,
		},
	}
}

func (s *Service) json(w http.ResponseWriter, res *Response) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("failed to encode json data, %v", err)
	}
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var res *Response
	table := strings.Trim(r.URL.Path, "/")
	if table == "" {
		res = &Response{
			Code: http.StatusOK,
			Msg:  "rest server is up and running",
		}
		s.json(w, res)
		return
	}
	if !isValidTableName(table) {
		res = &Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("invalid table name: %s", table),
		}
		s.json(w, res)
		return
	}

	if len(s.tables) > 0 {
		if _, ok := s.tables[table]; !ok {
			res = &Response{
				Code: http.StatusNotFound,
				Msg:  fmt.Sprintf("table not supported: %s", table),
			}
			s.json(w, res)
			return
		}
	}

	switch r.Method {
	case "POST":
		res = s.create(r, table)
	case "DELETE":
		res = s.delete(r, table)
	case "PUT", "PATCH":
		res = s.update(r, table)
	case "GET":
		res = s.get(r, table)
	default:
		res = &Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("method not supported: %s", r.Method),
		}
	}
	s.json(w, res)
}

func (s *Service) create(r *http.Request, tableName string) *Response {
	var data PostData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return &Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("failed to parse http body as json data, %v", err),
		}
	}
	query, err := data.valuesQuery()
	if err != nil {
		return &Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("failed to prepare sql query, %v", err),
		}
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", tableName, strings.Join(query.columns, ","), strings.Join(query.vals, ","))
	if isDebug(r.URL.Query()) {
		return s.debug(sql)
	}
	rows, err := s.execQuery(r.Context(), sql, query.args...)
	if err != nil {
		return &Response{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		}
	}
	if rows != int64(len(query.vals)) {
		return &Response{
			Code: http.StatusInternalServerError,
			Msg:  fmt.Sprintf("expected to insert %d rows, but affected %d rows", len(query.vals), rows),
		}
	}

	return &Response{
		Code: http.StatusOK,
		Msg:  "success",
	}
}

func (s *Service) delete(r *http.Request, tableName string) *Response {
	var sqlBuilder strings.Builder
	sqlBuilder.WriteString("DELETE FROM ")
	sqlBuilder.WriteString(tableName)
	_, query, args := buildWhereQuery(1, r.URL.Query())
	if len(query) > 0 {
		sqlBuilder.WriteString(" WHERE ")
		sqlBuilder.WriteString(query)
	}

	sql := sqlBuilder.String()
	if isDebug(r.URL.Query()) {
		return s.debug(sql, args...)
	}

	rows, err := s.execQuery(r.Context(), sql, args...)
	if err != nil {
		return &Response{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		}
	}

	return &Response{
		Code: http.StatusOK,
		Msg:  fmt.Sprintf("successfully deleted %d rows", rows),
	}
}

func (s *Service) update(r *http.Request, tableName string) *Response {
	var sqlBuilder strings.Builder
	sqlBuilder.WriteString("UPDATE ")
	sqlBuilder.WriteString(tableName)

	data := make(map[string]any)
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return &Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("failed to parse http body as json data, %v", err),
		}
	}
	if len(data) == 0 {
		return &Response{
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

	sql := sqlBuilder.String()
	if isDebug(r.URL.Query()) {
		return s.debug(sql, args...)
	}
	rows, err := s.execQuery(r.Context(), sql, args...)
	if err != nil {
		return &Response{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		}
	}
	return &Response{
		Code: http.StatusOK,
		Msg:  fmt.Sprintf("successfully updated %d rows", rows),
	}
}

func (s *Service) get(r *http.Request, tableName string) *Response {
	// Example
	// fetch all:
	//   http get "localhost:8080/artists"
	// fetch by query:
	//   http get "localhost:8080/artists?ArtistId=eq.1"
	// query by array:
	//   http get "localhost:8080/artists?ArtistId=in.(1,2)"
	values := r.URL.Query()
	if _, ok := values["count"]; ok {
		return s.count(r, tableName)
	}

	var sqlBuilder strings.Builder
	selects := buildSelects(values)
	sqlBuilder.WriteString(fmt.Sprintf("SELECT %s FROM %s", selects, tableName))
	_, query, args := buildWhereQuery(1, values)
	if len(query) > 0 {
		sqlBuilder.WriteString(" WHERE ")
		sqlBuilder.WriteString(query)
	}

	// order
	order := buildOrderQuery(values)
	if len(order) > 0 {
		sqlBuilder.WriteString(" ORDER BY ")
		sqlBuilder.WriteString(order)
	}

	// page operation
	page, pageSize := extractPage(values)
	sqlBuilder.WriteString(" LIMIT ")
	sqlBuilder.WriteString(fmt.Sprintf("%d", pageSize))
	if page != 1 {
		sqlBuilder.WriteString(" OFFSET ")
		sqlBuilder.WriteString(fmt.Sprintf("%d", (page-1)*pageSize))
	}

	sql := sqlBuilder.String()
	if isDebug(r.URL.Query()) {
		return s.debug(sql, args...)
	}
	objects, err := s.fetchData(r.Context(), sql, args...)
	if err != nil {
		return &Response{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		}
	}
	var data any = objects
	if _, ok := values["singular"]; ok {
		if len(objects) != 1 {
			return &Response{
				Code: http.StatusBadRequest,
				Msg:  fmt.Sprintf("expect singular data, but got %d rows", len(objects)),
			}
		}
		data = objects[0]
	}
	return &Response{
		Code: http.StatusOK,
		Msg:  "success",
		Data: data,
	}
}

func (s *Service) count(r *http.Request, tableName string) *Response {
	sql := fmt.Sprintf("SELECT COUNT(1) as count FROM %s", tableName)
	rows, err := s.fetchData(r.Context(), sql)
	if err != nil {
		return &Response{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		}
	}

	return &Response{
		Code: http.StatusOK,
		Msg:  "success",
		Data: rows[0],
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
