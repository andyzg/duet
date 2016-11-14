package main

import (
	"log"
	"net/http"

	"github.com/andyzg/duet/data"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/gabrielwong/graphql-go-handler"
	"github.com/mnmtanish/go-graphiql"

	"golang.org/x/net/context"
)

func main() {
	data.InitDatabase()
	defer data.CloseDatabase()

	graphqlHandler := handler.New(&handler.Config{
		Schema: &data.Schema,
		Pretty: true,
		Log:    true,
	})

	authGraphqlHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		graphqlHandler.ContextHandler(ctx, w, r)
	})

	restApi := rest.NewApi()
	restApi.Use(rest.DefaultDevStack...)

	restRouter, err := rest.MakeRouter(
		rest.Post("/login", data.ServeLogin),
		rest.Post("/signup", data.ServeCreateUser),
		rest.Get("/verify", data.ServeVerifyToken),
	)
	if err != nil {
		log.Fatalf("rest.MakeRouter failed, %v", err)
	}
	restApi.SetApp(restRouter)

	http.HandleFunc("/", graphiql.ServeGraphiQL)
	http.Handle("/rest/", http.StripPrefix("/rest", restApi.MakeHandler()))
	http.Handle("/graphql", authGraphqlHandler)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("ListenAndServe failed, %v", err)
	}
}
