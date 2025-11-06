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
	return &Agent{
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
	}
}

func (a *Agent) ExecuteTask(ctx context.Context, task *database.Task) error {
	if err := a.repo.UpdateTaskStatus(task.ID, "running", ""); err != nil {
		return fmt.Errorf("ошибка обновления статуса задачи: %w", err)
	}

	if err := a.browser.Launch(ctx); err != nil {
		return fmt.Errorf("ошибка запуска браузера: %w", err)
	}
	defer a.browser.Close()

	stepNo := 0
	for stepNo < a.maxSteps {
		stepNo++

		var pageContext string
		err := retryAction(ctx, a.retries, a.retryDelay, func() error {
			ctx, err := a.browser.GetPageContext(ctx)
			if err != nil {
				return err
			}
			pageContext = ctx
			return nil
		})
		if err != nil {
			a.log.Error("Ошибка получения контекста страницы после повторных попыток", zap.Error(err))
			if isCriticalError(err) {
				return fmt.Errorf("критичная ошибка получения контекста: %w", err)
			}
			a.log.Warn("Продолжаем работу после некритичной ошибки получения контекста", zap.Error(err))
			pageContext = ""
		}

		pageContext = a.limitContext(pageContext)

		stepID := uint(stepNo)
		var plan *llm.StepPlan
		err = retryAction(ctx, a.retries, a.retryDelay, func() error {
			p, e := a.llmClient.PlanAction(ctx, task.UserInput, pageContext, &task.ID, &stepID)
			if e != nil {
				return e
			}
			plan = p
			return nil
		})
		if err != nil {
			a.log.Error("Ошибка планирования действия после повторных попыток", zap.Error(err))
			if isCriticalError(err) {
				return fmt.Errorf("критичная ошибка планирования: %w", err)
			}
			a.log.Warn("Пропускаем шаг из-за ошибки планирования", zap.Error(err))
			continue
		}

		step := &database.AgentStep{
			TaskID:         task.ID,
			StepNo:         stepNo,
			ActionType:     plan.Action,
			TargetSelector: a.sanitizer.SanitizeSelector(plan.Selector),
			Reasoning:      a.sanitizer.Sanitize(plan.Reasoning),
		}

		if plan.Action == "complete" {
			step.Result = a.sanitizer.Sanitize(plan.Reasoning)
			if err := a.repo.CreateStep(step); err != nil {
				a.log.Error("Ошибка сохранения шага", zap.Error(err))
			}
			if err := a.repo.UpdateTaskStatus(task.ID, "completed", a.sanitizer.Sanitize(plan.Reasoning)); err != nil {
				a.log.Error("Ошибка обновления статуса", zap.Error(err))
			}
			return nil
		}

		isDangerous, llmMessage, err := a.securityChecker.IsDangerousAction(ctx, plan.Action, plan.Selector, plan.Value, plan.Reasoning)
		if err != nil {
			a.log.Warn("Ошибка проверки безопасности, продолжаем выполнение", zap.Error(err))
		} else if isDangerous {
			if a.userInputProvider == nil {
				a.log.Warn("Опасное действие обнаружено, но провайдер пользовательского ввода не настроен", zap.String("action", plan.Action))
			} else {
				confirmationMsg := a.securityChecker.GetConfirmationMessage(plan.Action, plan.Selector, plan.Value, plan.Reasoning, llmMessage)
				answer, err := a.userInputProvider.AskUser(ctx, confirmationMsg)
				if err != nil {
					step.Result = "Ошибка запроса подтверждения"
					if err := a.repo.CreateStep(step); err != nil {
						a.log.Error("Ошибка сохранения шага", zap.Error(err))
					}
					return fmt.Errorf("ошибка запроса подтверждения: %w", err)
				}

				answerLower := strings.ToLower(strings.TrimSpace(answer))
				if answerLower != "yes" && answerLower != "y" && answerLower != "да" && answerLower != "д" {
					step.Result = "Действие отменено пользователем"
					if err := a.repo.CreateStep(step); err != nil {
						a.log.Error("Ошибка сохранения шага", zap.Error(err))
					}
					a.log.Info("Пользователь отменил опасное действие", zap.String("action", plan.Action))
					fmt.Printf("[Шаг %d] Действие отменено пользователем\n", stepNo)
					continue
				}

				a.log.Info("Пользователь подтвердил опасное действие", zap.String("action", plan.Action))
			}
		}

		result, err := a.executeActionWithRetry(ctx, plan)
		if err != nil {
			actionErr := classifyError(plan.Action, err)
			step.Result = a.sanitizer.Sanitize(fmt.Sprintf("Ошибка: %v", err))
			a.log.Error("Ошибка выполнения действия", zap.String("action", plan.Action), zap.Error(err), zap.String("error_type", actionErr.Type.String()))

			if isCriticalError(err) {
				if err := a.repo.CreateStep(step); err != nil {
					a.log.Error("Ошибка сохранения шага", zap.Error(err))
				}
				return fmt.Errorf("критичная ошибка выполнения действия: %w", err)
			}

			a.log.Warn("Продолжаем работу после некритичной ошибки", zap.String("action", plan.Action), zap.Error(err))
		} else {
			step.Result = a.sanitizer.Sanitize(result)
		}

		if err := a.repo.CreateStep(step); err != nil {
			a.log.Error("Ошибка сохранения шага", zap.Error(err))
		}

		fmt.Printf("[Шаг %d] %s: %s\n", stepNo, plan.Action, plan.Reasoning)
		if err != nil {
			fmt.Printf("  Ошибка: %v\n", err)
		}
	}

	return fmt.Errorf("достигнут лимит шагов (%d)", a.maxSteps)
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
