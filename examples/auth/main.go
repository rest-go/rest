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
	authHandler, err := auth.NewHandler(dbURL, []byte(jwtSecret))
	if err != nil {
		log.Fatal("initialize auth error ", err)
	}
	http.Handle("/auth/", authHandler)

	middleware := auth.NewMiddleware([]byte(jwtSecret))
	http.Handle("/", middleware(s))
	log.Fatal(http.ListenAndServe(":3001", s)) //nolint:gosec
}
