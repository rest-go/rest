package main

import (
	"log"
	"net/http"

	"github.com/rest-go/rest/pkg/handler"
)

func main() {
	h := handler.New(&handler.DBConfig{URL: "sqlite://ci.db"})
	http.Handle("/admin/", h.WithPrefix("/admin"))
	log.Fatal(http.ListenAndServe(":3001", nil)) //nolint:gosec
}
