package data

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Task struct {
	Id        string `json:"id" gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Title     string     `json:"title" gorm:"not_null"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	Done      bool       `json:"done" gorm:"not_null;default:false"`
	UserId    uint64     `json:"user_id" gorm:"not_null"`
}

type ActionKind int

const (
	ActionProgress ActionKind = iota
	ActionDefer
)

type Action struct {
	Id     string     `json:"id" gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	Kind   ActionKind `gorm:"not_null"`
	When   time.Time  `gorm:"not_null"`
	TaskId string     `gorm:"not_null;type:uuid"`
}

type User struct {
	Id             uint64 `json:"id" gorm:"primary_key"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
	Username       string `json:"username" gorm:"not_null;unique"`
	HashedPassword []byte `json:"-" gorm:"not_null"`
	Tasks          []Task `json:"-" gorm:"ForeignKey:UserId"`
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

func GetTask(taskId string, userId uint64) (*Task, error) {
	task := Task{
		Id: taskId,
	}
	if err := db.Model(&User{Id: userId}).Related(&task).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func GetTasks(userId uint64) ([]Task, error) {
	var tasks []Task
	if err := db.Model(&User{Id: userId}).Related(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func AddTask(task *Task, userId uint64) error {
	task.UserId = userId
	return db.Create(task).Error
}

// Deletes the task with the given ID and returns whether a row was deleted.
func DeleteTask(taskId string, userId uint64) (bool, error) {
	task := Task{
		Id:     taskId,
		UserId: userId,
	}
	result := db.Where(&task).Delete(&task)
	if err := result.Error; err != nil {
		return false, err
	}
	return result.RowsAffected > 0, nil
}

// Updates a task with the given attributes and returns the updated Task if one exists for the ID.
func UpdateTask(taskId string, userId uint64, attrs map[string]interface{}) (*Task, error) {
	task := Task{
		Id: taskId,
	}
	result := db.Model(&task).Where("user_id = ?", userId).Updates(attrs)
	if err := result.Error; err != nil {
		return nil, err
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("Task ID \"%s\" does not exist for user \"%d\"", taskId, userId)
	}
	return &task, nil
}

func AddUser(user *User) error {
	return db.Create(user).Error
}

func GetUserById(id uint64) (*User, error) {
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

func AddAction(action *Action, userId uint64) error {
	task, err := GetTask(action.TaskId, userId)
	if task == nil {
		return fmt.Errorf("Task %s does not exist for user %d", action.TaskId, userId)
	}
	if err != nil {
		return err
	}
	return db.Create(action).Error
}
