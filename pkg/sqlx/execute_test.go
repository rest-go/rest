package sqlx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecQuery(t *testing.T) {
	db, err := setupDB()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	query := "UPDATE customers SET Name=$1 WHERE id=$2"
	args := []any{"another name", 1}

	rows, err := ExecQuery(ctx, db, query, args...)
	assert.Equal(t, int64(1), rows)
	assert.Nil(t, err)
}

func TestFetchData(t *testing.T) {
	db, err := setupDB()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := "SELECT * FROM customers WHERE id=$1"
	args := []any{1}
	objects, err := FetchData(ctx, db, query, args...)
	assert.Equal(t, 1, len(objects))
	assert.Nil(t, err)
}
