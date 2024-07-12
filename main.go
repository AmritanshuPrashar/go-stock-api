package main

import (
	"aiven-connect-to-pg/middleware"
	"aiven-connect-to-pg/router"
	"fmt"
	"log"
	"net/http"
)

func main() {
	middleware.InitDB()

	r := router.Router()

	addr := ":8080"

	fmt.Printf("Starting server on %s...\n", addr)

	log.Fatal(http.ListenAndServe(addr, r))
}
