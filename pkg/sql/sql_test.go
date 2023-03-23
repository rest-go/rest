package sql

import (
	"context"
	"os"
	"testing"
	"time"

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
		_ = db.FetchTables()
	})

	t.Run("mysql", func(t *testing.T) {
		url := os.Getenv("REST_MY_URL")
		if url == "" {
			url = "mysql://root:root@tcp(localhost:3306)/test"
		}
		db, err := Open(url)
		assert.Nil(t, err)
		if err == nil {
			tables := db.FetchTables()
			t.Log("get tables: ", tables)
		}
	})

	t.Run("sqlite", func(t *testing.T) {
		db, err := setupDB()
		assert.Nil(t, err)
		tables := db.FetchTables()
		assert.Equal(t, 1, len(tables))
		t.Log("get table: ", tables)
		columns := tables["customers"].Columns
		assert.Equal(t, 13, len(columns))
		assert.Equal(t, "Id", columns[0].ColumnName)
		assert.Equal(t, "INTEGER", columns[0].DataType)
	})
}

func TestDBExec(t *testing.T) {
	db, err := setupDB()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	query := "UPDATE customers SET Name=$1 WHERE id=$2"
	args := []any{"another name", 1}

	rows, err := db.ExecQuery(ctx, query, args...)
	assert.Equal(t, int64(1), rows)
	assert.Nil(t, err)
}

func TestDBFetchData(t *testing.T) {
	db, err := setupDB()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := "SELECT * FROM customers WHERE id=$1"
	args := []any{1}
	objects, err := db.FetchData(ctx, query, args...)
	assert.Equal(t, 1, len(objects))
	assert.Nil(t, err)
}

func TestDBFetchOne(t *testing.T) {
	db, err := setupDB()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := "SELECT * FROM customers WHERE id=$1"
	args := []any{1}
	_, err = db.FetchOne(ctx, query, args...)
	assert.Nil(t, err)

	args = []any{3}
	_, err = db.FetchOne(ctx, query, args...)
	assert.Contains(t, err.Error(), "not found")

	query = "SELECT * FROM customers"
	_, err = db.FetchOne(ctx, query)
	assert.Contains(t, err.Error(), "multiple rows found")
}
