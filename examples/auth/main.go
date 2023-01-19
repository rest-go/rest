package main

import (
	"log"
	"net/http"

	"github.com/rest-go/auth"
	"github.com/rest-go/rest/pkg/server"
)

func main() {
	dbURL := "sqlite://my.db"
	jwtSecret := "my-secret"
	s := server.New(&server.DBConfig{URL: dbURL}, server.EnableAuth(true))
	restAuth, err := auth.New(dbURL, []byte(jwtSecret))
	if err != nil {
		log.Fatal("initialize auth error ", err)
	}
	http.Handle("/auth/", restAuth)
	http.Handle("/", restAuth.Middleware(s))
	log.Fatal(http.ListenAndServe(":3001", s)) //nolint:gosec
}
