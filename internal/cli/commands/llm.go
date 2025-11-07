package commands

import (
	"context"
	"fmt"

	"aiAgent/internal/cli/ui"
	"aiAgent/internal/llm"
)

// LLMHandler обрабатывает команды тестирования LLM
type LLMHandler struct {
	llmClient llm.LLMClient
}

func NewLLMHandler(llmClient llm.LLMClient) *LLMHandler {
	return &LLMHandler{
		llmClient: llmClient,
	}
}

// TestPlan тестирует планирование LLM
func (h *LLMHandler) TestPlan(ctx context.Context, taskText string) {
	if h.llmClient == nil {
		fmt.Println(ui.ColorRed + ui.IconCross + " LLM клиент не инициализирован" + ui.ColorReset)
		return
	}
	pageContext := "Страница: https://example.com\nЭлементы: кнопка 'Найти', поле ввода 'Поиск'"

	fmt.Println(ui.ColorCyan + ui.IconRobot + " Запрос к OpenAI..." + ui.ColorReset)
	plan, err := h.llmClient.PlanAction(ctx, taskText, pageContext, nil, nil)
	if err != nil {
		fmt.Printf(ui.ColorRed+ui.IconCross+" Ошибка:"+ui.ColorReset+" %v\n", err)
		return
	}

	fmt.Println(ui.ColorGreen + ui.IconCheckmark + " Получен план действия:" + ui.ColorReset)
	fmt.Printf("  "+ui.ColorCyan+"Действие:"+ui.ColorReset+" %s\n", plan.Action)
	if plan.Selector != "" {
		fmt.Printf("  "+ui.ColorCyan+"Селектор:"+ui.ColorReset+" %s\n", plan.Selector)
	}
	if plan.Value != "" {
		fmt.Printf("  "+ui.ColorCyan+"Значение:"+ui.ColorReset+" %s\n", plan.Value)
	}
	if plan.Reasoning != "" {
		fmt.Printf("  "+ui.ColorCyan+"Обоснование:"+ui.ColorReset+" %s\n", plan.Reasoning)
	}
}
