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

func TestDBTables(t *testing.T) {
	t.Run("postgres", func(t *testing.T) {
		url := os.Getenv("REST_PG_URL")
		if url == "" {
			url = "postgres://postgres:postgres@localhost:5432/postgres"
		}
		db, err := Open(url)
		assert.Nil(t, err)
		_ = db.Tables()
	})

	t.Run("mysql", func(t *testing.T) {
		url := os.Getenv("REST_MY_URL")
		if url == "" {
			url = "mysql://root:root@tcp(localhost:3306)/test"
		}
		db, err := Open(url)
		assert.Nil(t, err)
		tables := db.Tables()
		t.Log("get tables: ", tables)
	})

	t.Run("sqlite", func(t *testing.T) {
		db, err := setupDB()
		assert.Nil(t, err)
		tables := db.Tables()
		assert.Equal(t, 1, len(tables))
		t.Log("get table: ", tables[0])
		columns := tables[0].Columns
		assert.Equal(t, 13, len(columns))
		assert.Equal(t, "Id", columns[0].ColumnName)
		assert.Equal(t, "INTEGER", columns[0].DataType)
	})
}
