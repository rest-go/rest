package sql

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgconn"
	"modernc.org/sqlite"
)

const (
	// PG errors
	// https://www.postgresql.org/docs/current/errcodes-appendix.html
	PGIntegrityConstraintViolation = "23"
	PGSyntaxError                  = "42"

	// MySQL errors
	// https://dev.mysql.com/doc/mysql-errors/8.0/en/server-error-reference.html
	MYErrNoDefaultForField = 1364

	// SQLite errors
	// https://www.sqlite.org/rescode.html
	SQLiteConstraintNotNULL    = 1299
	SQLiteConstraintPrimaryKey = 1555
	SQLiteConstraintUnique     = 2067
)

// Error converts database error to http code
type Error struct {
	Code int
	Msg  string
}

func (e Error) Error() string {
	return e.Msg
}

func NewError(code int, msg string) Error {
	return Error{code, msg}
}

func convertError(hint string, err error) Error {
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

	return Error{
		Code: code,
		Msg:  fmt.Sprintf("%s, %s", hint, msg),
	}
}

// pgErrCodeToHTTPCode converts PG Error to HTTP code
func pgErrCodeToHTTPCode(code string) int {
	if strings.HasPrefix(code, PGIntegrityConstraintViolation) {
		return http.StatusBadRequest
	}
	if strings.HasPrefix(code, PGSyntaxError) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

// sqliteErrCodeToHTTPCode converts SQLite Error to HTTP code
func sqliteErrCodeToHTTPCode(code int) int {
	switch code {
	case SQLiteConstraintNotNULL:
		return http.StatusBadRequest
	case SQLiteConstraintPrimaryKey, SQLiteConstraintUnique:
		return http.StatusConflict
	}

	return http.StatusInternalServerError
}

// myErrCodeToHTTPCode converts MySQL Error to HTTP code
func myErrCodeToHTTPCode(code int) int {
	switch code { //nolint:gocritic
	case MYErrNoDefaultForField:
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}
