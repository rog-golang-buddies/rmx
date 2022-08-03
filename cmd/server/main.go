package main

import (
	"net/http"

	"github.com/rog-golang-buddies/rapidmidiex/www"
)

func main() {
	// Feel free to delete this file.
	http.ListenAndServe(":8081", www.NewService())
}
