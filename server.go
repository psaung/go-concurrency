package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/psaung/go-concurrency/handlers"
)

func main() {
	fmt.Println("Welcome to the Orders App!")
	handler, err := handlers.New()
	if err != nil {
		log.Fatal(err)
	}

	router := handlers.ConfigureHandler(handler)
	fmt.Println("Listening on localhost:8000...")
	log.Fatal(http.ListenAndServe(":8000", router))
}
