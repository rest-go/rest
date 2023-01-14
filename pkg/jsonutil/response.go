// package json handle all http json operations like decode input or output
// json response
package jsonutil

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/rest-go/rest/pkg/sqlx"
)

// Response serves a default JSON output when no data fetched from data
type Response struct {
	Code int    `json:"-"` // write to http status code
	Msg  string `json:"msg"`
}

func Encode(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	if res, ok := data.(*Response); ok {
		w.WriteHeader(res.Code)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode json data, %v", err)
	}
}

func SQLErrResponse(err *sqlx.Error) *Response {
	return &Response{Code: err.Code, Msg: err.Msg}
}
