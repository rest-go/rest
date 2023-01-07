package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testDB struct {
	name       string
	urlEnv     string
	defaultURL string
	setupSQLs  []string
}

var databases []testDB = []testDB{
	{
		name:       "postgres",
		urlEnv:     "REST_PG_URL",
		defaultURL: "postgres://postgres:postgres@127.0.0.1:5432/postgres",
		setupSQLs: []string{`
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
	`},
	},
	{
		name:       "mysql",
		urlEnv:     "REST_MYSQL_URL",
		defaultURL: "mysql://root:root@tcp(127.0.0.1:3306)/mysql",
		setupSQLs: []string{
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
		},
	},
	{
		name:       "sqlite",
		urlEnv:     "REST_SQLITE_URL",
		defaultURL: "sqlite://ci.db",
		setupSQLs: []string{`
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
	`},
	},
}

func setupDB(testDB testDB) (*sql.DB, error) {
	url := os.Getenv(testDB.urlEnv)
	if url == "" {
		url = testDB.defaultURL
		fmt.Println("use default url: ", url)
	}

	db, err := Open(url)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, setupSQL := range testDB.setupSQLs {
		if _, err = db.ExecContext(ctx, setupSQL); err != nil {
			return nil, err
		}
	}
	return db, nil
}

func TestExecQuery(t *testing.T) {
	for _, testDB := range databases {
		testDBType := os.Getenv("REST_TEST_DB_TYPE")
		if testDBType != "" && testDB.name != testDBType {
			continue
		}

		t.Run(testDB.name, func(t *testing.T) {
			db, err := setupDB(testDB)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			query := fmt.Sprintf("UPDATE customers SET Name=%s WHERE id=%s", placeholder(testDB.name, 1), placeholder(testDB.name, 2))
			args := []any{"another name", 1}

			rows, err := ExecQuery(ctx, db, query, args...)
			assert.Equal(t, int64(1), rows)
			assert.Nil(t, err)
		})
	}
}

func TestFetchData(t *testing.T) {
	for _, testDB := range databases {
		testDBType := os.Getenv("REST_TEST_DB_TYPE")
		if testDBType != "" && testDB.name != testDBType {
			continue
		}

		t.Run(testDB.name, func(t *testing.T) {
			db, err := setupDB(testDB)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			query := fmt.Sprintf("SELECT * FROM customers WHERE id=%s", placeholder(testDB.name, 1))
			args := []any{1}

			objects, err := FetchData(ctx, db, query, args...)
			assert.Equal(t, 1, len(objects))
			assert.Nil(t, err)
		})
	}
}
