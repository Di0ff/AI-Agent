package ui

import "fmt"

// FormatStatus возвращает иконку, цвет и текст для статуса задачи
func FormatStatus(status string) (icon, color, text string) {
	switch status {
	case "completed":
		return IconCheckmark, ColorGreen, "завершена"
	case "failed":
		return IconCross, ColorRed, "ошибка"
	case "running":
		return IconPlay, ColorCyan, "выполняется"
	case "pending":
		return IconClock, ColorYellow, "ожидает"
	default:
		return IconClock, ColorYellow, status
	}
}

// ClearScreen очищает терминал
func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}
