package cli

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"strings"

	"aiAgent/internal/agent"
	"aiAgent/internal/browser"
	"aiAgent/internal/database"
	"aiAgent/internal/llm"
	"aiAgent/internal/logger"

	"go.uber.org/zap"
)

type CLI struct {
	repo      *database.TaskRepository
	log       *logger.Zap
	llmClient llm.LLMClient
	browser   browser.Browser
	agent     *agent.Agent
	reader    *bufio.Reader
}

func (c *CLI) Run(ctx context.Context) {
	fmt.Println("AI-Agent консоль запущена")
	fmt.Println("Команды: task <текст>, tasks, status <id>, show <id>, run <task_id>, test-llm <задача>, open <url>, exit")

	for {
		fmt.Print("> ")
		line, _ := c.reader.ReadString('\n')
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		switch {
		case line == "exit":
			fmt.Println("Выход.")
			return

		case strings.HasPrefix(line, "task "):
			userInput := strings.TrimPrefix(line, "task ")
			task := database.Task{UserInput: userInput, Status: "pending"}
			if err := c.repo.CreateTask(&task); err != nil {
				c.log.Error("Ошибка создания задачи", zap.Error(err))
				fmt.Println("Ошибка при создании задачи:", err)
				continue
			}
			fmt.Printf("[OK] Создана задача #%d\n", task.ID)

		case line == "tasks":
			tasks, err := c.repo.ListTasks(50, 0)
			if err != nil {
				c.log.Error("Ошибка чтения задач", zap.Error(err))
				continue
			}
			fmt.Println("Список задач:")
			for _, t := range tasks {
				fmt.Printf("#%d | %-40s | %s\n", t.ID, t.UserInput, t.Status)
			}

		case strings.HasPrefix(line, "status "):
			idStr := strings.TrimPrefix(line, "status ")
			id, _ := strconv.Atoi(idStr)
			task, err := c.repo.GetTaskByID(uint(id))
			if err != nil {
				fmt.Println("Задача не найдена")
				continue
			}
			fmt.Printf("#%d | %s | %s | %s\n", task.ID, task.UserInput, task.Status, task.CreatedAt.Format("15:04:05"))

		case strings.HasPrefix(line, "show "):
			idStr := strings.TrimPrefix(line, "show ")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				fmt.Println("Неверный ID задачи")
				continue
			}
			task, err := c.repo.GetTaskByID(uint(id))
			if err != nil {
				fmt.Println("Задача не найдена")
				continue
			}

			fmt.Printf("\n=== Задача #%d ===\n", task.ID)
			fmt.Printf("Описание: %s\n", task.UserInput)
			fmt.Printf("Статус: %s\n", task.Status)
			fmt.Printf("Создана: %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
			if task.ResultSummary != "" {
				fmt.Printf("Результат: %s\n", task.ResultSummary)
			}

			steps, err := c.repo.GetStepsByTaskID(task.ID)
			if err != nil {
				c.log.Error("Ошибка получения шагов", zap.Error(err))
				fmt.Println("Ошибка получения шагов")
				continue
			}

			if len(steps) > 0 {
				fmt.Printf("\nШаги выполнения (%d):\n", len(steps))
				for _, step := range steps {
					fmt.Printf("\n[Шаг %d] %s\n", step.StepNo, step.ActionType)
					if step.TargetSelector != "" {
						fmt.Printf("  Селектор: %s\n", step.TargetSelector)
					}
					if step.Reasoning != "" {
						fmt.Printf("  Обоснование: %s\n", step.Reasoning)
					}
					if step.Result != "" {
						fmt.Printf("  Результат: %s\n", step.Result)
					}
					fmt.Printf("  Время: %s\n", step.CreatedAt.Format("15:04:05"))
				}
			} else {
				fmt.Println("\nШаги не найдены")
			}
			fmt.Println()

		case strings.HasPrefix(line, "run "):
			if c.agent == nil {
				fmt.Println("Агент не инициализирован")
				continue
			}
			idStr := strings.TrimPrefix(line, "run ")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				fmt.Println("Неверный ID задачи")
				continue
			}
			task, err := c.repo.GetTaskByID(uint(id))
			if err != nil {
				fmt.Println("Задача не найдена")
				continue
			}
			fmt.Printf("Запуск задачи #%d: %s\n", task.ID, task.UserInput)
			if err := c.agent.ExecuteTask(ctx, task); err != nil {
				fmt.Printf("Ошибка выполнения задачи: %v\n", err)
				c.repo.UpdateTaskStatus(task.ID, "failed", err.Error())
			} else {
				fmt.Println("Задача выполнена успешно")
			}

		case strings.HasPrefix(line, "test-llm "):
			if c.llmClient == nil {
				fmt.Println("LLM клиент не инициализирован")
				continue
			}
			taskText := strings.TrimPrefix(line, "test-llm ")
			pageContext := "Страница: https://example.com\nЭлементы: кнопка 'Найти', поле ввода 'Поиск'"

			fmt.Println("Запрос к OpenAI...")
			plan, err := c.llmClient.PlanAction(ctx, taskText, pageContext, nil, nil)
			if err != nil {
				fmt.Printf("Ошибка: %v\n", err)
				continue
			}

			fmt.Println("Получен план действия:")
			fmt.Printf("  Действие: %s\n", plan.Action)
			if plan.Selector != "" {
				fmt.Printf("  Селектор: %s\n", plan.Selector)
			}
			if plan.Value != "" {
				fmt.Printf("  Значение: %s\n", plan.Value)
			}
			if plan.Reasoning != "" {
				fmt.Printf("  Обоснование: %s\n", plan.Reasoning)
			}

		case strings.HasPrefix(line, "open "):
			if c.browser == nil {
				fmt.Println("Браузер не инициализирован")
				continue
			}
			url := strings.TrimPrefix(line, "open ")
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				url = "https://" + url
			}

			fmt.Printf("Запуск браузера...\n")
			if err := c.browser.Launch(ctx); err != nil {
				fmt.Printf("Ошибка запуска браузера: %v\n", err)
				continue
			}

			fmt.Printf("Открытие %s...\n", url)
			if err := c.browser.Navigate(ctx, url); err != nil {
				fmt.Printf("Ошибка навигации: %v\n", err)
				c.browser.Close()
				continue
			}

			fmt.Println("Страница открыта. Нажмите Enter для закрытия браузера...")
			c.reader.ReadString('\n')
			c.browser.Close()
			fmt.Println("Браузер закрыт")

		default:
			fmt.Println("Неизвестная команда. Доступно: task <текст>, tasks, status <id>, show <id>, run <task_id>, test-llm <задача>, open <url>, exit")
		}
	}
}
