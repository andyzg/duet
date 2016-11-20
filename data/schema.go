package data

import (
	"log"
	"strconv"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

var taskType *graphql.Object
var habitType *graphql.Object

var Schema graphql.Schema

var UserIdKey string = "user_id"

func userIdOfContext(p graphql.ResolveParams) uint64 {
	id := p.Context.Value(UserIdKey).(uint64)
	log.Printf("Got from context user id \"%s\"\n", id)
	return id
}

func init() {
	dateType := graphql.NewScalar(graphql.ScalarConfig{
		Name:        "Date",
		Description: "Date and time",
		Serialize: func(t interface{}) interface{} {
			switch t := t.(type) {
			case *time.Time:
				if t != nil {
					return t.Unix()
				}
			}
			return nil
		},
		ParseValue: func(unix interface{}) interface{} {
			// Variables are parsed by graphql-go-hander and passed as a map
			// It parses ints as a float64 but we keep int64 for completeness
			switch unix := unix.(type) {
			case int64:
				t := time.Unix(unix, 0)
				return &t
			case float64:
				t := time.Unix(int64(unix), 0)
				return &t
			}
			return nil
		},
		ParseLiteral: func(valueAST ast.Value) interface{} {
			// This is when the value is part of the query
			switch valueAST := valueAST.(type) {
			case *ast.IntValue:
				if intValue, err := strconv.ParseInt(valueAST.Value, 10, 32); err == nil {
					t := time.Unix(intValue, 0)
					return &t
				}
			}
			return nil
		},
	})

	actionKind := graphql.NewEnum(graphql.EnumConfig{
		Name:        "ActionKind",
		Description: "The kind of action performed on a task or habit",
		Values: graphql.EnumValueConfigMap{
			"PROGRESS": &graphql.EnumValueConfig{
				Value:       ActionProgress,
				Description: "Indication of progress on the task",
			},
			"DEFER": &graphql.EnumValueConfig{
				Value:       ActionDefer,
				Description: "User is defering the task",
			},
		},
	})

	interval := graphql.NewEnum(graphql.EnumConfig{
		Name:        "Interval",
		Description: "The recurring period for a Habit",
		Values: graphql.EnumValueConfigMap{
			"DAILY": &graphql.EnumValueConfig{
				Value:       Daily,
				Description: "The habit resets daily",
			},
			"WEEKLY": &graphql.EnumValueConfig{
				Value:       Weekly,
				Description: "The habit resets weekly",
			},
			"MONTHLY": &graphql.EnumValueConfig{
				Value:       Monthly,
				Description: "The habit resets monthly",
			},
		},
	})

	actionType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Action",
		Description: "An action that is performed on a task or habit",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.ID,
			},
			"kind": &graphql.Field{
				Type: actionKind,
			},
			"when": &graphql.Field{
				Type: dateType,
			},
		},
	})

	baseTaskInterface := graphql.NewInterface(graphql.InterfaceConfig{
		Name:        "BaseTask",
		Description: "Fields common to task and habit",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.ID,
			},
			"title": &graphql.Field{
				Type: graphql.String,
			},
			"actions": &graphql.Field{
				Type: graphql.NewList(actionType),
			},
		},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			task, ok := p.Value.(Task)
			if !ok {
				log.Printf("Failed to get task %v", task)
				return taskType
			}
			if task.Kind == TaskEnum {
				return taskType
			}
			return habitType
		},
	})
	_ = baseTaskInterface

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
			"actions": &graphql.Field{
				Type: graphql.NewList(actionType),
			},
		},
	})

	habitType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Habit",
		Description: "A recurring habit",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.ID,
			},
			"title": &graphql.Field{
				Type: graphql.String,
			},
			"interval": &graphql.Field{
				Type: interval,
			},
			"frequency": &graphql.Field{
				Type: graphql.Int,
			},
			"done": &graphql.Field{
				Type: graphql.Boolean,
			},
			"actions": &graphql.Field{
				Type: graphql.NewList(actionType),
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
			kind := TaskEnum
			task, err := GetTask(id, userIdOfContext(p), &kind)
			if err != nil {
				return nil, err
			}
			// Type of the nil matters apparently
			if task == nil {
				return nil, nil
			}
			return task, nil
		},
	}

	habitQuery := &graphql.Field{
		Type: habitType,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type:         graphql.ID,
				DefaultValue: nil,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id := p.Args["id"].(string)
			kind := HabitEnum
			task, err := GetTask(id, userIdOfContext(p), &kind)
			if err != nil {
				return nil, err
			}
			// Type of the nil matters apparently
			if task == nil {
				return nil, nil
			}
			return task, nil
		},
	}

	tasksQuery := &graphql.Field{
		Type: graphql.NewList(taskType),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			kind := TaskEnum
			return GetTasks(userIdOfContext(p), &kind)
		},
	}

	habitsQuery := &graphql.Field{
		Type: graphql.NewList(habitType),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			kind := HabitEnum
			return GetTasks(userIdOfContext(p), &kind)
		},
	}

	addTaskMutation := &graphql.Field{
		Type: taskType,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.ID,
			},
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
			id, _ := p.Args["id"].(string)
			title, _ := p.Args["title"].(string)
			startDate, _ := p.Args["start_date"].(*time.Time)
			endDate, _ := p.Args["end_date"].(*time.Time)
			done, _ := p.Args["done"].(bool)

			newTask := &Task{
				Id:        id,
				Title:     title,
				StartDate: startDate,
				EndDate:   endDate,
				Done:      done,
				Kind:      TaskEnum,
			}

			if err := AddTask(newTask, userIdOfContext(p)); err != nil {
				return nil, err
			}
			return newTask, nil
		},
	}

	addHabitMutation := &graphql.Field{
		Type: habitType,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.ID,
			},
			"title": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"interval": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(interval),
			},
			"frequency": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"done": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id, _ := p.Args["id"].(string)
			title, _ := p.Args["title"].(string)
			interval, _ := p.Args["interval"].(Interval)
			frequency, _ := p.Args["frequency"].(int)
			done, _ := p.Args["done"].(bool)

			newTask := &Task{
				Id:        id,
				Title:     title,
				Interval:  interval,
				Frequency: frequency,
				Done:      done,
				Kind:      HabitEnum,
			}

			if err := AddTask(newTask, userIdOfContext(p)); err != nil {
				return nil, err
			}
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
			taskDeleted, err := DeleteTask(id, userIdOfContext(p))
			if err != nil {
				return nil, err
			}
			if !taskDeleted {
				return nil, nil
			}
			return id, nil
		},
		Description: "Deletes a task or habit by ID",
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

			attrs := make(map[string]interface{})

			if title, ok := p.Args["title"].(string); ok {
				attrs["title"] = title
			}
			if startDate, ok := p.Args["start_date"].(time.Time); ok {
				attrs["start_date"] = startDate
			}
			if endDate, ok := p.Args["end_date"].(time.Time); ok {
				attrs["end_date"] = endDate
			}
			if done, ok := p.Args["done"].(bool); ok {
				attrs["done"] = done
			}

			return UpdateTask(id, userIdOfContext(p), attrs)
		},
	}

	updateHabitMutation := &graphql.Field{
		Type: habitType,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.ID,
			},
			"title": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"interval": &graphql.ArgumentConfig{
				Type: interval,
			},
			"frequency": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
			"done": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id, _ := p.Args["id"].(string)

			attrs := make(map[string]interface{})

			if title, ok := p.Args["title"].(string); ok {
				attrs["title"] = title
			}
			if interval, ok := p.Args["interval"].(Interval); ok {
				attrs["interval"] = interval
			}
			if frequency, ok := p.Args["frequency"].(int); ok {
				attrs["frequency"] = frequency
			}
			if done, ok := p.Args["done"].(bool); ok {
				attrs["done"] = done
			}

			return UpdateTask(id, userIdOfContext(p), attrs)
		},
	}

	addActionMutation := &graphql.Field{
		Type: actionType,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.ID,
			},
			"taskId": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"kind": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(actionKind),
			},
			"when": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(dateType),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id, _ := p.Args["id"].(string)
			taskId, _ := p.Args["taskId"].(string)
			kind, _ := p.Args["kind"].(ActionKind)
			when, _ := p.Args["when"].(*time.Time)

			newAction := &Action{
				Id:     id,
				Kind:   kind,
				When:   when,
				TaskId: taskId,
			}

			if err := AddAction(newAction, userIdOfContext(p)); err != nil {
				return nil, err
			}
			return newAction, nil
		},
	}

	deleteActionMutation := &graphql.Field{
		Type: graphql.NewObject(graphql.ObjectConfig{
			Name: "removeActionPayload",
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
				Type: graphql.ID,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id, _ := p.Args["id"].(string)

			if err := DeleteAction(id, userIdOfContext(p)); err != nil {
				return nil, err
			}
			return id, nil
		},
	}

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"task":   taskQuery,
			"tasks":  tasksQuery,
			"habit":  habitQuery,
			"habits": habitsQuery,
		},
	})

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootMutation",
		Fields: graphql.Fields{
			"addTask":      addTaskMutation,
			"deleteTask":   deleteTaskMutation,
			"updateTask":   updateTaskMutation,
			"addHabit":     addHabitMutation,
			"updateHabit":  updateHabitMutation,
			"addAction":    addActionMutation,
			"deleteAction": deleteActionMutation,
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
