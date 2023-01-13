package sqlx

import (
	"net/http"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"modernc.org/sqlite"
)

func TestErrorPG(t *testing.T) {
	err := &pgconn.PgError{}
	dbErr := NewError("hint", err)
	assert.Equal(t, http.StatusInternalServerError, dbErr.Code)
}

func TestErrorMy(t *testing.T) {
	err := &mysql.MySQLError{}
	dbErr := NewError("hint", err)
	assert.Equal(t, http.StatusInternalServerError, dbErr.Code)
}

func TestErrorSQLite(t *testing.T) {
	err := &sqlite.Error{}
	dbErr := NewError("hint", err)
	assert.Equal(t, http.StatusInternalServerError, dbErr.Code)
}
