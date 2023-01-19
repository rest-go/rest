package sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBind(t *testing.T) {
	origin := "SElECT * FROM A WHERE a = ? AND b = ?"
	a := Rebind("mysql", origin)
	assert.Equal(t, origin, a)

	psql := "SElECT * FROM A WHERE a = $1 AND b = $2"
	b := Rebind("postgres", origin)
	assert.Equal(t, psql, b)
}
