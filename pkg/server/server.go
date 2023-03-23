// package handler provide RESTFul interfaces for all database tables
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rest-go/rest/pkg/auth"
	j "github.com/rest-go/rest/pkg/jsonutil"
	"github.com/rest-go/rest/pkg/log"
	"github.com/rest-go/rest/pkg/sql"
)

type UserAuthInfo struct {
	column string
	val    int64
}

// Server is the representation of a restful server which handles CRUD requests
type Server struct {
	db          *sql.DB
	prefix      string
	authEnabled bool

	tablesMu   sync.RWMutex
	tables     map[string]*sql.Table
	policiesMu sync.RWMutex
	policies   map[string]map[string]string // {table:action:exp}

	done chan struct{}
}

// New returns a Handler pointer
func New(dbConfig *DBConfig, options ...Option) *Server {
	log.Infof("start server with config %v", dbConfig)
	db, err := sql.Open(dbConfig.URL)
	if err != nil {
		log.Fatal(err)
	}
	defaultIdleConns := 50
	defaultOpenConns := 50
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(defaultIdleConns)
	db.SetMaxOpenConns(defaultOpenConns)
	h := &Server{db: db, done: make(chan struct{})}
	for _, opt := range options {
		opt(h)
	}
	h.updateMeta()
	return h
}

func (s *Server) Close() {
	close(s.done)
}

func (s *Server) updateMeta() {
	updateTask := func() {
		tables := s.db.FetchTables()
		ts := make([]string, 0, len(tables))
		for _, t := range tables {
			ts = append(ts, t.String())
		}
		log.Tracef("fetch tables from db: \n%s\n", strings.Join(ts, "\n"))
		s.tablesMu.Lock()
		s.tables = tables
		s.tablesMu.Unlock()

		s.updatePolicies()
	}
	updateTask()
	go func() {
		interval := 30 * time.Second
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-s.done:
				return
			case <-ticker.C:
				updateTask()
			}
		}
	}()
}

func (s *Server) updatePolicies() {
	if !s.authEnabled {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), sql.DefaultTimeout)
	defer cancel()
	policiesData, err := s.db.FetchData(ctx, "SELECT table_name, action, expression FROM auth_policies")
	if err != nil {
		log.Errorf("fetch policies from db error: %v", err)
	}
	policies := map[string]map[string]string{}
	for _, policyData := range policiesData {
		var policy auth.Policy
		err := j.MapToStruct(policyData, &policy)
		if err != nil {
			log.Errorf("get policy error: %v", err)
			continue
		}
		if t, ok := policies[policy.TableName]; ok {
			t[policy.Action] = policy.Expression
		} else {
			t := map[string]string{}
			t[policy.Action] = policy.Expression
			policies[policy.TableName] = t
		}
	}
	log.Tracef("fetch policies from db: \n%s\n", policies)
	s.policiesMu.Lock()
	s.policies = policies
	s.policiesMu.Unlock()
}

func (s *Server) getTables() map[string]*sql.Table {
	s.tablesMu.RLock()
	defer s.tablesMu.RUnlock()
	return s.tables
}

func (s *Server) getPolicies() map[string]map[string]string {
	s.tablesMu.RLock()
	defer s.tablesMu.RUnlock()
	return s.policies
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Infof("%s %s", r.Method, r.URL.RequestURI())
	path := strings.TrimPrefix(r.URL.Path, s.prefix)
	tableName := strings.Trim(path, "/")
	if tableName == "" {
		res := &j.Response{
			Code: http.StatusOK,
			Msg:  "rest server is up and running",
		}
		j.Write(w, res)
		return
	}

	// check table name
	pk := ""
	parts := strings.Split(tableName, "/")
	if len(parts) == 2 {
		tableName, pk = parts[0], parts[1]
	}
	table, ok := s.getTables()[tableName]
	if !ok {
		res := &j.Response{
			Code: http.StatusNotFound,
			Msg:  fmt.Sprintf("table does not exist: %s", tableName),
		}
		j.Write(w, res)
		return
	}

	urlQuery := sql.NewURLQuery(r.URL.Query(), s.db.DriverName)
	// check primary key
	if pk != "" {
		if table.PrimaryKey == "" {
			res := &j.Response{
				Code: http.StatusBadRequest,
				Msg:  fmt.Sprintf("primary key not found on table: %s", table),
			}
			j.Write(w, res)
			return
		}
		urlQuery.Set(table.PrimaryKey, fmt.Sprintf("eq.%s", pk))
		urlQuery.Set("singular", "")
	}

	var authInfo *UserAuthInfo
	if s.authEnabled {
		action := getAction(urlQuery, r.Method)
		user := auth.GetUser(r)
		hasPerm, userIDColumn := user.HasPerm(tableName, action, s.getPolicies())
		if !hasPerm {
			var res *j.Response
			if user.IsAnonymous() {
				res = &j.Response{
					Code: http.StatusUnauthorized,
					Msg:  "login required",
				}
			} else {
				res = &j.Response{
					Code: http.StatusForbidden,
					Msg:  "unauthorized",
				}
			}
			j.Write(w, res)
			return
		}
		if userIDColumn != "" {
			authInfo = &UserAuthInfo{userIDColumn, user.ID}
		}
	}

	var data any
	switch r.Method {
	case "POST":
		data = s.create(r, tableName, urlQuery, authInfo)
	case "DELETE":
		data = s.delete(r, tableName, urlQuery, authInfo)
	case "PUT", "PATCH":
		data = s.update(r, tableName, urlQuery, authInfo)
	case "GET":
		data = s.get(r, tableName, urlQuery, authInfo)
	default:
		data = &j.Response{
			Code: http.StatusMethodNotAllowed,
			Msg:  fmt.Sprintf("method not supported: %s", r.Method),
		}
	}
	j.Write(w, data)
}

func (s *Server) create(r *http.Request, tableName string, urlQuery *sql.URLQuery, userInfo *UserAuthInfo) any {
	var data sql.PostData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Warnf("failed to parse post json data: %v", err)
		return &j.Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("failed to parse post json data, %v", err),
		}
	}
	if userInfo != nil {
		// create for current auth user
		data.Set(userInfo.column, userInfo.val)
	}
	valuesQuery, err := data.ValuesQuery()
	if err != nil {
		log.Warnf("failed to generate values query %v", err)
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
		log.Errorf("create error: %v", dbErr)
		return j.ErrResponse(dbErr)
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

func (s *Server) delete(r *http.Request, tableName string, urlQuery *sql.URLQuery, userInfo *UserAuthInfo) any {
	if userInfo != nil {
		// filter by current auth user
		urlQuery.Set(userInfo.column, fmt.Sprintf("eq.%d", userInfo.val))
	}

	var queryBuilder strings.Builder
	queryBuilder.WriteString("DELETE FROM ")
	queryBuilder.WriteString(tableName)
	_, whereQuery, args := urlQuery.WhereQuery(1)
	if whereQuery != "" {
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(whereQuery)
	} else {
		return &j.Response{
			Code: http.StatusBadRequest,
			Msg: `delete without any condition is not allowed, please check the url query
If you really want to do it, uses 1=eq.1 to bypass it`,
		}
	}

	query := queryBuilder.String()
	if urlQuery.IsDebug() {
		return s.debug(query, args...)
	}

	rows, dbErr := s.db.ExecQuery(r.Context(), query, args...)
	if dbErr != nil {
		log.Errorf("delete error: %v", dbErr)
		return j.ErrResponse(dbErr)
	}

	return &j.Response{
		Code: http.StatusOK,
		Msg:  fmt.Sprintf("successfully deleted %d rows", rows),
	}
}

func (s *Server) update(r *http.Request, tableName string, urlQuery *sql.URLQuery, userInfo *UserAuthInfo) any {
	if userInfo != nil {
		// filter current auth user
		urlQuery.Set(userInfo.column, fmt.Sprintf("eq.%d", userInfo.val))
	}

	var data sql.PostData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Warnf("failed to parse update json data: %v", err)
		return &j.Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("failed to parse update json data, %v", err),
		}
	}
	setQuery, err := data.SetQuery(1)
	if err != nil {
		log.Warnf("failed to generate set query: %v", err)
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
	} else {
		return &j.Response{
			Code: http.StatusBadRequest,
			Msg: `update without any condition is not allowed, please check the url query
If you really want to do it, uses 1=eq.1 to bypass it`,
		}
	}

	query := queryBuilder.String()
	if urlQuery.IsDebug() {
		return s.debug(query, args...)
	}

	rows, dbErr := s.db.ExecQuery(r.Context(), query, args...)
	if dbErr != nil {
		log.Errorf("update error: %v", dbErr)
		return j.ErrResponse(dbErr)
	}
	return &j.Response{
		Code: http.StatusOK,
		Msg:  fmt.Sprintf("successfully updated %d rows", rows),
	}
}

func (s *Server) get(r *http.Request, tableName string, urlQuery *sql.URLQuery, userInfo *UserAuthInfo) any {
	if userInfo != nil {
		// filter current auth user
		urlQuery.Set(userInfo.column, fmt.Sprintf("eq.%d", userInfo.val))
	}

	if urlQuery.IsCount() {
		return s.count(r, tableName, urlQuery)
	}

	var queryBuilder strings.Builder
	selects, err := urlQuery.SelectQuery()
	if err != nil {
		log.Errorf("invalid select query %v", urlQuery)
		return &j.Response{
			Code: http.StatusBadRequest,
			Msg:  err.Error(),
		}
	}
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
		log.Errorf("read error: %v", dbErr)
		return j.ErrResponse(dbErr)
	}

	if urlQuery.IsSingular() {
		if len(objects) == 0 {
			return &j.Response{
				Code: http.StatusNotFound,
				Msg:  "data not found in database",
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

func (s *Server) count(r *http.Request, tableName string, urlQuery *sql.URLQuery) any {
	query := fmt.Sprintf("SELECT COUNT(1) AS count FROM %s", tableName)
	_, whereQuery, args := urlQuery.WhereQuery(1)
	if whereQuery != "" {
		query += fmt.Sprintf(" WHERE %s", whereQuery)
	}

	objects, dbErr := s.db.FetchData(r.Context(), query, args...)
	if dbErr != nil {
		log.Errorf("fetch count error: %v", dbErr)
		return j.ErrResponse(dbErr)
	}
	return objects[0]["count"]
}

func (s *Server) debug(query string, args ...any) any {
	return &struct {
		Query string `json:"query"`
		Args  []any  `json:"args"`
	}{
		query, args,
	}
}

func getAction(query *sql.URLQuery, method string) auth.Action {
	switch method {
	case "POST":
		return auth.ActionCreate
	case "PUT", "PATCH":
		return auth.ActionUpdate
	case "DELETE":
		return auth.ActionDelete
	default:
		if query.IsMine() {
			return auth.ActionReadMine
		}
		return auth.ActionRead
	}
}
