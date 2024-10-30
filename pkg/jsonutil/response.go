// package json handle all http json operations like decode input or output
// json response
package jsonutil

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rest-go/rest/pkg/log"
	"github.com/rest-go/rest/pkg/sql"
)

// Response serves a default JSON output when no data fetched from data
type Response struct {
	Code int    `json:"-"` // write to http status code
	Msg  string `json:"msg"`
}

func Write(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")

	if res, ok := data.(*Response); ok {
		w.WriteHeader(res.Code)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Errorf("failed to encode json data, %v", err)
	}
}

func ErrResponse(err error) *Response {
	var dbErr sql.Error
	if errors.As(err, &dbErr) {
		return &Response{Code: dbErr.Code, Msg: dbErr.Msg}
	}
	return &Response{Code: http.StatusInternalServerError, Msg: err.Error()}
}
