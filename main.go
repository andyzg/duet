package main

import (
	"log"
	"net/http"
	"time"

	"github.com/andyzg/duet/data"
	"github.com/andyzg/duet/graphiql"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/gabrielwong/graphql-go-handler"

	"golang.org/x/net/context"
)

func main() {
	db := data.InitDatabase()
	defer db.Close()

	graphqlHandler := handler.New(&handler.Config{
		Schema: data.GetSchema(db),
		Pretty: true,
		Log:    false,
	})

	authGraphqlHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := data.GetBearerToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		userId, err := data.AuthUserId(token)
		if err != nil {
			log.Printf("Error verifying token: %s", err.Error())
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		ctx = context.WithValue(ctx, data.UserIdKey, userId)

		graphqlHandler.ContextHandler(ctx, w, r)
	})

	restApi := rest.NewApi()
	restApi.Use(rest.DefaultDevStack...)

	restRouter, err := rest.MakeRouter(
		rest.Post("/login", data.ServeLogin(db)),
		rest.Post("/signup", data.ServeCreateUser(db)),
		rest.Get("/verify", data.ServeVerifyToken(db)),
	)
	if err != nil {
		log.Fatalf("rest.MakeRouter failed, %v", err)
	}
	restApi.SetApp(restRouter)

	http.HandleFunc("/", graphiql.ServeGraphiQL)
	http.Handle("/rest/", http.StripPrefix("/rest", restApi.MakeHandler()))
	http.Handle("/graphql", authGraphqlHandler)
	http.Handle("/oauth/todoist/login", data.HandleTodoistLogin(db))
	http.Handle("/oauth/todoist/callback", data.HandleTodoistCallback(db))

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("ListenAndServe failed, %v", err)
	}
}
