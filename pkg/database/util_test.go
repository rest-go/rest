package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidTableName(t *testing.T) {
	assert.True(t, IsValidTableName("table"))
	assert.False(t, IsValidTableName("0table"))
}

// Open connects to database by specify database url and ping it
func TestOpen(t *testing.T) {
	_, err := Open("sqlite://ci.db")
	assert.Nil(t, err)
}
