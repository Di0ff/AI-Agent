package commands

import (
	"context"
	"fmt"
	"strconv"

	"aiAgent/internal/agent"
	"aiAgent/internal/cli/ui"
	"aiAgent/internal/database"

	"go.uber.org/zap"
)

// TaskHandler обрабатывает команды связанные с задачами
type TaskHandler struct {
	repo  *database.TaskRepository
	agent *agent.Agent
	log   *zap.Logger
}

func NewTaskHandler(repo *database.TaskRepository, agent *agent.Agent, log *zap.Logger) *TaskHandler {
	return &TaskHandler{
		repo:  repo,
		agent: agent,
		log:   log,
	}
}

// Create создает новую задачу
func (h *TaskHandler) Create(userInput string) {
	task := database.Task{UserInput: userInput, Status: "pending"}
	if err := h.repo.CreateTask(&task); err != nil {
		h.log.Error("Ошибка создания задачи", zap.Error(err))
		fmt.Printf(ui.ColorRed+ui.IconCross+" Ошибка:"+ui.ColorReset+" %v\n", err)
		return
	}
	fmt.Printf(ui.ColorGreen+ui.IconCheckmark+" Создана задача #%d"+ui.ColorReset+"\n", task.ID)
}

// List выводит список всех задач
func (h *TaskHandler) List() {
	tasks, err := h.repo.ListTasks(50, 0)
	if err != nil {
		h.log.Error("Ошибка чтения задач", zap.Error(err))
		fmt.Println(ui.ColorRed + ui.IconCross + " Ошибка чтения задач" + ui.ColorReset)
		return
	}
	fmt.Println("\n" + ui.ColorBold + ui.IconList + " Список задач:" + ui.ColorReset)
	fmt.Println()
	for _, t := range tasks {
		icon, color, text := ui.FormatStatus(t.Status)
		fmt.Printf("  "+ui.ColorBold+"#%d"+ui.ColorReset+" %s%s %s"+ui.ColorReset+"\n", t.ID, color, icon, text)
		fmt.Printf("  "+ui.ColorGray+"└─"+ui.ColorReset+" %s\n", t.UserInput)
		fmt.Println()
	}
}

// Status показывает статус задачи
func (h *TaskHandler) Status(idStr string) {
	id, _ := strconv.Atoi(idStr)
	task, err := h.repo.GetTaskByID(uint(id))
	if err != nil {
		fmt.Println(ui.ColorRed + ui.IconCross + " Задача не найдена" + ui.ColorReset)
		return
	}
	icon, color, text := ui.FormatStatus(task.Status)
	fmt.Println()
	fmt.Printf(ui.ColorBold+"Задача #%d"+ui.ColorReset+" %s%s %s"+ui.ColorReset+"\n", task.ID, color, icon, text)
	fmt.Printf("  "+ui.ColorCyan+ui.IconDocument+ui.ColorReset+" %s\n", task.UserInput)
	fmt.Printf("  "+ui.ColorGray+ui.IconTime+ui.ColorReset+" %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()
}

// Run выполняет задачу
func (h *TaskHandler) Run(ctx context.Context, idStr string) {
	if h.agent == nil {
		fmt.Println(ui.ColorRed + ui.IconCross + " Агент не инициализирован" + ui.ColorReset)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println(ui.ColorRed + ui.IconCross + " Неверный ID задачи" + ui.ColorReset)
		return
	}
	task, err := h.repo.GetTaskByID(uint(id))
	if err != nil {
		fmt.Println(ui.ColorRed + ui.IconCross + " Задача не найдена" + ui.ColorReset)
		return
	}
	fmt.Printf(ui.ColorCyan+ui.IconPlay+" Запуск задачи #%d:"+ui.ColorReset+" %s\n", task.ID, task.UserInput)
	if err := h.agent.ExecuteTask(ctx, task); err != nil {
		fmt.Printf(ui.ColorRed+ui.IconCross+" Ошибка:"+ui.ColorReset+" %v\n", err)
		h.repo.UpdateTaskStatus(task.ID, "failed", err.Error())
	} else {
		fmt.Println(ui.ColorGreen + ui.IconCheckmark + " Задача выполнена успешно!" + ui.ColorReset)
	}
}
