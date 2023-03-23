package main

import (
	"log"
	"net/http"

	"github.com/rest-go/rest/pkg/auth"
)

func handle(w http.ResponseWriter, req *http.Request) {
	user := auth.GetUser(req)
	if user.IsAnonymous() {
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func main() {
	dbURL := "sqlite://my.db"
	jwtSecret := "my secret"
	authHandler, err := auth.NewHandler(dbURL, []byte(jwtSecret))
	if err != nil {
		log.Fatal(err)
	}
	middleware := auth.NewMiddleware([]byte(jwtSecret))

	http.Handle("/auth/", authHandler)
	http.Handle("/", middleware(http.HandlerFunc(handle)))
	log.Fatal(http.ListenAndServe(":8000", nil)) //nolint:gosec
}
