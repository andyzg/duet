package data

import (
	"strconv"
	"time"
)

type Task struct {
	Id        string    `json:"id"`
	Title     string    `json:"title"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Done      bool      `json:"done"`
}

type User struct {
	Id string `json:"id"`
}

var tasks = map[string]*Task{
	"0": &Task{
		Id:        "0",
		Title:     "Get milk",
		StartDate: time.Now(),
		EndDate:   time.Now(),
		Done:      false,
	},
	"1": &Task{
		Id:        "1",
		Title:     "Call my momma",
		StartDate: time.Now(),
		EndDate:   time.Now(),
		Done:      true,
	},
}

func GetTask(id string) *Task {
	return tasks[id]
}

func GetTasks() []*Task {
	slice := make([]*Task, 0, len(tasks))
	for _, v := range tasks {
		slice = append(slice, v)
	}
	return slice
}

func AddTask(task *Task) {
	task.Id = strconv.Itoa(len(tasks))
	tasks[task.Id] = task
}
