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

	taskQuery := &graphql.Field{
		Type: taskType,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type:         graphql.ID,
				DefaultValue: nil,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id := p.Args["id"].(string)
			// Type of the nil matters apparently
			if task := GetTask(id); task != nil {
				return task, nil
			}
			return nil, nil
		},
	}

	tasksQuery := &graphql.Field{
		Type: graphql.NewList(taskType),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return GetTasks(), nil
		},
	}

	addTaskMutation := &graphql.Field{
		Type: taskType,
		Args: graphql.FieldConfigArgument{
			"title": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"start_date": &graphql.ArgumentConfig{
				Type: dateType,
			},
			"end_date": &graphql.ArgumentConfig{
				Type: dateType,
			},
			"done": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			title, _ := p.Args["title"].(string)
			startDate, _ := p.Args["start_date"].(time.Time)
			endDate, _ := p.Args["end_date"].(time.Time)
			done, _ := p.Args["done"].(bool)

			newTask := &Task{
				Title:     title,
				StartDate: startDate,
				EndDate:   endDate,
				Done:      done,
			}

			AddTask(newTask)
			return newTask, nil
		},
	}

	deleteTaskMutation := &graphql.Field{
		Type: graphql.NewObject(graphql.ObjectConfig{
			Name: "removeTaskPayload",
			Fields: graphql.Fields{
				"deletedId": &graphql.Field{
					Type: graphql.ID,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return p.Source, nil
					},
				},
			},
		}),
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id, _ := p.Args["id"].(string)
			DeleteTask(id)
			return id, nil
		},
	}

	updateTaskMutation := &graphql.Field{
		Type: taskType,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.ID,
			},
			"title": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"start_date": &graphql.ArgumentConfig{
				Type: dateType,
			},
			"end_date": &graphql.ArgumentConfig{
				Type: dateType,
			},
			"done": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id, _ := p.Args["id"].(string)
			task := GetTask(id)
			if task == nil {
				return nil, nil
			}

			if title, ok := p.Args["title"].(string); ok {
				task.Title = title
			}
			if startDate, ok := p.Args["start_date"].(time.Time); ok {
				task.StartDate = startDate
			}
			if endDate, ok := p.Args["end_date"].(time.Time); ok {
				task.EndDate = endDate
			}
			if done, ok := p.Args["done"].(bool); ok {
				task.Done = done
			}

			return task, nil
		},
	}

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"task":  taskQuery,
			"tasks": tasksQuery,
		},
	})

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootMutation",
		Fields: graphql.Fields{
			"addTask":    addTaskMutation,
			"deleteTask": deleteTaskMutation,
			"updateTask": updateTaskMutation,
		},
	})

	var err error
	Schema, err = graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
	if err != nil {
		panic(err)
	}
}
