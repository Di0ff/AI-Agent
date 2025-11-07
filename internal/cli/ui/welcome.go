package ui

import (
	"fmt"
	"os"
)

// PrintWelcome выводит приветствие и лого
func PrintWelcome() {
	logoBytes, err := os.ReadFile("logo.txt")
	if err == nil {
		fmt.Println(ColorCyan + string(logoBytes) + ColorReset)
	}
	fmt.Println(ColorBold + IconRobot + " AI-Agent v0.1.0" + ColorReset)
	fmt.Println(ColorGray + "Автономный AI-агент для управления браузером" + ColorReset)
	fmt.Println(ColorGray + "Используется: Firefox + OpenAI GPT-4o" + ColorReset)
	fmt.Println()
	PrintHelp()
	fmt.Println(ColorCyan + IconBulb + " Совет:" + ColorReset + " Используйте " + ColorYellow + "open-persistent" + ColorReset + " для входа на сайты, затем " + ColorYellow + "run" + ColorReset + " для выполнения задач")
	fmt.Println()
	fmt.Println(ColorGray + "⬆️ ⬇️" + ColorReset + " Используйте стрелки для навигации по истории команд")
	fmt.Println()
}

// PrintHelp выводит список доступных команд
func PrintHelp() {
	fmt.Println(ColorYellow + IconList + " Доступные команды:" + ColorReset)
	fmt.Println("  " + ColorGreen + "task" + ColorReset + " <текст>        - Создать новую задачу")
	fmt.Println("  " + ColorGreen + "tasks" + ColorReset + "               - Список всех задач")
	fmt.Println("  " + ColorGreen + "run" + ColorReset + " <id>            - Выполнить задачу")
	fmt.Println("  " + ColorGreen + "status" + ColorReset + " <id>         - Статус задачи")
	fmt.Println("  " + ColorGreen + "show" + ColorReset + " <id>           - Детали задачи")
	fmt.Println("  " + ColorGreen + "logs" + ColorReset + " <id>           - LLM логи задачи")
	fmt.Println("  " + ColorGreen + "test-llm" + ColorReset + " <задача>   - Тест планирования LLM")
	fmt.Println("  " + ColorGreen + "open" + ColorReset + " <url>          - Открыть URL в браузере")
	fmt.Println("  " + ColorGreen + "open-persistent" + ColorReset + "     - Открыть браузер для ручной настройки")
	fmt.Println("  " + ColorGreen + "clear" + ColorReset + "               - Очистить экран")
	fmt.Println("  " + ColorGreen + "exit" + ColorReset + "                - Выход")
	fmt.Println()
}
