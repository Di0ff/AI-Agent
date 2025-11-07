package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

type userInputProvider struct {
	reader *bufio.Reader
}

func NewUserInputProvider() *userInputProvider {
	return &userInputProvider{
		reader: bufio.NewReader(os.Stdin),
	}
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
