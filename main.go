package main

import (
	"log"
	"net/http"
)

func main() {
	s := NewService()
	log.Fatal(http.ListenAndServe(":8080", s))
}
