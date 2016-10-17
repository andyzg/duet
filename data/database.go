package data

import (
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// TODO: Fix null dates
type Task struct {
	Id        string    `json:"id" gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	Title     string    `json:"title" gorm:"not_null"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Done      bool      `json:"done" gorm:"not_null;default:false"`
}

type User struct {
	Id string `json:"id" gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
}

// TODO remove
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

var db *gorm.DB

func InitDatabase() {
	var err error
	db, err = gorm.Open("postgres", "host=localhost user=duet DB.name=duet sslmode=disable")
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&Task{}, &User{})
}

func CloseDatabase() {
	db.Close()
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
	db.Create(task) // TODO check status
	tasks[task.Id] = task
}

func DeleteTask(id string) {
	task := &Task{
		Id: id,
	}
	db.Delete(task) // TODO check status
	delete(tasks, id)
}
