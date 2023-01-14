package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	j "github.com/rest-go/rest/pkg/jsonutil"
	"github.com/rest-go/rest/pkg/sqlx"
)

// Server is the representation of a restful server which handles CRUD requests
type Server struct {
	db     *sqlx.DB
	prefix string
}

// NewServer returns a Server pointer
// TODO: receive a config
func NewServer(url string) *Server {
	log.Printf("connecting to database: %s", url)
	db, err := sqlx.Open(url)
	if err != nil {
		log.Fatal(err)
	}
	defaultIdleConns := 50
	defaultOpenConns := 50
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(defaultIdleConns)
	db.SetMaxOpenConns(defaultOpenConns)

	return &Server{db: db}
}

func (s *Server) WithPrefix(prefix string) *Server {
	s.prefix = prefix
	return s
}

func (s *Server) debug(query string, args ...any) any {
	return &struct {
		Query string `json:"query"`
		Args  []any  `json:"args"`
	}{
		query, args,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, s.prefix)
	table := strings.Trim(path, "/")
	if table == "" {
		res := &j.Response{
			Code: http.StatusOK,
			Msg:  "rest server is up and running",
		}
		j.Encode(w, res)
		return
	}

	id := ""
	parts := strings.Split(table, "/")
	if len(parts) == 2 {
		table, id = parts[0], parts[1]
	}
	if !sqlx.IsValidTableName(table) {
		res := &j.Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("invalid table name: %s", table),
		}
		j.Encode(w, res)
		return
	}

	var data any
	values := r.URL.Query()
	if id != "" {
		values.Set("id", fmt.Sprintf("eq.%s", id))
	}
	urlQuery := sqlx.NewURLQuery(values, s.db.DriverName)
	switch r.Method {
	case "POST":
		data = s.create(r, table, urlQuery)
	case "DELETE":
		data = s.delete(r, table, urlQuery)
	case "PUT", "PATCH":
		data = s.update(r, table, urlQuery)
	case "GET":
		data = s.get(r, table, urlQuery)
	default:
		data = &j.Response{
			Code: http.StatusMethodNotAllowed,
			Msg:  fmt.Sprintf("method not supported: %s", r.Method),
		}
	}
	j.Encode(w, data)
}

func (s *Server) create(r *http.Request, tableName string, urlQuery *sqlx.URLQuery) any {
	var data sqlx.PostData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return &j.Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("failed to parse post json data, %v", err),
		}
	}

	valuesQuery, err := data.ValuesQuery()
	if err != nil {
		return &j.Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("failed to prepare values query, %v", err),
		}
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		tableName,
		strings.Join(valuesQuery.Columns, ","),
		strings.Join(valuesQuery.Placeholders, ","))
	args := valuesQuery.Args
	if urlQuery.IsDebug() {
		return s.debug(query, args...)
	}

	rows, dbErr := s.db.ExecQuery(r.Context(), query, args...)
	if dbErr != nil {
		return j.SQLErrResponse(dbErr)
	}
	if rows != int64(len(valuesQuery.Placeholders)) {
		return &j.Response{
			Code: http.StatusInternalServerError,
			Msg:  fmt.Sprintf("expected to insert %d rows, but affected %d rows", len(valuesQuery.Placeholders), rows),
		}
	}

	return &j.Response{
		Code: http.StatusOK,
		Msg:  fmt.Sprintf("successfully inserted %d rows", rows),
	}
}

func (s *Server) delete(r *http.Request, tableName string, urlQuery *sqlx.URLQuery) any {
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

	rows, dbErr := s.db.ExecQuery(r.Context(), query, args...)
	if dbErr != nil {
		return j.SQLErrResponse(dbErr)
	}

	return &j.Response{
		Code: http.StatusOK,
		Msg:  fmt.Sprintf("successfully deleted %d rows", rows),
	}
}

func (s *Server) update(r *http.Request, tableName string, urlQuery *sqlx.URLQuery) any {
	var data sqlx.PostData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return &j.Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("failed to parse update json data, %v", err),
		}
	}
	setQuery, err := data.SetQuery(1)
	if err != nil {
		return &j.Response{
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

	rows, dbErr := s.db.ExecQuery(r.Context(), query, args...)
	if dbErr != nil {
		return j.SQLErrResponse(dbErr)
	}
	return &j.Response{
		Code: http.StatusOK,
		Msg:  fmt.Sprintf("successfully updated %d rows", rows),
	}
}

func (s *Server) get(r *http.Request, tableName string, urlQuery *sqlx.URLQuery) any {
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

	objects, dbErr := s.db.FetchData(r.Context(), query, args...)
	if dbErr != nil {
		return j.SQLErrResponse(dbErr)
	}

	if urlQuery.IsSingular() || urlQuery.HasID() {
		if len(objects) == 0 {
			return &j.Response{
				Code: http.StatusNotFound,
				Msg:  "query data not found in database",
			}
		} else if len(objects) > 1 {
			return &j.Response{
				Code: http.StatusBadRequest,
				Msg:  fmt.Sprintf("expect singular data, but got %d rows", len(objects)),
			}
		}
		return objects[0] // return single map[string]any
	}
	return objects // return  []map[string]any
}

func (s *Server) count(r *http.Request, tableName string) any {
	query := fmt.Sprintf("SELECT COUNT(1) AS count FROM %s", tableName)
	objects, dbErr := s.db.FetchData(r.Context(), query)
	if dbErr != nil {
		return j.SQLErrResponse(dbErr)
	}
	return objects[0]["count"]
}
