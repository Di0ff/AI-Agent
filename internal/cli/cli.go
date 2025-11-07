package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"aiAgent/internal/agent"
	"aiAgent/internal/browser"
	"aiAgent/internal/database"
	"aiAgent/internal/llm"
	"aiAgent/internal/logger"

	"github.com/chzyer/readline"
	"go.uber.org/zap"
)

type CLI struct {
	repo      *database.TaskRepository
	log       *logger.Zap
	llmClient llm.LLMClient
	browser   browser.Browser
	agent     *agent.Agent
	rl        *readline.Instance
}

func (c *CLI) Run(ctx context.Context) {
	// Вывод логотипа
	logoBytes, err := os.ReadFile("logo.txt")
	if err == nil {
		fmt.Println("\033[36m" + string(logoBytes) + "\033[0m")
	}
	fmt.Println("\033[1m🤖 AI-Agent v0.1.0\033[0m")
	fmt.Println("\033[90mАвтономный AI-агент для управления браузером\033[0m")
	fmt.Println("\033[90mИспользуется: Firefox + OpenAI GPT-4o\033[0m")
	fmt.Println()
	fmt.Println("\033[33m📋 Доступные команды:\033[0m")
	fmt.Println("  \033[32mtask\033[0m <текст>        - Создать новую задачу")
	fmt.Println("  \033[32mtasks\033[0m               - Список всех задач")
	fmt.Println("  \033[32mrun\033[0m <id>            - Выполнить задачу")
	fmt.Println("  \033[32mstatus\033[0m <id>         - Статус задачи")
	fmt.Println("  \033[32mshow\033[0m <id>           - Детали задачи")
	fmt.Println("  \033[32mlogs\033[0m <id>           - LLM логи задачи")
	fmt.Println("  \033[32mtest-llm\033[0m <задача>   - Тест планирования LLM")
	fmt.Println("  \033[32mopen\033[0m <url>          - Открыть URL в браузере")
	fmt.Println("  \033[32mopen-persistent\033[0m     - Открыть браузер для ручной настройки")
	fmt.Println("  \033[32mclear\033[0m               - Очистить экран")
	fmt.Println("  \033[32mexit\033[0m                - Выход")
	fmt.Println()
	fmt.Println("\033[36m💡 Совет:\033[0m Используйте \033[33mopen-persistent\033[0m для входа на сайты, затем \033[33mrun\033[0m для выполнения задач")
	fmt.Println()
	fmt.Println("\033[90m⬆️ ⬇️\033[0m Используйте стрелки для навигации по истории команд")
	fmt.Println()

	defer c.rl.Close()

	for {
		line, err := c.rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		switch {
		case line == "exit":
			fmt.Println("\033[36m👋 До свидания!\033[0m")
			return

		case line == "clear":
			fmt.Print("\033[H\033[2J")

		case strings.HasPrefix(line, "task "):
			userInput := strings.TrimPrefix(line, "task ")
			task := database.Task{UserInput: userInput, Status: "pending"}
			if err := c.repo.CreateTask(&task); err != nil {
				c.log.Error("Ошибка создания задачи", zap.Error(err))
				fmt.Printf("\033[31m❌ Ошибка:\033[0m %v\n", err)
				continue
			}
			fmt.Printf("\033[32m✓ Создана задача #%d\033[0m\n", task.ID)

		case line == "tasks":
			tasks, err := c.repo.ListTasks(50, 0)
			if err != nil {
				c.log.Error("Ошибка чтения задач", zap.Error(err))
				fmt.Println("\033[31m❌ Ошибка чтения задач\033[0m")
				continue
			}
			fmt.Println("\n\033[1m📋 Список задач:\033[0m")
			fmt.Println()
			for _, t := range tasks {
				statusIcon := "⏳"
				statusColor := "\033[33m"
				statusText := "pending"
				switch t.Status {
				case "completed":
					statusIcon = "✓"
					statusColor = "\033[32m"
					statusText = "завершена"
				case "failed":
					statusIcon = "✗"
					statusColor = "\033[31m"
					statusText = "ошибка"
				case "running":
					statusIcon = "▶"
					statusColor = "\033[36m"
					statusText = "выполняется"
				case "pending":
					statusText = "ожидает"
				}
				fmt.Printf("  \033[1m#%d\033[0m %s%s %s\033[0m\n", t.ID, statusColor, statusIcon, statusText)
				fmt.Printf("  \033[90m└─\033[0m %s\n", t.UserInput)
				fmt.Println()
			}

		case strings.HasPrefix(line, "status "):
			idStr := strings.TrimPrefix(line, "status ")
			id, _ := strconv.Atoi(idStr)
			task, err := c.repo.GetTaskByID(uint(id))
			if err != nil {
				fmt.Println("\033[31m❌ Задача не найдена\033[0m")
				continue
			}
			statusIcon := "⏳"
			statusColor := "\033[33m"
			statusText := "ожидает"
			switch task.Status {
			case "completed":
				statusIcon = "✓"
				statusColor = "\033[32m"
				statusText = "завершена"
			case "failed":
				statusIcon = "✗"
				statusColor = "\033[31m"
				statusText = "ошибка"
			case "running":
				statusIcon = "▶"
				statusColor = "\033[36m"
				statusText = "выполняется"
			}
			fmt.Println()
			fmt.Printf("\033[1mЗадача #%d\033[0m %s%s %s\033[0m\n", task.ID, statusColor, statusIcon, statusText)
			fmt.Printf("  \033[36m📝\033[0m %s\n", task.UserInput)
			fmt.Printf("  \033[90m🕐\033[0m %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Println()

		case strings.HasPrefix(line, "show "):
			idStr := strings.TrimPrefix(line, "show ")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				fmt.Println("\033[31m❌ Неверный ID задачи\033[0m")
				continue
			}
			task, err := c.repo.GetTaskByID(uint(id))
			if err != nil {
				fmt.Println("\033[31m❌ Задача не найдена\033[0m")
				continue
			}

			statusText := task.Status
			switch task.Status {
			case "completed":
				statusText = "завершена"
			case "failed":
				statusText = "ошибка"
			case "running":
				statusText = "выполняется"
			case "pending":
				statusText = "ожидает"
			}

			fmt.Printf("\n\033[1m=== Задача #%d ===\033[0m\n", task.ID)
			fmt.Printf("\033[36m📝 Описание:\033[0m %s\n", task.UserInput)
			fmt.Printf("\033[36m📊 Статус:\033[0m %s\n", statusText)
			fmt.Printf("\033[36m🕐 Создана:\033[0m %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
			if task.ResultSummary != "" {
				fmt.Printf("\033[36m💬 Результат:\033[0m %s\n", task.ResultSummary)
			}

			steps, err := c.repo.GetStepsByTaskID(task.ID)
			if err != nil {
				c.log.Error("Ошибка получения шагов", zap.Error(err))
				fmt.Println("\033[31m❌ Ошибка получения шагов\033[0m")
				continue
			}

			if len(steps) > 0 {
				fmt.Printf("\n\033[33m🔄 Шаги выполнения (%d):\033[0m\n", len(steps))
				for _, step := range steps {
					fmt.Printf("\n\033[1m[Шаг %d]\033[0m \033[36m%s\033[0m\n", step.StepNo, step.ActionType)
					if step.TargetSelector != "" {
						fmt.Printf("  \033[90mСелектор:\033[0m %s\n", step.TargetSelector)
					}
					if step.Reasoning != "" {
						fmt.Printf("  \033[90mОбоснование:\033[0m %s\n", step.Reasoning)
					}
					if step.Result != "" {
						resultColor := "\033[32m"
						if strings.Contains(step.Result, "Ошибка") {
							resultColor = "\033[31m"
						}
						fmt.Printf("  %sРезультат:\033[0m %s\n", resultColor, step.Result)
					}
					fmt.Printf("  \033[90m🕐 %s\033[0m\n", step.CreatedAt.Format("15:04:05"))
				}
			} else {
				fmt.Println("\n\033[90mШаги не найдены\033[0m")
			}
			fmt.Println()

		case strings.HasPrefix(line, "logs "):
			idStr := strings.TrimPrefix(line, "logs ")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				fmt.Println("\033[31m❌ Неверный ID задачи\033[0m")
				continue
			}
			task, err := c.repo.GetTaskByID(uint(id))
			if err != nil {
				fmt.Println("\033[31m❌ Задача не найдена\033[0m")
				continue
			}

			fmt.Printf("\n\033[1m=== 📋 Логи задачи #%d ===\033[0m\n", task.ID)
			fmt.Printf("\033[36mЗадача:\033[0m %s\n", task.UserInput)
			fmt.Printf("\033[36mСтатус:\033[0m %s\n\n", task.Status)

			steps, err := c.repo.GetStepsByTaskID(task.ID)
			if err != nil {
				c.log.Error("Ошибка получения шагов", zap.Error(err))
				fmt.Println("\033[31m❌ Ошибка получения шагов\033[0m")
				continue
			}

			if len(steps) == 0 {
				fmt.Println("\033[90mШаги не найдены\033[0m")
				continue
			}

			for _, step := range steps {
				fmt.Printf("\033[90m[%s]\033[0m \033[36m%s\033[0m", step.CreatedAt.Format("15:04:05"), step.ActionType)
				if step.TargetSelector != "" {
					fmt.Printf(" → \033[33m%s\033[0m", step.TargetSelector)
				}
				fmt.Println()
				if step.Reasoning != "" && len(step.Reasoning) < 100 {
					fmt.Printf("  \033[90m%s\033[0m\n", step.Reasoning)
				}
				if step.Result != "" {
					if strings.HasPrefix(step.Result, "Ошибка") {
						fmt.Printf("  \033[31m[ОШИБКА]\033[0m %s\n", step.Result)
					} else if len(step.Result) < 80 {
						fmt.Printf("  \033[32m[OK]\033[0m %s\n", step.Result)
					}
				}
			}
			fmt.Println()

		case strings.HasPrefix(line, "run "):
			if c.agent == nil {
				fmt.Println("\033[31m❌ Агент не инициализирован\033[0m")
				continue
			}
			idStr := strings.TrimPrefix(line, "run ")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				fmt.Println("\033[31m❌ Неверный ID задачи\033[0m")
				continue
			}
			task, err := c.repo.GetTaskByID(uint(id))
			if err != nil {
				fmt.Println("\033[31m❌ Задача не найдена\033[0m")
				continue
			}
			fmt.Printf("\033[36m▶ Запуск задачи #%d:\033[0m %s\n", task.ID, task.UserInput)
			if err := c.agent.ExecuteTask(ctx, task); err != nil {
				fmt.Printf("\033[31m✗ Ошибка:\033[0m %v\n", err)
				c.repo.UpdateTaskStatus(task.ID, "failed", err.Error())
			} else {
				fmt.Println("\033[32m✓ Задача выполнена успешно!\033[0m")
			}

		case strings.HasPrefix(line, "test-llm "):
			if c.llmClient == nil {
				fmt.Println("\033[31m❌ LLM клиент не инициализирован\033[0m")
				continue
			}
			taskText := strings.TrimPrefix(line, "test-llm ")
			pageContext := "Страница: https://example.com\nЭлементы: кнопка 'Найти', поле ввода 'Поиск'"

			fmt.Println("\033[36m🤖 Запрос к OpenAI...\033[0m")
			plan, err := c.llmClient.PlanAction(ctx, taskText, pageContext, nil, nil)
			if err != nil {
				fmt.Printf("\033[31m❌ Ошибка:\033[0m %v\n", err)
				continue
			}

			fmt.Println("\033[32m✓ Получен план действия:\033[0m")
			fmt.Printf("  \033[36mДействие:\033[0m %s\n", plan.Action)
			if plan.Selector != "" {
				fmt.Printf("  \033[36mСелектор:\033[0m %s\n", plan.Selector)
			}
			if plan.Value != "" {
				fmt.Printf("  \033[36mЗначение:\033[0m %s\n", plan.Value)
			}
			if plan.Reasoning != "" {
				fmt.Printf("  \033[Обоснование:\033[0m %s\n", plan.Reasoning)
			}

		case line == "open-persistent":
			if c.browser == nil {
				fmt.Println("\033[31m❌ Браузер не инициализирован\033[0m")
				continue
			}

			fmt.Println("\033[36m🌐 Запуск браузера в persistent режиме...\033[0m")
			if err := c.browser.Launch(ctx); err != nil {
				fmt.Printf("\033[31m❌ Ошибка запуска:\033[0m %v\n", err)
				continue
			}

			fmt.Println("\033[32m✓ Браузер открыт с сохранением сессии\033[0m")
			fmt.Println("\033[90mВы можете вручную залогиниться и настроить браузер\033[0m")
			fmt.Println("\033[90mЗатем используйте команду '\033[33mrun <id>\033[90m' для выполнения задач\033[0m")
			fmt.Println("\033[33m⏎ Нажмите Enter для закрытия браузера...\033[0m")
			c.rl.Readline()
			c.browser.Close()
			fmt.Println("\033[32m✓ Сессия сохранена в ./browser-data\033[0m")

		case strings.HasPrefix(line, "open "):
			if c.browser == nil {
				fmt.Println("\033[31m❌ Браузер не инициализирован\033[0m")
				continue
			}
			url := strings.TrimPrefix(line, "open ")
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				url = "https://" + url
			}

			fmt.Println("\033[36m🌐 Запуск браузера...\033[0m")
			if err := c.browser.Launch(ctx); err != nil {
				fmt.Printf("\033[31m❌ Ошибка запуска:\033[0m %v\n", err)
				continue
			}

			fmt.Printf("\033[36m↗ Открытие %s...\033[0m\n", url)
			if err := c.browser.Navigate(ctx, url); err != nil {
				fmt.Printf("\033[31m❌ Ошибка навигации:\033[0m %v\n", err)
				c.browser.Close()
				continue
			}

			fmt.Println("\033[32m✓ Страница открыта\033[0m")
			fmt.Println("\033[33m⏎ Нажмите Enter для закрытия браузера...\033[0m")
			c.rl.Readline()
			c.browser.Close()
			fmt.Println("\033[90mБраузер закрыт\033[0m")

		default:
			fmt.Println()
			fmt.Println("Неизвестная команда")
			fmt.Println()
			fmt.Println("\033[33m📋 Доступные команды:\033[0m")
			fmt.Println("  \033[32mtask\033[0m <текст>        - Создать новую задачу")
			fmt.Println("  \033[32mtasks\033[0m               - Список всех задач")
			fmt.Println("  \033[32mrun\033[0m <id>            - Выполнить задачу")
			fmt.Println("  \033[32mstatus\033[0m <id>         - Статус задачи")
			fmt.Println("  \033[32mshow\033[0m <id>           - Детали задачи")
			fmt.Println("  \033[32mlogs\033[0m <id>           - LLM логи задачи")
			fmt.Println("  \033[32mtest-llm\033[0m <задача>   - Тест планирования LLM")
			fmt.Println("  \033[32mopen\033[0m <url>          - Открыть URL в браузере")
			fmt.Println("  \033[32mopen-persistent\033[0m     - Открыть браузер для ручной настройки")
			fmt.Println("  \033[32mclear\033[0m               - Очистить экран")
			fmt.Println("  \033[32mexit\033[0m                - Выход")
			fmt.Println()
		}
	}
}
