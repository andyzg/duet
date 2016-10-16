package data

import (
	"time"
)

type Task struct {
	Id        int64     `json:"id"`
	Title     string    `json:"title"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Done      bool      `json:"done"`
}

type User struct {
	Id int64 `json:"id"`
}

var tasks = []*Task{
	&Task{
		Id:        0,
		Title:     "Get milk",
		StartDate: time.Now(),
		EndDate:   time.Now(),
		Done:      false,
	},
	&Task{
		Id:        1,
		Title:     "Call my momma",
		StartDate: time.Now(),
		EndDate:   time.Now(),
		Done:      true,
	},
}

func GetTask(id int64) *Task {
	return tasks[id]
}

func GetTasks() []*Task {
	return tasks
}
