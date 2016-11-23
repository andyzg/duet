package data

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type TaskKind int

const (
	TaskEnum TaskKind = iota
	HabitEnum
)

type Interval int

const (
	Daily Interval = iota
	Weekly
	Monthly
)

type Task struct {
	// Common fields
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time
	Id        string   `json:"id" gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	Kind      TaskKind `json:"kind" gorm:"not_null"`
	Title     string   `json:"title" gorm:"not_null"`
	Done      bool     `json:"done" gorm:"not_null;default:false"`
	UserId    uint64   `json:"user_id" gorm:"not_null"`
	Actions   []Action `json:"actions" gorm:"ForeignKey:TaskId"`
	// Task Fields
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	// Habit Fields
	Interval  Interval `json:"interval"`
	Frequency int      `json:"frequency"`
}

type ActionKind int

const (
	ActionProgress ActionKind = iota
	ActionDefer
	ActionDone
)

type Action struct {
	Id     string     `json:"id" gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	Kind   ActionKind `json:"kind" gorm:"not_null"`
	When   *time.Time `json:"when" gorm:"not_null"`
	TaskId string     `json:"task_id" gorm:"not_null;type:uuid"`
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
	db.AutoMigrate(&Task{}, &User{}, &Action{})
}

func CloseDatabase() {
	db.Close()
}

func GetTask(taskId string, userId uint64, kind *TaskKind) (*Task, error) {
	whereFields := map[string]interface{}{
		"id":      taskId,
		"user_id": userId,
	}
	if kind != nil {
		whereFields["kind"] = *kind
	}

	var task Task
	// TODO: Only preload actions if necessary
	if err := db.Preload("Actions").Where(whereFields).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func GetTasks(userId uint64, kind *TaskKind) ([]Task, error) {
	whereFields := map[string]interface{}{
		"user_id": userId,
	}
	if kind != nil {
		whereFields["kind"] = *kind
	}

	var tasks []Task
	// TODO: Only preload actions if necessary
	if err := db.Preload("Actions").Where(whereFields).Find(&tasks).Error; err != nil {
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
	// TODO: Only query actions if necessary
	if err := db.Model(&task).Related(&task.Actions).Error; err != nil {
		return nil, err
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
	task, err := GetTask(action.TaskId, userId, nil)
	if task == nil {
		return fmt.Errorf("Task %s does not exist for user %d", action.TaskId, userId)
	}
	if err != nil {
		return err
	}
	return db.Create(action).Error
}

func DeleteAction(id string, userId uint64) error {
	action := &Action{
		Id: id,
	}
	if err := db.Where(action).First(action).Error; err != nil {
		return err
	}
	task, err := GetTask(action.TaskId, userId, nil)
	if err != nil {
		return err
	}
	if task == nil {
		return fmt.Errorf("Not authorized to delete action %s", id)
	}
	return db.Delete(action).Error
}
