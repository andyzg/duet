package main

import (
	"log"
	"net/http"

	"github.com/andyzg/duet/data"
	"github.com/graphql-go/handler"
)

func main() {
	h := handler.New(&handler.Config{
		Schema: &data.Schema,
		Pretty: true,
	})
	http.Handle("/", h)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("ListenAndServe failed, %v", err)
	}
}
