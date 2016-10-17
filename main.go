package main

import (
	"log"
	"net/http"

	"github.com/andyzg/duet/data"
	"github.com/graphql-go/handler"
	"github.com/mnmtanish/go-graphiql"
)

func main() {
	data.InitDatabase()
	h := handler.New(&handler.Config{
		Schema: &data.Schema,
		Pretty: true,
	})
	http.Handle("/graphql", h)
	http.HandleFunc("/", graphiql.ServeGraphiQL)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("ListenAndServe failed, %v", err)
	}
}
