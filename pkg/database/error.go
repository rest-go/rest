package database

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgconn"
	"modernc.org/sqlite"
)

// Error converts database error to http code
type Error struct {
	Code int
	Msg  string
}

func (e Error) Error() string {
	return e.Msg
}

func NewError(hint string, err error) *Error {
	code := http.StatusInternalServerError
	msg := err.Error()

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		code = pgErrCodeToHTTPCode(pgErr.Code)
	}

	var myError *mysql.MySQLError
	if errors.As(err, &myError) {
		code = myErrCodeToHTTPCode(int(myError.Number))
	}

	var sqliteError *sqlite.Error
	if errors.As(err, &sqliteError) {
		code = sqliteErrCodeToHTTPCode(sqliteError.Code())
	}

	return &Error{
		Code: code,
		Msg:  fmt.Sprintf("%s, %s", hint, msg),
	}
}

// pgErrCodeToHTTPCode converts PG Error to HTTP code
// reference:
//   https://www.postgresql.org/docs/current/errcodes-appendix.html
func pgErrCodeToHTTPCode(code string) int {
	if strings.HasPrefix(code, "23") {
		// Integrity Constraint Violation
		return http.StatusBadRequest
	}
	if strings.HasPrefix(code, "42") {
		// Syntax Error or Access Rule Violation
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

// sqliteErrCodeToHTTPCode converts PG Error to HTTP code
// reference:
//   https://www.sqlite.org/rescode.html
func sqliteErrCodeToHTTPCode(code int) int {
	switch code {
	case 1299:
		return http.StatusBadRequest
	case 1555:
		return http.StatusConflict
	}

	return http.StatusInternalServerError
}

// myErrCodeToHTTPCode converts PG Error to HTTP code
// reference:
//   https://dev.mysql.com/doc/mysql-errors/8.0/en/server-error-reference.html
func myErrCodeToHTTPCode(code int) int {
	return http.StatusInternalServerError
}
