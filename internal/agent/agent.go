package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"aiAgent/internal/browser"
	"aiAgent/internal/database"
	"aiAgent/internal/llm"
	"aiAgent/internal/logger"
	"aiAgent/internal/sanitizer"

	"go.uber.org/zap"
)

func getSanitizer(llmClient llm.LLMClient) *sanitizer.DataSanitizer {
	if llmClient != nil {
		if client, ok := llmClient.(*llm.Client); ok {
			return sanitizer.NewWithAI(client)
		}
	}
	return sanitizer.New()
}

// New создает новый экземпляр агента с заданной конфигурацией.
// Агент инициализируется с браузером, LLM клиентом, репозиторием для хранения задач и логгером.
// Если в конфигурации не указаны дефолтные значения, они устанавливаются автоматически:
//   - MaxSteps: 50
//   - MaxTokens: 2000
//   - Retries: 3
//   - RetryDelay: 2 секунды
//   - ConfidenceMin: 0.7
//
// При использовании подагентов (UseSubAgents=true) инициализируются специализированные агенты
// для навигации, работы с формами, извлечения данных и взаимодействия с элементами.
func New(br browser.Browser, llmClient llm.LLMClient, repo *database.TaskRepository, log *logger.Zap, cfg Config) *Agent {
	if cfg.MaxSteps == 0 {
		cfg.MaxSteps = 50
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 2000
	}
	if cfg.Retries == 0 {
		cfg.Retries = 3
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 2 * time.Second
	}
	if cfg.ConfidenceMin == 0 {
		cfg.ConfidenceMin = 0.7
	}

	agent := &Agent{
		browser:           br,
		llmClient:         llmClient,
		repo:              repo,
		log:               log,
		maxSteps:          cfg.MaxSteps,
		maxTokens:         cfg.MaxTokens,
		retries:           cfg.Retries,
		retryDelay:        cfg.RetryDelay,
		userInputProvider: cfg.UserInputProvider,
		securityChecker:   NewSecurityChecker(llmClient),
		sanitizer:         getSanitizer(llmClient),
		cfg:               cfg,
	}

	if cfg.UseSubAgents {
		router := NewAgentRouter(llmClient, cfg.ConfidenceMin)

		// Регистрируем специализированных агентов для трех задач
		emailAgent := NewEmailSpamAgent(agent)
		foodAgent := NewFoodDeliveryAgent(agent)
		jobAgent := NewJobSearchAgent(agent)

		router.RegisterAgent(emailAgent)
		router.RegisterAgent(foodAgent)
		router.RegisterAgent(jobAgent)

		// Устанавливаем EmailSpamAgent как дефолтный (можно изменить на любой другой)
		router.SetDefaultAgent(emailAgent)

		agent.router = router
	}

	if cfg.UseMemory {
		memory := NewAgentMemory(repo)
		agent.memory = memory
	}

	agent.circuitBreakers = NewCircuitBreakerPool()

	return agent
}

// contextFields создаёт набор контекстных полей для логирования
func (a *Agent) contextFields(taskID *uint, stepNo int, fields ...zap.Field) []zap.Field {
	result := make([]zap.Field, 0, len(fields)+2)
	if taskID != nil {
		result = append(result, zap.Uint("task_id", *taskID))
	}
	if stepNo > 0 {
		result = append(result, zap.Int("step", stepNo))
	}
	result = append(result, fields...)
	return result
}

func (a *Agent) getPageContext(ctx context.Context) (string, error) {
	snapshot, err := a.browser.GetPageSnapshot(ctx)
	if err == nil && snapshot != nil {
		return a.limitContextFromSnapshot(snapshot), nil
	}

	htmlContext, err := a.browser.GetPageContext(ctx)
	if err != nil {
		return "", err
	}
	return a.limitContext(htmlContext), nil
}

func (a *Agent) performReasoning(ctx context.Context, userInput, pageContext string, taskID *uint, stepNo int) (*llm.ReasoningStep, error) {
	// Проверяем есть ли memory и релевантные patterns
	if a.memory != nil && a.cfg.UseMemory {
		// Пытаемся найти релевантный опыт из памяти
		// TODO: В будущем добавить поиск relevant reasoning patterns из memory
		// Пока используем обычный Reason
	}

	// Выполняем reasoning с retry logic
	var reasoning *llm.ReasoningStep
	err := retryAction(ctx, a.retries, a.retryDelay, func() error {
		r, e := a.llmClient.Reason(ctx, userInput, pageContext, a.reasoningHistory, taskID, nil)
		if e != nil {
			return e
		}
		reasoning = r
		return nil
	})

	return reasoning, err
}

// performBasicReflection выполняет базовую рефлексию после действия.
// Это упрощенная версия - полная рефлексия с LLM анализом будет добавлена в Фазе 4.
func (a *Agent) performBasicReflection(stepNo int, plan *llm.StepPlan, result string, execErr error) {
	// Базовая рефлексия
	if execErr != nil {
		// TODO (Фаза 4): Добавить LLM-based анализ ошибки
		// TODO (Фаза 4): Обновить reasoning history с результатом
		// TODO (Фаза 4): Сохранить в Memory как failed pattern
	} else {
		// TODO (Фаза 4): Добавить LLM-based оценку успешности
		// TODO (Фаза 4): Проверить достигнута ли цель шага
		// TODO (Фаза 4): Сохранить в Memory как successful pattern
	}
}

func (a *Agent) getPlanForStep(ctx context.Context, userInput, pageContext string, taskID *uint) (*llm.StepPlan, error) {
	var plan *llm.StepPlan
	err := retryAction(ctx, a.retries, a.retryDelay, func() error {
		// Получаем последний reasoning step для передачи в планирование
		var latestReasoning *llm.ReasoningStep
		if a.reasoningHistory != nil {
			latestReasoning = a.reasoningHistory.GetLastStep()
		}

		// Используем новый метод PlanActionWithReasoning для ReAct pattern
		p, e := a.llmClient.PlanActionWithReasoning(ctx, userInput, pageContext, latestReasoning, taskID, nil)
		if e != nil {
			return e
		}
		plan = p
		return nil
	})
	return plan, err
}

func (a *Agent) checkSecurityAndConfirm(ctx context.Context, plan *llm.StepPlan, stepNo int) (bool, error) {
	isDangerous, llmMessage, err := a.securityChecker.IsDangerousAction(ctx, plan.Action, plan.Selector, plan.Value, plan.Reasoning)
	if err != nil {
		a.log.Warn("Ошибка проверки безопасности, продолжаем выполнение", a.contextFields(nil, stepNo, zap.Error(err))...)
		return true, nil
	}

	if !isDangerous {
		return true, nil
	}

	if a.userInputProvider == nil {
		a.log.Warn("Опасное действие обнаружено, но провайдер пользовательского ввода не настроен", a.contextFields(nil, stepNo, zap.String("action", plan.Action))...)
		return true, nil
	}

	confirmationMsg := a.securityChecker.GetConfirmationMessage(plan.Action, plan.Selector, plan.Value, plan.Reasoning, llmMessage)
	answer, err := a.userInputProvider.AskUser(ctx, confirmationMsg)
	if err != nil {
		return false, fmt.Errorf("ошибка запроса подтверждения: %w", err)
	}

	answerLower := strings.ToLower(strings.TrimSpace(answer))
	if answerLower != "yes" && answerLower != "y" && answerLower != "да" && answerLower != "д" {
		a.log.Info("Пользователь отменил опасное действие", a.contextFields(nil, stepNo, zap.String("action", plan.Action))...)
		fmt.Printf("[Шаг %d] Действие отменено пользователем\n", stepNo)
		return false, nil
	}

	a.log.Info("Пользователь подтвердил опасное действие", a.contextFields(nil, stepNo, zap.String("action", plan.Action))...)
	return true, nil
}

func (a *Agent) createStepRecord(task *database.Task, stepNo int, plan *llm.StepPlan, result string) *database.AgentStep {
	return &database.AgentStep{
		TaskID:         task.ID,
		StepNo:         stepNo,
		ActionType:     plan.Action,
		TargetSelector: a.sanitizer.SanitizeSelector(plan.Selector),
		Reasoning:      a.sanitizer.Sanitize(plan.Reasoning),
		Result:         a.sanitizer.Sanitize(result),
	}
}

type executeStepsParams struct {
	ctx        context.Context
	userInput  string
	maxSteps   int
	taskID     *uint
	saveSteps  bool
	updateTask bool
}

func (a *Agent) executeSteps(params executeStepsParams) error {
	task := &database.Task{}
	if params.taskID != nil {
		t, err := a.repo.GetTaskByID(*params.taskID)
		if err != nil {
			return fmt.Errorf("ошибка получения задачи: %w", err)
		}
		task = t
	}

	// Инициализация reasoning history для этой задачи (ReAct pattern)
	a.reasoningHistory = &llm.ReasoningHistory{}

	for stepNo := 1; stepNo <= params.maxSteps; stepNo++ {
		// Проверка отмены контекста
		select {
		case <-params.ctx.Done():
			a.log.Info("Выполнение задачи отменено через контекст", a.contextFields(params.taskID, stepNo)...)
			return params.ctx.Err()
		default:
		}

		pageContext, err := a.getPageContext(params.ctx)
		if err != nil {
			a.log.Warn("Ошибка получения контекста страницы", a.contextFields(params.taskID, stepNo, zap.Error(err))...)
			pageContext = ""
		}

		// ========================================
		// ФАЗА 1: REASONING (новое!)
		// Явное рассуждение перед планированием действия
		// ========================================
		reasoning, err := a.performReasoning(params.ctx, params.userInput, pageContext, params.taskID, stepNo)
		if err != nil {
			a.log.Warn("Ошибка reasoning, продолжаем с планированием", a.contextFields(params.taskID, stepNo, zap.Error(err))...)
			// Не критично - можем продолжить без explicit reasoning
		} else {
			// Добавляем reasoning в историю
			a.reasoningHistory.AddStep(*reasoning)

			// Если reasoning говорит что нужен user input - обработаем это
			if reasoning.RequiresUserInput && a.userInputProvider != nil {
				// В будущем можно добавить автоматический ask_user action здесь
			}
		}

		// ========================================
		// ФАЗА 2: PLANNING
		// Планирование действия (теперь с учетом reasoning)
		// ========================================
		plan, err := a.getPlanForStep(params.ctx, params.userInput, pageContext, params.taskID)
		if err != nil {
			a.log.Error("Ошибка планирования действия", a.contextFields(params.taskID, stepNo, zap.Error(err))...)
			if isCriticalError(err) {
				return fmt.Errorf("критичная ошибка планирования: %w", err)
			}
			continue
		}

		if plan.Action == "complete" {
			if params.saveSteps {
				step := a.createStepRecord(task, stepNo, plan, plan.Reasoning)
				if err := a.repo.CreateStep(step); err != nil {
					a.log.Error("Ошибка сохранения шага", a.contextFields(params.taskID, stepNo, zap.Error(err))...)
				}
			}
			if params.updateTask && params.taskID != nil {
				if err := a.repo.UpdateTaskStatus(*params.taskID, "completed", a.sanitizer.Sanitize(plan.Reasoning)); err != nil {
					a.log.Error("Ошибка обновления статуса", a.contextFields(params.taskID, stepNo, zap.Error(err))...)
				}
			}
			return nil
		}

		approved, err := a.checkSecurityAndConfirm(params.ctx, plan, stepNo)
		if err != nil {
			if params.saveSteps {
				step := a.createStepRecord(task, stepNo, plan, "Ошибка запроса подтверждения")
				if err := a.repo.CreateStep(step); err != nil {
					a.log.Error("Ошибка сохранения шага", a.contextFields(params.taskID, stepNo, zap.Error(err))...)
				}
			}
			return err
		}
		if !approved {
			if params.saveSteps {
				step := a.createStepRecord(task, stepNo, plan, "Действие отменено пользователем")
				if err := a.repo.CreateStep(step); err != nil {
					a.log.Error("Ошибка сохранения шага", a.contextFields(params.taskID, stepNo, zap.Error(err))...)
				}
			}
			continue
		}

		result, err := a.executeActionWithRetry(params.ctx, plan)
		if err != nil {
			actionErr := classifyError(plan.Action, err)
			errorMsg := fmt.Sprintf("Ошибка: %v", err)
			a.log.Error("Ошибка выполнения действия", a.contextFields(params.taskID, stepNo,
				zap.String("action", plan.Action),
				zap.Error(err),
				zap.String("error_type", actionErr.Type.String()))...)

			if params.saveSteps {
				step := a.createStepRecord(task, stepNo, plan, errorMsg)
				if err := a.repo.CreateStep(step); err != nil {
					a.log.Error("Ошибка сохранения шага", a.contextFields(params.taskID, stepNo, zap.Error(err))...)
				}
			}

			if isCriticalError(err) {
				return fmt.Errorf("критичная ошибка выполнения действия: %w", err)
			}

			a.log.Warn("Продолжаем работу после некритичной ошибки", a.contextFields(params.taskID, stepNo,
				zap.String("action", plan.Action),
				zap.Error(err))...)
		} else {
			if params.saveSteps {
				step := a.createStepRecord(task, stepNo, plan, result)
				if err := a.repo.CreateStep(step); err != nil {
					a.log.Error("Ошибка сохранения шага", a.contextFields(params.taskID, stepNo, zap.Error(err))...)
				}
			}

			// ========================================
			// ФАЗА 3: REFLECTION (базовая версия)
			// Рефлексия после успешного выполнения действия
			// ========================================
			a.performBasicReflection(stepNo, plan, result, err)
		}

		a.logStep(stepNo, plan, err)
	}

	return fmt.Errorf("достигнут лимит шагов (%d)", params.maxSteps)
}

func (a *Agent) ExecuteTask(ctx context.Context, task *database.Task) error {
	if err := a.repo.UpdateTaskStatus(task.ID, "running", ""); err != nil {
		return fmt.Errorf("ошибка обновления статуса задачи: %w", err)
	}

	if err := a.browser.Launch(ctx); err != nil {
		return fmt.Errorf("ошибка запуска браузера: %w", err)
	}
	defer a.browser.Close()

	multiStepSize := 5
	if a.cfg.MultiStepSize > 0 {
		multiStepSize = a.cfg.MultiStepSize
	}

	if a.cfg.UseMultiStep {
		a.log.Info("Использование multi-step планирования", a.contextFields(&task.ID, 0, zap.Int("max_steps", multiStepSize))...)
		return a.ExecuteTaskMultiStep(ctx, task.UserInput, multiStepSize)
	}

	if a.router != nil {
		pageContext, _ := a.getPageContext(ctx)
		selectedAgent, err := a.router.RouteTask(ctx, task.UserInput, pageContext)
		if err != nil {
			a.log.Warn("Ошибка маршрутизации задачи, используем основной агент", a.contextFields(&task.ID, 0, zap.Error(err))...)
		} else {
			a.log.Info("Задача маршрутизирована", a.contextFields(&task.ID, 0, zap.String("agent_type", string(selectedAgent.GetType())))...)
			// Делегируем выполнение специализированному агенту
			return selectedAgent.Execute(ctx, task.UserInput, a.maxSteps)
		}
	}

	return a.executeSteps(executeStepsParams{
		ctx:        ctx,
		userInput:  task.UserInput,
		maxSteps:   a.maxSteps,
		taskID:     &task.ID,
		saveSteps:  true,
		updateTask: true,
	})
}

func (a *Agent) executeTaskString(ctx context.Context, taskText string, maxSteps int) error {
	return a.executeSteps(executeStepsParams{
		ctx:        ctx,
		userInput:  taskText,
		maxSteps:   maxSteps,
		taskID:     nil,
		saveSteps:  false,
		updateTask: false,
	})
}

func (a *Agent) logStep(stepNo int, plan *llm.StepPlan, err error) {
	if err != nil {
		fmt.Printf("[Шаг %d] %s - ОШИБКА\n", stepNo, plan.Action)
	} else {
		reasoning := plan.Reasoning
		if len(reasoning) > 60 {
			reasoning = reasoning[:57] + "..."
		}
		fmt.Printf("[Шаг %d] %s: %s\n", stepNo, plan.Action, reasoning)
	}
}

func (a *Agent) executeActionWithRetry(ctx context.Context, plan *llm.StepPlan) (string, error) {
	var result string

	err := retryAction(ctx, a.retries, a.retryDelay, func() error {
		res, e := a.executeAction(ctx, plan)
		if e != nil {
			actionErr := classifyError(plan.Action, e)
			if actionErr.Type == ErrorTypeCritical {
				return e
			}
			return e
		}
		result = res
		return nil
	})

	if err != nil {
		return "", err
	}

	return result, nil
}

func (a *Agent) executeAction(ctx context.Context, plan *llm.StepPlan) (string, error) {
	switch plan.Action {
	case "navigate":
		if err := a.browser.Navigate(ctx, plan.Value); err != nil {
			return "", fmt.Errorf("навигация: %w", err)
		}
		return fmt.Sprintf("Переход на %s", plan.Value), nil

	case "click":
		if err := a.browser.Click(ctx, plan.Selector); err != nil {
			return "", fmt.Errorf("клик: %w", err)
		}
		return fmt.Sprintf("Клик по %s", plan.Selector), nil

	case "type":
		if err := a.browser.Type(ctx, plan.Selector, plan.Value); err != nil {
			return "", fmt.Errorf("ввод: %w", err)
		}
		return fmt.Sprintf("Ввод '%s' в %s", plan.Value, plan.Selector), nil

	case "extract_info":
		context, err := a.browser.GetPageContext(ctx)
		if err != nil {
			return "", fmt.Errorf("извлечение: %w", err)
		}
		return fmt.Sprintf("Извлечено: %s", a.limitContext(context)), nil

	case "ask_user":
		if a.userInputProvider == nil {
			return "", fmt.Errorf("провайдер пользовательского ввода не настроен")
		}
		answer, err := a.userInputProvider.AskUser(ctx, plan.Value)
		if err != nil {
			return "", fmt.Errorf("ошибка запроса пользователя: %w", err)
		}
		return fmt.Sprintf("Ответ пользователя: %s", answer), nil

	default:
		return "", fmt.Errorf("неизвестное действие: %s", plan.Action)
	}
}
