package agent

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"aiAgent/internal/browser"
	"aiAgent/internal/llm"

	"go.uber.org/zap"
)

func (a *Agent) ExecuteTaskMultiStep(ctx context.Context, taskText string, maxSteps int) error {
	var pageSnapshot *browser.PageSnapshot
	err := retryAction(ctx, a.retries, a.retryDelay, func() error {
		snapshot, err := a.browser.GetPageSnapshot(ctx)
		if err != nil {
			return err
		}
		pageSnapshot = snapshot
		return nil
	})
	if err != nil {
		a.log.Error("Ошибка получения начального snapshot", a.contextFields(nil, 0, zap.Error(err))...)
		return fmt.Errorf("failed to get initial snapshot: %w", err)
	}

	var pageContext string
	if pageSnapshot != nil {
		pageContext = a.limitContextFromSnapshot(pageSnapshot)
	}

	domain := extractDomain(pageSnapshot.URL)

	if a.memory != nil {
		existingPath := a.memory.FindSimilarSuccessfulPath(ctx, taskText, domain)
		if existingPath != nil {
			a.log.Info("Найден успешный путь в памяти",
				a.contextFields(nil, 0,
					zap.String("strategy", existingPath.Strategy),
					zap.Int("success_count", existingPath.SuccessCount))...)

			fmt.Printf("\n[Memory] Найден проверенный путь (использован %d раз)\n", existingPath.SuccessCount)
			fmt.Printf("[Стратегия] %s\n\n", existingPath.Strategy)

			plan := &llm.MultiStepPlan{
				Steps:            existingPath.Steps,
				OverallStrategy:  existingPath.Strategy,
				FallbackStrategy: "Replan if step fails",
				EstimatedSteps:   len(existingPath.Steps),
			}

			return a.executeMultiStepPlanWithMemory(ctx, taskText, plan, maxSteps, domain)
		}
	}

	plan, err := a.llmClient.PlanMultiStep(ctx, taskText, pageContext, maxSteps, nil, nil)
	if err != nil {
		a.log.Error("Ошибка планирования multi-step", a.contextFields(nil, 0, zap.Error(err))...)
		return fmt.Errorf("failed to plan multi-step: %w", err)
	}

	a.log.Info("Multi-step план создан",
		a.contextFields(nil, 0,
			zap.Int("steps", len(plan.Steps)),
			zap.String("strategy", plan.OverallStrategy))...)

	fmt.Printf("\n[Стратегия] %s\n", plan.OverallStrategy)
	fmt.Printf("[Запланировано шагов] %d\n\n", len(plan.Steps))

	return a.executeMultiStepPlanWithMemory(ctx, taskText, plan, maxSteps, domain)
}

func (a *Agent) executeMultiStepPlanWithMemory(ctx context.Context, taskText string, plan *llm.MultiStepPlan, maxSteps int, domain string) error {
	startTime := time.Now()

	err := a.executeMultiStepPlan(ctx, taskText, plan, maxSteps, domain)

	if err == nil && a.memory != nil {
		duration := time.Since(startTime)
		if saveErr := a.memory.RecordSuccess(ctx, taskText, plan.Steps, plan.OverallStrategy, duration, domain); saveErr != nil {
			a.log.Warn("Не удалось сохранить успешный путь", a.contextFields(nil, 0, zap.Error(saveErr))...)
		} else {
			a.log.Info("Успешный путь сохранен в память", a.contextFields(nil, 0, zap.Duration("duration", duration))...)
		}
	}

	return err
}

func (a *Agent) executeMultiStepPlan(ctx context.Context, taskText string, plan *llm.MultiStepPlan, maxSteps int, domain string) error {
	for stepNo, step := range plan.Steps {
		if stepNo >= maxSteps {
			break
		}

		stepNumber := stepNo + 1
		fmt.Printf("[Шаг %d/%d] %s: %s\n", stepNumber, len(plan.Steps), step.Action, step.Reasoning)

		if step.Action == "complete" {
			a.log.Info("Задача завершена согласно плану", a.contextFields(nil, stepNumber)...)
			return nil
		}

		pageSnapshot, err := a.browser.GetPageSnapshot(ctx)
		if err != nil {
			a.log.Warn("Не удалось получить snapshot перед шагом", a.contextFields(nil, stepNumber, zap.Error(err))...)
		}

		isDangerous, llmMessage, err := a.securityChecker.IsDangerousAction(ctx, step.Action, step.Selector, step.Value, step.Reasoning)
		if err != nil {
			a.log.Warn("Ошибка проверки безопасности", a.contextFields(nil, stepNumber, zap.Error(err))...)
		} else if isDangerous {
			if a.userInputProvider == nil {
				a.log.Warn("Опасное действие обнаружено, но провайдер не настроен", a.contextFields(nil, stepNumber, zap.String("action", step.Action))...)
			} else {
				confirmationMsg := a.securityChecker.GetConfirmationMessage(step.Action, step.Selector, step.Value, step.Reasoning, llmMessage)
				answer, err := a.userInputProvider.AskUser(ctx, confirmationMsg)
				if err != nil {
					return fmt.Errorf("ошибка запроса подтверждения: %w", err)
				}

				answerLower := strings.ToLower(strings.TrimSpace(answer))
				if answerLower != "yes" && answerLower != "y" && answerLower != "да" && answerLower != "д" {
					a.log.Info("Пользователь отменил опасное действие", a.contextFields(nil, stepNumber, zap.String("action", step.Action))...)
					fmt.Printf("[Шаг %d] Действие отменено пользователем\n", stepNumber)
					continue
				}
			}
		}

		_, err = a.executeActionWithRetry(ctx, &step)
		if err != nil {
			a.log.Error("Ошибка выполнения шага", a.contextFields(nil, stepNumber, zap.String("action", step.Action), zap.Error(err))...)

			if isCriticalError(err) {
				a.log.Error("Критическая ошибка, требуется replan", a.contextFields(nil, stepNumber, zap.Error(err))...)

				var currentContext string
				if pageSnapshot != nil {
					currentContext = a.limitContextFromSnapshot(pageSnapshot)
				}

				newPlan, replanErr := a.llmClient.Replan(ctx, taskText, currentContext, plan, &step, err.Error(), maxSteps-stepNumber, nil, nil)
				if replanErr != nil {
					a.log.Error("Не удалось создать новый план", a.contextFields(nil, stepNumber, zap.Error(replanErr))...)
					return fmt.Errorf("failed to replan after error: %w", replanErr)
				}

				a.log.Info("Новый план создан после ошибки", a.contextFields(nil, stepNumber, zap.Int("new_steps", len(newPlan.Steps)))...)
				fmt.Printf("\n[Replan] Новая стратегия: %s\n", newPlan.OverallStrategy)
				fmt.Printf("[Новых шагов] %d\n\n", len(newPlan.Steps))

				return a.executeMultiStepPlan(ctx, taskText, newPlan, maxSteps-stepNumber, domain)
			}

			a.log.Warn("Некритическая ошибка, продолжаем выполнение", a.contextFields(nil, stepNumber, zap.Error(err))...)
			fmt.Printf("[Шаг %d] Ошибка: %v (продолжаем)\n", stepNumber, err)

			if a.memory != nil {
				recovery := a.memory.GetFailureRecovery(ctx, step.Action, step.Selector, err.Error())
				if recovery != "" {
					a.log.Info("Найдена стратегия восстановления в памяти", a.contextFields(nil, stepNumber, zap.String("recovery", recovery))...)
					fmt.Printf("[Memory] Применяем известную стратегию восстановления: %s\n", recovery)
				}
				a.memory.RecordFailure(ctx, step.Action, step.Selector, err.Error(), "")
			}
		}
	}

	a.log.Info("Все шаги выполнены", a.contextFields(nil, 0)...)
	return nil
}

func extractDomain(urlStr string) string {
	if urlStr == "" {
		return ""
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	return parsedURL.Host
}
