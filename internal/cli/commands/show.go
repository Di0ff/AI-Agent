package commands

import (
	"fmt"
	"strconv"
	"strings"

	"aiAgent/internal/cli/ui"
	"aiAgent/internal/database"

	"go.uber.org/zap"
)

// ShowHandler обрабатывает команды просмотра деталей
type ShowHandler struct {
	repo *database.TaskRepository
	log  *zap.Logger
}

func NewShowHandler(repo *database.TaskRepository, log *zap.Logger) *ShowHandler {
	return &ShowHandler{
		repo: repo,
		log:  log,
	}
}

// Show выводит детали задачи со всеми шагами
func (h *ShowHandler) Show(idStr string) {
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

	_, _, statusText := ui.FormatStatus(task.Status)

	fmt.Printf("\n"+ui.ColorBold+"=== Задача #%d ==="+ui.ColorReset+"\n", task.ID)
	fmt.Printf(ui.ColorCyan+ui.IconDocument+" Описание:"+ui.ColorReset+" %s\n", task.UserInput)
	fmt.Printf(ui.ColorCyan+ui.IconChart+" Статус:"+ui.ColorReset+" %s\n", statusText)
	fmt.Printf(ui.ColorCyan+ui.IconTime+" Создана:"+ui.ColorReset+" %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
	if task.ResultSummary != "" {
		fmt.Printf(ui.ColorCyan+ui.IconChat+" Результат:"+ui.ColorReset+" %s\n", task.ResultSummary)
	}

	steps, err := h.repo.GetStepsByTaskID(task.ID)
	if err != nil {
		h.log.Error("Ошибка получения шагов", zap.Error(err))
		fmt.Println(ui.ColorRed + ui.IconCross + " Ошибка получения шагов" + ui.ColorReset)
		return
	}

	if len(steps) > 0 {
		fmt.Printf("\n"+ui.ColorYellow+ui.IconLoop+" Шаги выполнения (%d):"+ui.ColorReset+"\n", len(steps))
		for _, step := range steps {
			fmt.Printf("\n"+ui.ColorBold+"[Шаг %d]"+ui.ColorReset+" "+ui.ColorCyan+"%s"+ui.ColorReset+"\n", step.StepNo, step.ActionType)
			if step.TargetSelector != "" {
				fmt.Printf("  "+ui.ColorGray+"Селектор:"+ui.ColorReset+" %s\n", step.TargetSelector)
			}
			if step.Reasoning != "" {
				fmt.Printf("  "+ui.ColorGray+"Обоснование:"+ui.ColorReset+" %s\n", step.Reasoning)
			}
			if step.Result != "" {
				resultColor := ui.ColorGreen
				if strings.Contains(step.Result, "Ошибка") {
					resultColor = ui.ColorRed
				}
				fmt.Printf("  %sРезультат:"+ui.ColorReset+" %s\n", resultColor, step.Result)
			}
			fmt.Printf("  "+ui.ColorGray+ui.IconTime+" %s"+ui.ColorReset+"\n", step.CreatedAt.Format("15:04:05"))
		}
	} else {
		fmt.Println("\n" + ui.ColorGray + "Шаги не найдены" + ui.ColorReset)
	}
	fmt.Println()
}
