package sql

import (
	"context"
	"time"
)

const (
	testDBURL = "sqlite://ci.db"
	setupSQL  = `
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
		INSERT INTO customers VALUES
		(1, "name", 10.2, 1, 2, 3, 4, true, false, 1.0, 2.0, 3.0, '{"a":1, "b":"hello", "c":[1,2,3], "d":{"foo":"bar"}}'),
		(2, "name2", 10.2, 1, 2, 3, 4, true, false, 1.0, 2.0, 3.0, '{"a":1, "b":"hello", "c":[1,2,3], "d":{"foo":"bar"}}');
	`
)

func setupDB() (*DB, error) {
	db, err := Open(testDBURL)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err = db.ExecContext(ctx, setupSQL); err != nil {
		return nil, err
	}
	return db, nil
}
