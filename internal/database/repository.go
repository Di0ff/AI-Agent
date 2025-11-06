package database

import (
	"context"

	"gorm.io/gorm"
)

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

func (r *TaskRepository) LogLLMRequest(ctx context.Context, taskID *uint, stepID *uint, role, promptText, responseText, model string, tokensUsed int) error {
	log := &LlmLog{
		TaskID:       taskID,
		StepID:       stepID,
		Role:         role,
		PromptText:   promptText,
		ResponseText: responseText,
		Model:        model,
		TokensUsed:   tokensUsed,
	}
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *TaskRepository) CreateStep(step *AgentStep) error {
	return r.db.Create(step).Error
}

func (r *TaskRepository) GetStepsByTaskID(taskID uint) ([]AgentStep, error) {
	var steps []AgentStep
	if err := r.db.Where("task_id = ?", taskID).Order("step_no ASC").Find(&steps).Error; err != nil {
		return nil, err
	}
	return steps, nil
}
