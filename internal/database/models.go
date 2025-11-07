// Package database предоставляет модели данных и репозиторий для работы с PostgreSQL.
// Использует GORM ORM с prepared statements для защиты от SQL injection.
package database

import "time"

// Task представляет задачу для выполнения агентом.
// Статусы: pending, running, completed, failed.
type Task struct {
	ID            uint      `gorm:"primaryKey"`
	UserInput     string    `gorm:"type:text;not null"`           // Текст задачи от пользователя
	Status        string    `gorm:"type:varchar(32);not null;default:'pending'"` // Статус выполнения
	ResultSummary string    `gorm:"type:text"`                    // Итоговый результат выполнения
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

// AgentStep представляет один шаг выполнения задачи.
// Содержит информацию о действии, селекторе, результате и скриншоте.
type AgentStep struct {
	ID             uint      `gorm:"primaryKey"`
	TaskID         uint      `gorm:"index;not null"`               // ID задачи
	StepNo         int       `gorm:"not null"`                     // Номер шага
	ActionType     string    `gorm:"type:varchar(64);not null"`    // Тип действия (navigate, click, type)
	TargetSelector string    `gorm:"type:text"`                    // CSS селектор элемента
	Reasoning      string    `gorm:"type:text"`                    // Обоснование действия от LLM
	Result         string    `gorm:"type:text"`                    // Результат выполнения шага
	ScreenshotPath string    `gorm:"type:text"`                    // Путь к скриншоту (если есть)
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}

// LlmLog представляет лог запроса к LLM.
// Сохраняет промпт, ответ, модель и количество использованных токенов.
type LlmLog struct {
	ID           uint   `gorm:"primaryKey"`
	TaskID       *uint  `gorm:"index"`                       // ID задачи (опционально)
	StepID       *uint  `gorm:"index"`                       // ID шага (опционально)
	Role         string `gorm:"type:varchar(16);not null"`   // Роль (user, assistant, system)
	PromptText   string `gorm:"type:text;not null"`          // Текст промпта
	ResponseText string `gorm:"type:text"`                   // Текст ответа
	Model        string `gorm:"type:varchar(64)"`            // Модель (gpt-4o)
	TokensUsed   int                                         // Количество токенов
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}
