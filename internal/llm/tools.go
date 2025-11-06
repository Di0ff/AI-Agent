package llm

import "github.com/sashabaranov/go-openai"

func getTools() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "click",
				Description: "Кликнуть по элементу на странице. Используй когда нужно нажать на кнопку, ссылку или другой интерактивный элемент.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"selector": map[string]interface{}{
							"type":        "string",
							"description": "CSS селектор элемента для клика (например: '#button', '.link', 'button[type=submit]')",
						},
						"reasoning": map[string]interface{}{
							"type":        "string",
							"description": "Объяснение почему нужно кликнуть именно по этому элементу",
						},
					},
					"required": []string{"selector", "reasoning"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "type",
				Description: "Ввести текст в поле ввода. Используй для заполнения форм, поисковых запросов и т.д.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"selector": map[string]interface{}{
							"type":        "string",
							"description": "CSS селектор поля ввода (например: '#search-input', 'input[name=email]')",
						},
						"value": map[string]interface{}{
							"type":        "string",
							"description": "Текст для ввода",
						},
						"reasoning": map[string]interface{}{
							"type":        "string",
							"description": "Объяснение что и зачем вводится",
						},
					},
					"required": []string{"selector", "value", "reasoning"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "navigate",
				Description: "Перейти на указанный URL. Используй для открытия новой страницы или перехода по ссылке.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"url": map[string]interface{}{
							"type":        "string",
							"description": "URL для перехода (например: 'https://example.com' или относительный путь '/page')",
						},
						"reasoning": map[string]interface{}{
							"type":        "string",
							"description": "Объяснение зачем нужен переход на эту страницу",
						},
					},
					"required": []string{"url", "reasoning"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "extract_info",
				Description: "Извлечь информацию со страницы. Используй когда нужно получить текст, данные или другую информацию с текущей страницы.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"selector": map[string]interface{}{
							"type":        "string",
							"description": "CSS селектор элемента для извлечения информации (например: '.price', '#title', 'article')",
						},
						"reasoning": map[string]interface{}{
							"type":        "string",
							"description": "Объяснение какую информацию нужно извлечь и зачем",
						},
					},
					"required": []string{"selector", "reasoning"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "ask_user",
				Description: "Спросить пользователя. Используй когда нужна дополнительная информация от пользователя для продолжения выполнения задачи.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"question": map[string]interface{}{
							"type":        "string",
							"description": "Вопрос для пользователя",
						},
						"reasoning": map[string]interface{}{
							"type":        "string",
							"description": "Объяснение почему нужна эта информация",
						},
					},
					"required": []string{"question", "reasoning"},
				},
			},
		},
	}
}
