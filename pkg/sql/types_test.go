package sql

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeJSON(t *testing.T) {
	_, converter := getTypeAndConverter("json")
	val := converter(&sql.NullString{String: "1", Valid: true})
	assert.Equal(t, float64(1), val.(float64))

	val = converter(&sql.NullString{String: "10.0", Valid: true})
	assert.Equal(t, float64(10.0), val.(float64))

	val = converter(&sql.NullString{String: "True", Valid: true})
	assert.Equal(t, true, val.(bool))

	val = converter(&sql.NullString{String: "normal string", Valid: true})
	assert.Equal(t, "normal string", val.(string))
}

func TestTypeInt(t *testing.T) {
	for _, typeName := range []string{
		"TINYINT", "SMALLINT", "INT", "INTEGER", "BIGINT",
		"SMALLSERIAL", "SERIAL", "BIGSERIAL"} {
		t.Log("test int type: ", typeName)
		obj, converter := getTypeAndConverter(typeName)
		err := obj.(sql.Scanner).Scan(100)
		assert.Nil(t, err)

		val := converter(obj)
		assert.Equal(t, int64(100), val.(int64))
	}
}

func TestTypeFloat(t *testing.T) {
	for _, typeName := range []string{
		"FLOAT2", "DEC(10,2)", "DOUBLE PRECISION", "REAL", "DECIMAL",
		"NUMERIC(10,2)", "FLOAT"} {
		t.Log("test float type: ", typeName)
		obj, converter := getTypeAndConverter(typeName)
		err := obj.(sql.Scanner).Scan(3.1415926)
		assert.Nil(t, err)

		val := converter(obj)
		assert.Equal(t, 3.1415926, val.(float64))
	}
}

func TestTypeBool(t *testing.T) {
	for _, typeName := range []string{"bool", "Boolean"} {
		t.Log("test bool type: ", typeName)
		obj, converter := getTypeAndConverter(typeName)
		err := obj.(sql.Scanner).Scan(true)
		assert.Nil(t, err)

		val := converter(obj)
		assert.Equal(t, true, val.(bool))
	}
}

func TestTypeString(t *testing.T) {
	for _, typeName := range []string{
		"BINARY", "BLOB", "CHAR", "CLOB", "DATE", "DATETIME", "ENUM",
		"NVARCHAR(40)", "TEXT", "timestamp", "UUID", "VARCHAR(40)", "XML",
	} {
		t.Log("test string type: ", typeName)
		obj, converter := getTypeAndConverter(typeName)
		err := obj.(sql.Scanner).Scan("to be or not to be, that's a question")
		assert.Nil(t, err)

		val := converter(obj)
		assert.Equal(t, "to be or not to be, that's a question", val.(string))
	}
}

func TestTypeUnknown(t *testing.T) {
	obj, converter := getTypeAndConverter("")
	err := obj.(sql.Scanner).Scan("to be or not to be, that's a question")
	assert.Nil(t, err)

	val := converter(obj)
	assert.Equal(t, "to be or not to be, that's a question", val.(string))
}
