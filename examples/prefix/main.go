package main

import (
	"log"
	"net/http"

	"github.com/rest-go/rest/pkg/server"
)

func main() {
	h := server.New(&server.DBConfig{URL: "sqlite://my.db"}, server.Prefix("/admin"))
	http.Handle("/admin/", h)
	log.Fatal(http.ListenAndServe(":3001", nil)) //nolint:gosec
}
