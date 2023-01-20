package jsonutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testS struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestMapToStruct(t *testing.T) {
	data := map[string]any{"name": "hello", "age": 10}
	var s testS
	err := MapToStruct(data, &s)
	assert.Nil(t, err)
	assert.Equal(t, "hello", s.Name)
	assert.Equal(t, 10, s.Age)
}
