package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var databases = map[string]func() (*sql.DB, error){
	"PostgreSQL": setupPG,
	"MySQL":      setupMy,
	"SQLite":     setupSQLite,
}

func TestExecQuery(t *testing.T) {
	for name, setupFunc := range databases {
		testDBType := os.Getenv("REST_TEST_DB_TYPE")
		if testDBType != "" && name != testDBType {
			continue
		}

		t.Run(name, func(t *testing.T) {
			db, err := setupFunc()
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			query := fmt.Sprintf("UPDATE customers SET Name=%s WHERE id=%s", placeholder(name, 1), placeholder(name, 2))
			args := []any{"another name", 1}

			rows, err := ExecQuery(ctx, db, query, args...)
			assert.Equal(t, int64(1), rows)
			assert.Nil(t, err)
		})
	}
}

func TestFetchData(t *testing.T) {
	for name, setupFunc := range databases {
		testDBType := os.Getenv("REST_TEST_DB_TYPE")
		if testDBType != "" && name != testDBType {
			continue
		}

		t.Run(name, func(t *testing.T) {
			db, err := setupFunc()
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			query := fmt.Sprintf("SELECT * FROM customers WHERE id=%s", placeholder(name, 1))
			args := []any{1}

			objects, err := FetchData(ctx, db, query, args...)
			assert.Equal(t, 1, len(objects))
			assert.Nil(t, err)
		})
	}
}

func setupPG() (*sql.DB, error) {
	const (
		setupSQL = `
	DROP TABLE IF EXISTS "customers";
	CREATE TABLE IF NOT EXISTS "customers"
	(
		Id SERIAL PRIMARY KEY,
		Name VARCHAR(40)  NOT NULL,
		Number NUMERIC(10,2) NOT NULL,
		I1 INTEGER NOT NULL,
		I2 SMALLINT NOT NULL,
		I3 BIGINT NOT NULL,
		I4 BIGSERIAL NOT NULL,
		B1 BOOLEAN NOT NULL,
		F1 REAL NOT NULL,
		F2 FLOAT NOT NULL,
		F3 DOUBLE PRECISION NOT NULL,
		F4 DECIMAl NOT NULL,
		Data JSON NOT NULL
	);
	INSERT INTO customers VALUES (1, 'name', 10.2, 1, 2, 3, 4, true, 1.0, 2.0, 3.0, 4.0, '{"a":1, "b":"hello"}');
	`
	)
	url := os.Getenv("REST_PG_URL")
	if url == "" {
		url = "postgres://postgres:postgres@127.0.0.1:5432/postgres"
		fmt.Println("use default url: ", url)
	} else if !strings.HasPrefix(url, "postgres://") {
		return nil, fmt.Errorf("invalid postgres url: %s", url)
	}
	db, err := Open(url)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, setupSQL)
	return db, err
}

func setupMy() (*sql.DB, error) {
	var setupSQLs = []string{
		`DROP TABLE IF EXISTS customers`,
		`CREATE TABLE IF NOT EXISTS customers
		(
			Id INT AUTO_INCREMENT PRIMARY KEY NOT NULL,
			Name VARCHAR(40) NOT NULL,
			Number DEC(10,2) NOT NULL,
			I1 TINYINT NOT NULL,
			I2 SMALLINT NOT NULL,
			I3 BIGINT NOT NULL,
			I4 BIT(64) NOT NULL,
			B1 BOOLEAN NOT NULL,
			B2 BOOL NOT NULL,
			F1 REAL NOT NULL,
			F2 FLOAT NOT NULL,
			F3 DOUBLE NOT NULL,
			Data JSON NOT NULL
		)`,
		`INSERT INTO customers VALUES (1, "name", 10.2, 1, 2, 3, 4, true, false, 1.0, 2.0, 3.0, '{"a":1, "b":"hello"}')`,
	}

	url := os.Getenv("REST_MYSQL_URL")
	if url == "" {
		url = "mysql://root:root@tcp(127.0.0.1:3306)/mysql"
		fmt.Println("use default url: ", url)
	} else if !strings.HasPrefix(url, "mysql://") {
		return nil, fmt.Errorf("invalid mysql url: %s", url)
	}
	db, err := Open(url)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, setupSQL := range setupSQLs {
		if _, err = db.ExecContext(ctx, setupSQL); err != nil {
			return nil, err
		}
	}
	return db, nil
}

func setupSQLite() (*sql.DB, error) {
	const (
		setupSQL = `
	DROP TABLE IF EXISTS "customers";
	CREATE TABLE IF NOT EXISTS "customers"
	(
		[Id] INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
		[Name] NVARCHAR(40)  NOT NULL,
		[Number] NUMERIC(10,2) NOT NULL,
		[I1] TINYINT NOT NULL,
		[I2] SMALLINT NOT NULL,
		[I3] BIGINT NOT NULL,
		[I4] INT NOT NULL,
		[B1] BOOLEAN NOT NULL,
		[B2] BOOL NOT NULL,
		[F1] REAL NOT NULL,
		[F2] FLOAT NOT NULL,
		[F3] DOUBLE NOT NULL,
		[Data] JSON NOT NULL
	);
	INSERT INTO customers VALUES (1, "name", 10.2, 1, 2, 3, 4, true, false, 1.0, 2.0, 3.0, '{"a":1, "b":"hello"}');
	`
	)
	url := os.Getenv("REST_SQLITE_URL")
	if url == "" {
		url = "sqlite://ci.db"
		fmt.Println("use default url: ", url)
	} else if !strings.HasPrefix(url, "sqlite://") {
		return nil, fmt.Errorf("invalid sqlite url: %s", url)
	}
	db, err := Open(url)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, setupSQL)
	return db, err
}
