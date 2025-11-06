package database

import "time"

type Task struct {
	ID            uint      `gorm:"primaryKey"`
	UserInput     string    `gorm:"type:text;not null"`
	Status        string    `gorm:"type:varchar(32);not null;default:'pending'"`
	ResultSummary string    `gorm:"type:text"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

type AgentStep struct {
	ID             uint      `gorm:"primaryKey"`
	TaskID         uint      `gorm:"index;not null"`
	StepNo         int       `gorm:"not null"`
	ActionType     string    `gorm:"type:varchar(64);not null"`
	TargetSelector string    `gorm:"type:text"`
	Reasoning      string    `gorm:"type:text"`
	Result         string    `gorm:"type:text"`
	ScreenshotPath string    `gorm:"type:text"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}

type LlmLog struct {
	ID           uint   `gorm:"primaryKey"`
	TaskID       *uint  `gorm:"index"`
	StepID       *uint  `gorm:"index"`
	Role         string `gorm:"type:varchar(16);not null"`
	PromptText   string `gorm:"type:text;not null"`
	ResponseText string `gorm:"type:text"`
	Model        string `gorm:"type:varchar(64)"`
	TokensUsed   int
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}
