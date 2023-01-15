package jsonutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rest-go/rest/pkg/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	rr := httptest.NewRecorder()
	res := &Response{Code: http.StatusBadGateway, Msg: "bad gateway"}
	Write(rr, res)
	assert.Equal(t, rr.Code, http.StatusBadGateway)
	assert.Equal(t, rr.Body.String(), "{\"msg\":\"bad gateway\"}\n")
}
func TestNewErrResponse(t *testing.T) {
	err := sqlx.Error{Code: 1, Msg: "hello"}
	res := SQLErrResponse(err)
	assert.Equal(t, res.Code, err.Code)
	assert.Equal(t, res.Msg, err.Msg)
}
