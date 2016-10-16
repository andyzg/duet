package data

import (
	"errors"
	"strconv"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

var taskType *graphql.Object
var dateType *graphql.Scalar

var Schema graphql.Schema

func init() {
	dateType = graphql.NewScalar(graphql.ScalarConfig{
		Name:        "Date",
		Description: "Date and time",
		Serialize: func(t interface{}) interface{} {
			switch t := t.(type) {
			case time.Time:
				return t.Unix()
			}
			return 0
		},
		ParseValue: func(unix interface{}) interface{} {
			switch unix := unix.(type) {
			case int64:
				return time.Unix(unix, 0)
			}
			return nil
		},
		ParseLiteral: func(valueAST ast.Value) interface{} {
			switch valueAST := valueAST.(type) {
			case *ast.IntValue:
				if intValue, err := strconv.ParseInt(valueAST.Value, 10, 32); err == nil {
					return time.Unix(intValue, 0)
				}
			}
			return nil
		},
	})

	taskType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Task",
		Description: "A TODO task",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.ID,
			},
			"title": &graphql.Field{
				Type: graphql.String,
			},
			"start_date": &graphql.Field{
				Type: dateType,
			},
			"end_date": &graphql.Field{
				Type: dateType,
			},
			"done": &graphql.Field{
				Type: graphql.Boolean,
			},
		},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"task": &graphql.Field{
				Type: taskType,
				Args: map[string]*graphql.ArgumentConfig{
					"id": &graphql.ArgumentConfig{
						Type:         graphql.ID,
						DefaultValue: nil,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					idArg := p.Args["id"].(string)
					if id, err := strconv.ParseInt(idArg, 10, 64); err == nil {
						if id >= 0 && id < 2 {
							return GetTask(id), nil
						} else {
							return nil, errors.New("Out of bounds")
						}
					} else {
						return nil, err
					}
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
