package data

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"
)

type Task struct {
	Id        string      `json:"id" gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	Title     string      `json:"title" gorm:"not_null"`
	StartDate pq.NullTime `json:"start_date"`
	EndDate   pq.NullTime `json:"end_date"`
	Done      bool        `json:"done" gorm:"not_null;default:false"`
}

type User struct {
	Id             string `json:"id" gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	Username       string `gorm:"not_null;unique"`
	HashedPassword []byte `gorm:"not_null"`
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

func GetTasks() ([]Task, error) {
	var tasks []Task
	if err := db.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func AddTask(task *Task) error {
	return db.Create(task).Error
}

// Deletes the task with the given ID and returns whether a row was deleted.
func DeleteTask(id string) (bool, error) {
	task := &Task{
		Id: id,
	}
	result := db.Delete(task)
	if err := result.Error; err != nil {
		return false, err
	}
	return result.RowsAffected > 0, nil
}

// Updates a task with the given attributes and returns the updated Task if one exists for the ID.
func UpdateTask(id string, attrs map[string]interface{}) (*Task, error) {
	task := Task{
		Id: id,
	}
	result := db.Model(&task).Updates(attrs)
	if err := result.Error; err != nil {
		return nil, err
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &task, nil
}

func AddUser(user *User) error {
	return db.Create(user).Error
}

func GetUserById(id string) (*User, error) {
	user := &User{
		Id: id,
	}
	if err := db.Where(user).First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByUsername(username string) (*User, error) {
	user := &User{
		Username: username,
	}
	if err := db.Where(user).First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}
