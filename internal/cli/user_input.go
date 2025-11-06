package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"aiAgent/internal/agent"
	"aiAgent/internal/browser"
	"aiAgent/internal/database"
	"aiAgent/internal/llm"
	"aiAgent/internal/logger"
)

type userInputProvider struct {
	reader *bufio.Reader
}

func (p *userInputProvider) AskUser(ctx context.Context, question string) (string, error) {
	fmt.Printf("\n[Агент спрашивает] %s\n", question)
	fmt.Print("Ваш ответ: ")

	answerChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		answer, err := p.reader.ReadString('\n')
		if err != nil {
			errChan <- err
			return
		}
		answerChan <- strings.TrimSpace(answer)
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case err := <-errChan:
		return "", err
	case answer := <-answerChan:
		return answer, nil
	}
}

func New(repo *database.TaskRepository, log *logger.Zap, llmClient llm.LLMClient, br browser.Browser) *CLI {
	reader := bufio.NewReader(os.Stdin)
	userInput := &userInputProvider{reader: reader}

	ag := agent.New(br, llmClient, repo, log, agent.Config{
		MaxSteps:          50,
		MaxTokens:         2000,
		UserInputProvider: userInput,
	})
	return &CLI{
		repo:      repo,
		log:       log,
		llmClient: llmClient,
		browser:   br,
		agent:     ag,
		reader:    reader,
	}
}
