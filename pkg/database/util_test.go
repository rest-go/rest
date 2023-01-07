package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidTableName(t *testing.T) {
	assert.True(t, IsValidTableName("table"))
	assert.False(t, IsValidTableName("0table"))
}
