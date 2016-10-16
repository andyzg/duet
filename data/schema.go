package data

import (
	"github.com/graphql-go/graphql"
)

var taskType *graphql.Object

var Schema graphql.Schema

func init() {
	taskType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Task",
		Description: "A TODO task",
		Fields: graphql.Fields{
			"title": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return "Change me", nil
				},
			},
		},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"task": &graphql.Field{
				Type: taskType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return "Foo", nil
				},
			},
		},
	})

	var err error
	Schema, err = graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})
	if err != nil {
		panic(err)
	}
}
