package commands

import (
	"fmt"
	"strconv"
	"strings"

	"aiAgent/internal/cli/ui"
	"aiAgent/internal/database"

	"go.uber.org/zap"
)

// LogsHandler обрабатывает команды просмотра логов
type LogsHandler struct {
	repo *database.TaskRepository
	log  *zap.Logger
}

func NewLogsHandler(repo *database.TaskRepository, log *zap.Logger) *LogsHandler {
	return &LogsHandler{
		repo: repo,
		log:  log,
	}
}

// Show выводит логи задачи
func (h *LogsHandler) Show(idStr string) {
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

	fmt.Printf("\n"+ui.ColorBold+"=== "+ui.IconList+" Логи задачи #%d ==="+ui.ColorReset+"\n", task.ID)
	fmt.Printf(ui.ColorCyan+"Задача:"+ui.ColorReset+" %s\n", task.UserInput)
	fmt.Printf(ui.ColorCyan+"Статус:"+ui.ColorReset+" %s\n\n", task.Status)

	steps, err := h.repo.GetStepsByTaskID(task.ID)
	if err != nil {
		h.log.Error("Ошибка получения шагов", zap.Error(err))
		fmt.Println(ui.ColorRed + ui.IconCross + " Ошибка получения шагов" + ui.ColorReset)
		return
	}

	if len(steps) == 0 {
		fmt.Println(ui.ColorGray + "Шаги не найдены" + ui.ColorReset)
		return
	}

	for _, step := range steps {
		fmt.Printf(ui.ColorGray+"[%s]"+ui.ColorReset+" "+ui.ColorCyan+"%s"+ui.ColorReset, step.CreatedAt.Format("15:04:05"), step.ActionType)
		if step.TargetSelector != "" {
			fmt.Printf(" → "+ui.ColorYellow+"%s"+ui.ColorReset, step.TargetSelector)
		}
		fmt.Println()
		if step.Reasoning != "" && len(step.Reasoning) < 100 {
			fmt.Printf("  "+ui.ColorGray+"%s"+ui.ColorReset+"\n", step.Reasoning)
		}
		if step.Result != "" {
			if strings.HasPrefix(step.Result, "Ошибка") {
				fmt.Printf("  "+ui.ColorRed+"[ОШИБКА]"+ui.ColorReset+" %s\n", step.Result)
			} else if len(step.Result) < 80 {
				fmt.Printf("  "+ui.ColorGreen+"[OK]"+ui.ColorReset+" %s\n", step.Result)
			}
		}
	}
	fmt.Println()
}
