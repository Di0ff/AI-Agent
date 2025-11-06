package database

import "gorm.io/gorm"

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) CreateTask(t *Task) error {
	return r.db.Create(t).Error
}

func (r *TaskRepository) GetTaskByID(id uint) (*Task, error) {
	var task Task
	if err := r.db.First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) ListTasks(limit, offset int) ([]Task, error) {
	var tasks []Task
	if err := r.db.Order("id DESC").Limit(limit).Offset(offset).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *TaskRepository) UpdateTaskStatus(id uint, status, summary string) error {
	return r.db.Model(&Task{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":         status,
			"result_summary": summary,
		}).Error
}
