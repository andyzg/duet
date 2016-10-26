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

func GetTask(id string) (*Task, error) {
	var task Task
	if err := db.Where(&User{Id: id}).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func GetTasks() []Task {
	var tasks []Task
	db.Find(&tasks)
	return tasks
}

func AddTask(task *Task) {
	db.Create(task) // TODO check status
}

func DeleteTask(id string) {
	task := &Task{
		Id: id,
	}
	db.Delete(task) // TODO check status
}

func UpdateTask(id string, attrs map[string]interface{}) *Task {
	task := Task{
		Id: id,
	}
	db.Model(&task).Updates(attrs) // TODO check status
	return &task
}
