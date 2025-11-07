package cli

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"

	"aiAgent/internal/agent"
	"aiAgent/internal/browser"
	"aiAgent/internal/cli/commands"
	"aiAgent/internal/cli/ui"
	"aiAgent/internal/database"
	"aiAgent/internal/llm"
	"aiAgent/internal/logger"

	"github.com/chzyer/readline"
)

type CLI struct {
	repo           *database.TaskRepository
	log            *logger.Zap
	llmClient      llm.LLMClient
	browser        browser.Browser
	agent          *agent.Agent
	rl             *readline.Instance
	taskHandler    *commands.TaskHandler
	showHandler    *commands.ShowHandler
	logsHandler    *commands.LogsHandler
	browserHandler *commands.BrowserHandler
	llmHandler     *commands.LLMHandler
}

func New(repo *database.TaskRepository, log *logger.Zap, llmClient llm.LLMClient, br browser.Browser, ag *agent.Agent) *CLI {
	cli := &CLI{
		repo:      repo,
		log:       log,
		llmClient: llmClient,
		browser:   br,
		agent:     ag,
	}

	// Инициализация handlers
	cli.taskHandler = commands.NewTaskHandler(repo, ag, log.Logger)
	cli.showHandler = commands.NewShowHandler(repo, log.Logger)
	cli.logsHandler = commands.NewLogsHandler(repo, log.Logger)
	cli.browserHandler = commands.NewBrowserHandler(br, cli.readLine)
	cli.llmHandler = commands.NewLLMHandler(llmClient)

	// Инициализация readline
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		HistoryFile:     ".ai-agent-history",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		log.Warn("Не удалось инициализировать readline, будет использован fallback режим")
	} else {
		cli.rl = rl
	}

	return cli
}

func (c *CLI) readLine() (string, error) {
	if c.rl != nil {
		return c.rl.Readline()
	}
	// Fallback для работы без readline
	reader := bufio.NewReader(os.Stdin)
	println(ui.ColorCyan + "> " + ui.ColorReset)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func (c *CLI) closeReadline() {
	if c.rl != nil {
		c.rl.Close()
	}
}

func (c *CLI) Run(ctx context.Context) {
	ui.PrintWelcome()
	defer c.closeReadline()

	for {
		// Проверка отмены контекста
		select {
		case <-ctx.Done():
			println("\n" + ui.ColorCyan + ui.IconWave + " Получен сигнал завершения..." + ui.ColorReset)
			return
		default:
		}

		line, err := c.readLine()
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

		c.handleCommand(ctx, line)
	}
}

func (c *CLI) handleCommand(ctx context.Context, line string) {
	switch {
	case line == "exit":
		println(ui.ColorCyan + ui.IconWave + " До свидания!" + ui.ColorReset)
		os.Exit(0)

	case line == "clear":
		ui.ClearScreen()

	case strings.HasPrefix(line, "task "):
		userInput := strings.TrimPrefix(line, "task ")
		c.taskHandler.Create(userInput)

	case line == "tasks":
		c.taskHandler.List()

	case strings.HasPrefix(line, "status "):
		idStr := strings.TrimPrefix(line, "status ")
		c.taskHandler.Status(idStr)

	case strings.HasPrefix(line, "show "):
		idStr := strings.TrimPrefix(line, "show ")
		c.showHandler.Show(idStr)

	case strings.HasPrefix(line, "logs "):
		idStr := strings.TrimPrefix(line, "logs ")
		c.logsHandler.Show(idStr)

	case strings.HasPrefix(line, "run "):
		idStr := strings.TrimPrefix(line, "run ")
		c.taskHandler.Run(ctx, idStr)

	case strings.HasPrefix(line, "test-llm "):
		taskText := strings.TrimPrefix(line, "test-llm ")
		c.llmHandler.TestPlan(ctx, taskText)

	case line == "open-persistent":
		c.browserHandler.OpenPersistent(ctx)

	case strings.HasPrefix(line, "open "):
		url := strings.TrimPrefix(line, "open ")
		c.browserHandler.Open(ctx, url)

	default:
		ui.PrintHelp()
	}
}
