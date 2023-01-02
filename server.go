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

	"github.com/shellfly/rest/pkg/database"
)

// A Response serves JSON output for all restful apis
type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

// Server is the representation of a restful server handler which handles CRUD operations.
type Server struct {
	db     *sql.DB
	tables map[string]struct{}
}

// TODO: accept function options to config database
// NewServer returns a Service pointer.
func NewServer(url string, limitedTables ...string) *Server {
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

	return &Server{
		db:     db,
		tables: tables,
	}
}

func (s *Server) debug(sql string, args ...any) *Response {
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

func (s *Server) json(w http.ResponseWriter, res *Response) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("failed to encode json data, %v", err)
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	table := strings.Trim(r.URL.Path, "/")
	if table == "" {
		res := &Response{
			Code: http.StatusOK,
			Msg:  "rest server is up and running",
		}
		s.json(w, res)
		return
	}
	if !database.IsValidTableName(table) {
		res := &Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("invalid table name: %s", table),
		}
		s.json(w, res)
		return
	}

	if len(s.tables) > 0 {
		if _, ok := s.tables[table]; !ok {
			res := &Response{
				Code: http.StatusNotFound,
				Msg:  fmt.Sprintf("table not supported: %s", table),
			}
			s.json(w, res)
			return
		}
	}

	var res *Response
	query := database.Query(r.URL.Query())
	switch r.Method {
	case "POST":
		res = s.create(r, table, query)
	case "DELETE":
		res = s.delete(r, table, query)
	case "PUT", "PATCH":
		res = s.update(r, table, query)
	case "GET":
		res = s.get(r, table, query)
	default:
		res = &Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("method not supported: %s", r.Method),
		}
	}
	s.json(w, res)
}

func (s *Server) create(r *http.Request, tableName string, query database.Query) *Response {
	var data database.PostData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return &Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("failed to parse post json data, %v", err),
		}
	}
	valuesQuery, err := data.ValuesQuery()
	if err != nil {
		return &Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("failed to prepare values query, %v", err),
		}
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", tableName, strings.Join(valuesQuery.Columns, ","), strings.Join(valuesQuery.Vals, ","))
	args := valuesQuery.Args
	if query.IsDebug() {
		return s.debug(sql, args...)
	}

	rows, err := s.execQuery(r.Context(), sql, args...)
	if err != nil {
		return &Response{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		}
	}
	if rows != int64(len(valuesQuery.Vals)) {
		return &Response{
			Code: http.StatusInternalServerError,
			Msg:  fmt.Sprintf("expected to insert %d rows, but affected %d rows", len(valuesQuery.Vals), rows),
		}
	}

	return &Response{
		Code: http.StatusOK,
		Msg:  "success",
	}
}

func (s *Server) delete(r *http.Request, tableName string, query database.Query) *Response {
	var sqlBuilder strings.Builder
	sqlBuilder.WriteString("DELETE FROM ")
	sqlBuilder.WriteString(tableName)
	_, whereQuery, args := query.WhereQuery(1)
	if whereQuery != "" {
		sqlBuilder.WriteString(" WHERE ")
		sqlBuilder.WriteString(whereQuery)
	}

	sql := sqlBuilder.String()
	if query.IsDebug() {
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

func (s *Server) update(r *http.Request, tableName string, query database.Query) *Response {
	var data database.PostData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return &Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("failed to parse update json data, %v", err),
		}
	}
	setQuery, err := data.SetQuery(1)
	if err != nil {
		return &Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("failed to prepare set query, %v", err),
		}
	}

	var sqlBuilder strings.Builder
	sqlBuilder.WriteString(fmt.Sprintf("UPDATE %s SET %s", tableName, setQuery.Sql))

	args := setQuery.Args
	_, whereQuery, args2 := query.WhereQuery(setQuery.Index)
	if whereQuery != "" {
		sqlBuilder.WriteString(" WHERE ")
		sqlBuilder.WriteString(whereQuery)
		args = append(args, args2...)
	}

	sql := sqlBuilder.String()
	if query.IsDebug() {
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

func (s *Server) get(r *http.Request, tableName string, query database.Query) *Response {
	if query.IsCount() {
		return s.count(r, tableName)
	}

	var sqlBuilder strings.Builder
	selects := query.SelectQuery()
	sqlBuilder.WriteString(fmt.Sprintf("SELECT %s FROM %s", selects, tableName))
	_, whereQuery, args := query.WhereQuery(1)
	if whereQuery != "" {
		sqlBuilder.WriteString(" WHERE ")
		sqlBuilder.WriteString(whereQuery)
	}

	// order
	order := query.OrderQuery()
	if len(order) > 0 {
		sqlBuilder.WriteString(" ORDER BY ")
		sqlBuilder.WriteString(order)
	}

	// page operation
	page, pageSize := query.Page()
	sqlBuilder.WriteString(" LIMIT ")
	sqlBuilder.WriteString(fmt.Sprintf("%d", pageSize))
	if page != 1 {
		sqlBuilder.WriteString(" OFFSET ")
		sqlBuilder.WriteString(fmt.Sprintf("%d", (page-1)*pageSize))
	}

	sql := sqlBuilder.String()
	if query.IsDebug() {
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
	if query.IsSingular() {
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

func (s *Server) count(r *http.Request, tableName string) *Response {
	sql := fmt.Sprintf("SELECT COUNT(1) AS count FROM %s", tableName)
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
func (s *Server) execQuery(ctx context.Context, sql string, args ...any) (int64, error) {
	log.Printf("exec query in database, sql: %v, args: %v", sql, args)
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to exec sql, %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows, %w", err)
	}
	return rows, nil
}

// fetchData ...
func (s *Server) fetchData(ctx context.Context, sql string, args ...any) ([]any, error) {
	log.Printf("fetch data in database, sql: %v, args: %v", sql, args)
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to run query, %w", err)
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
		converters := make([]database.TypeConverter, count)
		for i, v := range columnTypes {
			t, converter := database.GetTypeAndConverter(v.DatabaseTypeName())
			scanArgs[i] = t
			converters[i] = converter
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
