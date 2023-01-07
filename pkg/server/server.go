package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

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
	db *sql.DB
}

// TODO: accept function options to config database
// NewServer returns a Service pointer.
func NewServer(url string) *Server {
	log.Printf("connecting to database: %s", url)
	db, err := database.Open(url)
	if err != nil {
		log.Fatal(err)
	}
	defaultIdleConns := 50
	defaultOpenConns := 50
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(defaultIdleConns)
	db.SetMaxOpenConns(defaultOpenConns)

	return &Server{db}
}

func (s *Server) debug(query string, args ...any) *Response {
	return &Response{
		Code: http.StatusOK,
		Msg:  "success",
		Data: struct {
			Query string `json:"query"`
			Args  []any  `json:"args"`
		}{
			query, args,
		},
	}
}

func (s *Server) json(w http.ResponseWriter, res *Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(res.Code)
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

	var res *Response
	urlQuery := database.URLQuery(r.URL.Query())
	switch r.Method {
	case "POST":
		res = s.create(r, table, urlQuery)
	case "DELETE":
		res = s.delete(r, table, urlQuery)
	case "PUT", "PATCH":
		res = s.update(r, table, urlQuery)
	case "GET":
		res = s.get(r, table, urlQuery)
	default:
		res = &Response{
			Code: http.StatusMethodNotAllowed,
			Msg:  fmt.Sprintf("method not supported: %s", r.Method),
		}
	}
	s.json(w, res)
}

func (s *Server) create(r *http.Request, tableName string, urlQuery database.URLQuery) *Response {
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

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		tableName,
		strings.Join(valuesQuery.Columns, ","),
		strings.Join(valuesQuery.Vals, ","))
	args := valuesQuery.Args
	if urlQuery.IsDebug() {
		return s.debug(query, args...)
	}

	rows, dbErr := database.ExecQuery(r.Context(), s.db, query, args...)
	if dbErr != nil {
		return &Response{
			Code: dbErr.Code,
			Msg:  dbErr.Msg,
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

func (s *Server) delete(r *http.Request, tableName string, urlQuery database.URLQuery) *Response {
	var queryBuilder strings.Builder
	queryBuilder.WriteString("DELETE FROM ")
	queryBuilder.WriteString(tableName)
	_, whereQuery, args := urlQuery.WhereQuery(1)
	if whereQuery != "" {
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(whereQuery)
	}

	query := queryBuilder.String()
	if urlQuery.IsDebug() {
		return s.debug(query, args...)
	}

	rows, dbErr := database.ExecQuery(r.Context(), s.db, query, args...)
	if dbErr != nil {
		return &Response{
			Code: dbErr.Code,
			Msg:  dbErr.Msg,
		}
	}

	return &Response{
		Code: http.StatusOK,
		Msg:  fmt.Sprintf("successfully deleted %d rows", rows),
	}
}

func (s *Server) update(r *http.Request, tableName string, urlQuery database.URLQuery) *Response {
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

	var queryBuilder strings.Builder
	queryBuilder.WriteString(fmt.Sprintf("UPDATE %s SET %s", tableName, setQuery.Query))

	args := setQuery.Args
	_, whereQuery, args2 := urlQuery.WhereQuery(setQuery.Index)
	if whereQuery != "" {
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(whereQuery)
		args = append(args, args2...)
	}

	query := queryBuilder.String()
	if urlQuery.IsDebug() {
		return s.debug(query, args...)
	}

	rows, dbErr := database.ExecQuery(r.Context(), s.db, query, args...)
	if dbErr != nil {
		return &Response{
			Code: dbErr.Code,
			Msg:  dbErr.Msg,
		}
	}
	return &Response{
		Code: http.StatusOK,
		Msg:  fmt.Sprintf("successfully updated %d rows", rows),
	}
}

func (s *Server) get(r *http.Request, tableName string, urlQuery database.URLQuery) *Response {
	if urlQuery.IsCount() {
		return s.count(r, tableName)
	}

	var queryBuilder strings.Builder
	selects := urlQuery.SelectQuery()
	queryBuilder.WriteString(fmt.Sprintf("SELECT %s FROM %s", selects, tableName))
	_, whereQuery, args := urlQuery.WhereQuery(1)
	if whereQuery != "" {
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(whereQuery)
	}

	// order
	order := urlQuery.OrderQuery()
	if len(order) > 0 {
		queryBuilder.WriteString(" ORDER BY ")
		queryBuilder.WriteString(order)
	}

	// page operation
	page, pageSize := urlQuery.Page()
	queryBuilder.WriteString(" LIMIT ")
	queryBuilder.WriteString(fmt.Sprintf("%d", pageSize))
	if page != 1 {
		queryBuilder.WriteString(" OFFSET ")
		queryBuilder.WriteString(fmt.Sprintf("%d", (page-1)*pageSize))
	}

	query := queryBuilder.String()
	if urlQuery.IsDebug() {
		return s.debug(query, args...)
	}

	objects, dbErr := database.FetchData(r.Context(), s.db, query, args...)
	if dbErr != nil {
		return &Response{
			Code: dbErr.Code,
			Msg:  dbErr.Msg,
		}
	}
	var data any = objects
	if urlQuery.IsSingular() {
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
	query := fmt.Sprintf("SELECT COUNT(1) AS count FROM %s", tableName)
	rows, dbErr := database.FetchData(r.Context(), s.db, query)
	if dbErr != nil {
		return &Response{
			Code: dbErr.Code,
			Msg:  dbErr.Msg,
		}
	}

	return &Response{
		Code: http.StatusOK,
		Msg:  "success",
		Data: rows[0],
	}
}
