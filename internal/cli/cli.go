package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"aiAgent/internal/database"
	"aiAgent/internal/logger"

	"go.uber.org/zap"
)

type CLI struct {
	repo *database.TaskRepository
	log  *logger.Zap
}

func New(repo *database.TaskRepository, log *logger.Zap) *CLI {
	return &CLI{repo: repo, log: log}
}

func (c *CLI) Run(ctx context.Context) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("🤖  AI-Agent консоль запущена")
	fmt.Println("Команды: task <текст>, tasks, status <id>, exit")

	for {
		fmt.Print("> ")
		line, _ := reader.ReadString('\n')
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
			fmt.Println("📋 Список задач:")
			for _, t := range tasks {
				fmt.Printf("#%d | %-40s | %s\n", t.ID, t.UserInput, t.Status)
			}

		case strings.HasPrefix(line, "status "):
			idStr := strings.TrimPrefix(line, "status ")
			id, _ := strconv.Atoi(idStr)
			task, err := c.repo.GetTaskByID(uint(id))
			if err != nil {
				fmt.Println("❌ Задача не найдена")
				continue
			}
			fmt.Printf("#%d | %s | %s | %s\n", task.ID, task.UserInput, task.Status, task.CreatedAt.Format("15:04:05"))

		default:
			fmt.Println("Неизвестная команда. Доступно: task <текст>, tasks, status <id>, exit")
		}
	}
}
