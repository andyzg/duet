package data

import (
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
		Name:        "date",
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
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return GetTask(0), nil
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
