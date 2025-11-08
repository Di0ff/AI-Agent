package agent

import (
	"context"
	"strings"
)

// FoodDeliveryAgent - специализированный агент для оформления заказов на доставку еды
// Ключевые слова: "еда", "заказ", "ресторан", "доставка", "яндекс.еда", "delivery club"
type FoodDeliveryAgent struct {
	baseAgent *Agent
	keywords  []string
}

// NewFoodDeliveryAgent создает новый FoodDeliveryAgent
func NewFoodDeliveryAgent(baseAgent *Agent) *FoodDeliveryAgent {
	return &FoodDeliveryAgent{
		baseAgent: baseAgent,
		keywords: []string{
			// Русские ключевые слова
			"еда", "заказ", "заказать", "ресторан", "доставка", "доставить",
			"яндекс.еда", "яндекс еда", "delivery club", "деливери клаб",
			"бургер", "пицца", "суши", "роллы", "шаурма", "корзина",
			// Английские ключевые слова
			"food", "order", "restaurant", "delivery", "deliver",
			"burger", "pizza", "sushi", "cart", "checkout",
		},
	}
}

// CanHandle проверяет, может ли FoodDeliveryAgent обработать данную задачу
func (a *FoodDeliveryAgent) CanHandle(ctx context.Context, task string, pageContext string) (float64, error) {
	taskLower := strings.ToLower(task)
	contextLower := strings.ToLower(pageContext)

	// Проверяем ключевые слова в задаче
	taskScore := 0.0
	for _, keyword := range a.keywords {
		if strings.Contains(taskLower, keyword) {
			taskScore = 0.9
			break
		}
	}

	// Проверяем контекст страницы (сервисы доставки еды)
	contextScore := 0.0
	foodIndicators := []string{
		"eda.yandex", "delivery-club", "deliveryclub",
		"ресторан", "меню", "корзина", "заказ",
	}
	for _, indicator := range foodIndicators {
		if strings.Contains(contextLower, indicator) {
			contextScore = 0.3
			break
		}
	}

	return taskScore + contextScore, nil
}

// Execute выполняет задачу через базового агента
func (a *FoodDeliveryAgent) Execute(ctx context.Context, task string, maxSteps int) error {
	return a.baseAgent.executeTaskString(ctx, task, maxSteps)
}

// GetExpertise возвращает список экспертиз агента
func (a *FoodDeliveryAgent) GetExpertise() []string {
	return []string{
		"food_ordering",
		"restaurant_search",
		"cart_management",
		"checkout",
		"yandex_eda",
		"delivery_club",
	}
}

// GetType возвращает тип задачи
func (a *FoodDeliveryAgent) GetType() TaskType {
	return TaskTypeFoodDelivery
}

// GetDescription возвращает описание агента
func (a *FoodDeliveryAgent) GetDescription() string {
	return "Специализированный агент для оформления заказов на доставку еды: поиск ресторанов, добавление в корзину, оформление заказа"
}

