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

type PostData struct {
	objects []map[string]any
}

type PostQuery struct {
	index   int
	columns []string
	vals    []string
	args    []any
}

// valuesQuery convert post data to values query for insertion
func (pd *PostData) valuesQuery() (*PostQuery, error) {
	objects := pd.objects
	if len(objects) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	columns := make([]string, 0, len(objects[0]))
	vals := make([]string, 0, len(objects))
	args := make([]any, 0, cap(columns)*cap(vals))
	first := true
	index := 1
	for _, object := range objects {
		val := make([]string, 0, len(object))
		if first {
			for k, v := range object {
				columns = append(columns, k)
				val = append(val, fmt.Sprintf("$%d", index))
				args = append(args, v)
				index++
			}
			first = false
		} else {
			if !identKeys(object, columns) {
				return nil, fmt.Errorf("columns must be same for all objects, invalid object: %v", object)
			}
			// consistent column order with first object
			for _, c := range columns {
				val = append(val, fmt.Sprintf("$%d", index))
				args = append(args, object[c])
				index++
			}
		}
		vals = append(vals, fmt.Sprintf("(%s)", strings.Join(val, ",")))
	}
	return &PostQuery{
		index, columns, vals, args,
	}, nil
}

// UnmarshalJSON implements json.Unmarshaler
func (pd *PostData) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return fmt.Errorf("no bytes to unmarshal")
	}
	// See if we can guess based on the first character
	switch b[0] {
	case '{':
		return pd.unmarshalSingle(b)
	case '[':
		return pd.unmarshalMany(b)
	}
	// This shouldn't really happen as the standard library seems to strip
	// whitespace from the bytes being passed in, but just in case let's guess at
	// multiple tags and fall back to a single one if that doesn't work.
	err := pd.unmarshalMany(b)
	if err != nil {
		return pd.unmarshalSingle(b)
	}
	return nil
}

func (pd *PostData) unmarshalSingle(b []byte) error {
	var data map[string]any
	err := json.Unmarshal(b, &data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal single data, %v", err)
	}

	pd.objects = []map[string]any{data}
	return nil
}

func (pd *PostData) unmarshalMany(b []byte) error {
	var data []map[string]any
	err := json.Unmarshal(b, &data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal many data, %v", err)
	}

	pd.objects = data
	return nil
}

type Service struct {
	db     *sql.DB
	tables map[string]struct{}
}

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

func (s *Service) json(w http.ResponseWriter, res *Response) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("failed to encode json data, %v", err)
	}
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var res *Response
	table := strings.Trim(r.URL.Path, "/")
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

	rows, err := s.execQuery(r.Context(), sqlBuilder.String(), args...)
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

	rows, err := s.execQuery(r.Context(), sqlBuilder.String(), args...)
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

	objects, err := s.fetchData(r.Context(), sqlBuilder.String(), args...)
	if err != nil {
		return &Response{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		}
	}

	return &Response{
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
