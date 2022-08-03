package main

import (
	"log"
	"net/http"

	"github.com/rog-golang-buddies/rapidmidiex/www"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	return http.ListenAndServe(":8081", www.NewService())
}
