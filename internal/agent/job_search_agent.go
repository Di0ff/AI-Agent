package agent

import (
	"context"
	"strings"
)

// JobSearchAgent - специализированный агент для поиска вакансий и отправки откликов
// Ключевые слова: "вакансия", "работа", "hh.ru", "резюме", "отклик", "сопроводительное"
type JobSearchAgent struct {
	baseAgent *Agent
	keywords  []string
}

// NewJobSearchAgent создает новый JobSearchAgent
func NewJobSearchAgent(baseAgent *Agent) *JobSearchAgent {
	return &JobSearchAgent{
		baseAgent: baseAgent,
		keywords: []string{
			// Русские ключевые слова
			"вакансия", "вакансии", "работа", "работу", "hh.ru", "хедхантер",
			"резюме", "отклик", "откликнуться", "сопроводительное", "письмо",
			"соискатель", "трудоустройство", "карьера",
			// Английские ключевые слова
			"job", "vacancy", "vacancies", "resume", "cv",
			"cover letter", "apply", "application",
		},
	}
}

// CanHandle проверяет, может ли JobSearchAgent обработать данную задачу
func (a *JobSearchAgent) CanHandle(ctx context.Context, task string, pageContext string) (float64, error) {
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

	// Проверяем контекст страницы (hh.ru)
	contextScore := 0.0
	jobIndicators := []string{
		"hh.ru", "headhunter", "вакансия", "резюме",
		"отклик", "сопроводительное", "job",
	}
	for _, indicator := range jobIndicators {
		if strings.Contains(contextLower, indicator) {
			contextScore = 0.3
			break
		}
	}

	return taskScore + contextScore, nil
}

// Execute выполняет задачу через базового агента
func (a *JobSearchAgent) Execute(ctx context.Context, task string, maxSteps int) error {
	return a.baseAgent.executeTaskString(ctx, task, maxSteps)
}

// GetExpertise возвращает список экспертиз агента
func (a *JobSearchAgent) GetExpertise() []string {
	return []string{
		"job_search",
		"vacancy_search",
		"resume_extraction",
		"cover_letter_generation",
		"job_application",
		"hh_ru",
	}
}

// GetType возвращает тип задачи
func (a *JobSearchAgent) GetType() TaskType {
	return TaskTypeJobSearch
}

// GetDescription возвращает описание агента
func (a *JobSearchAgent) GetDescription() string {
	return "Специализированный агент для поиска вакансий и отправки откликов: поиск вакансий, изучение резюме, генерация сопроводительных писем"
}

