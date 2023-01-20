package jsonutil

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rest-go/rest/pkg/sql"
	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	t.Run("json response", func(t *testing.T) {
		rr := httptest.NewRecorder()
		res := &Response{Code: http.StatusBadGateway, Msg: "bad gateway"}
		Write(rr, res)
		assert.Equal(t, rr.Code, http.StatusBadGateway)
		assert.Equal(t, rr.Body.String(), "{\"msg\":\"bad gateway\"}\n")
	})

	t.Run("map data", func(t *testing.T) {
		rr := httptest.NewRecorder()
		res := map[string]string{"hello": "world"}
		Write(rr, res)
		assert.Equal(t, rr.Code, http.StatusOK)
		assert.Equal(t, rr.Body.String(), "{\"hello\":\"world\"}\n")
	})

}

func TestNewErrResponse(t *testing.T) {
	t.Run("sql error", func(t *testing.T) {
		err := sql.Error{Code: 1, Msg: "hello"}
		res := ErrResponse(err)
		assert.Equal(t, res.Code, err.Code)
		assert.Equal(t, res.Msg, err.Msg)
	})

	t.Run("non sql error", func(t *testing.T) {
		err := errors.New("error")
		res := ErrResponse(err)
		assert.Equal(t, res.Code, http.StatusInternalServerError)
		assert.Equal(t, res.Msg, err.Error())
	})
}
