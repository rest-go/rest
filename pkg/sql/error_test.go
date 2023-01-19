package sql

import (
	"net/http"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"modernc.org/sqlite"
)

func TestNewError(t *testing.T) {
	dbErr := NewError(http.StatusInternalServerError, "msg")
	assert.Equal(t, http.StatusInternalServerError, dbErr.Code)
}

func TestErrorPG(t *testing.T) {
	err := &pgconn.PgError{}
	dbErr := convertError("hint", err)
	assert.Equal(t, http.StatusInternalServerError, dbErr.Code)
}

func TestErrorMy(t *testing.T) {
	err := &mysql.MySQLError{}
	dbErr := convertError("hint", err)
	assert.Equal(t, http.StatusInternalServerError, dbErr.Code)
}

func TestErrorSQLite(t *testing.T) {
	err := &sqlite.Error{}
	dbErr := convertError("hint", err)
	assert.Equal(t, http.StatusInternalServerError, dbErr.Code)
}
