package sqlx

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpen(t *testing.T) {
	_, err := Open("invalid url")
	assert.NotNil(t, err)

	t.Run("postgres", func(t *testing.T) {
		url := os.Getenv("REST_PG_URL")
		if url == "" {
			url = "postgres://postgres:postgres@localhost:5432/postgres"
		}
		_, err := Open(url)
		assert.Nil(t, err)
	})

	t.Run("mysql", func(t *testing.T) {
		url := os.Getenv("REST_MY_URL")
		if url == "" {
			url = "mysql://root:root@tcp(localhost:3306)/mysql"
		}
		_, err := Open(url)
		assert.Nil(t, err)
	})

	t.Run("sqlite", func(t *testing.T) {
		url := os.Getenv("REST_SQLITE_URL")
		if url == "" {
			url = "sqlite://ci.db"
		}
		_, err := Open(url)
		assert.Nil(t, err)
	})
}
