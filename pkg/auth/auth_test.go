// package auth provide restful interface for authentication
package auth

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/rest-go/rest/pkg/log"
	"github.com/rest-go/rest/pkg/sql"
)

func TestJWTToken(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		data := map[string]any{
			"a": "b",
		}
		token, err := GenJWTToken([]byte(testSecret), data)
		assert.Nil(t, err)

		parsedData, err := ParseJWTToken([]byte(testSecret), token)
		assert.Nil(t, err)
		assert.True(t, reflect.DeepEqual(data, parsedData))
	})

	t.Run("invalid token", func(t *testing.T) {
		data := map[string]any{
			"a": "b",
		}
		token, err := GenJWTToken([]byte(testSecret), data)
		assert.Nil(t, err)

		parsedData, err := ParseJWTToken([]byte(testSecret), token[:len(token)-1])
		assert.Nil(t, parsedData)
		assert.NotNil(t, err)
		t.Log(err)
	})

	t.Run("expired token", func(t *testing.T) {
		data := map[string]any{
			"a":   "b",
			"exp": time.Now().Add(-24 * time.Hour).Unix(),
		}
		token, err := GenJWTToken([]byte(testSecret), data)
		assert.Nil(t, err)

		parsedData, err := ParseJWTToken([]byte(testSecret), token)
		assert.Nil(t, parsedData)
		assert.NotNil(t, err)
		t.Log(err)
	})
}

func TestSetup(t *testing.T) {
	file, err := os.CreateTemp(".", "test-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())
	db, err := sql.Open("sqlite://" + file.Name())
	assert.Nil(t, err)
	_, _, err = Setup(db)
	assert.Nil(t, err)

	// call Setup again will return an error
	_, _, err = Setup(db)
	assert.NotNil(t, err)
	t.Log(err)
}
