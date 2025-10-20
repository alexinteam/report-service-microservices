package events

import (
	"context"
	"fmt"
	"log"
	"time"
)

// IdempotentReportCreationSaga представляет идемпотентную Saga для создания отчета
type IdempotentReportCreationSaga struct {
	ID    string
	Steps []*SagaStep
}

// NewIdempotentReportCreationSaga создает новую идемпотентную Saga для создания отчета
func NewIdempotentReportCreationSaga(reportID, userID, templateID string, parameters map[string]interface{}) *IdempotentReportCreationSaga {
	return &IdempotentReportCreationSaga{
		ID: generateSagaID(),
		Steps: []*SagaStep{
			{
				ID:         "validate-user",
				Name:       "Validate User",
				Service:    "user-service",
				Action:     "validate_user",
				Compensate: "none", // Нет компенсации для валидации
				Data: map[string]interface{}{
					"user_id": userID,
				},
				Status: SagaStepPending,
			},
			{
				ID:         "validate-template",
				Name:       "Validate Template",
				Service:    "template-service",
				Action:     "validate_template",
				Compensate: "none", // Нет компенсации для валидации
				Data: map[string]interface{}{
					"template_id": templateID,
				},
				Status: SagaStepPending,
			},
			{
				ID:         "collect-data",
				Name:       "Collect Data",
				Service:    "data-service",
				Action:     "collect_data",
				Compensate: "none", // Данные можно пересобрать
				Data: map[string]interface{}{
					"template_id": templateID,
					"parameters":  parameters,
				},
				Status: SagaStepPending,
			},
			{
				ID:         "generate-report",
				Name:       "Generate Report",
				Service:    "report-service",
				Action:     "generate_report",
				Compensate: "delete_report",
				Data: map[string]interface{}{
					"report_id":   reportID,
					"template_id": templateID,
					"user_id":     userID,
					"parameters":  parameters,
				},
				Status: SagaStepPending,
			},
			{
				ID:         "store-file",
				Name:       "Store File",
				Service:    "storage-service",
				Action:     "store_file",
				Compensate: "delete_file",
				Data: map[string]interface{}{
					"report_id": reportID,
					"file_type": "report",
					"user_id":   userID,
				},
				Status: SagaStepPending,
			},
			{
				ID:         "send-notification",
				Name:       "Send Notification",
				Service:    "notification-service",
				Action:     "send_notification",
				Compensate: "none", // Уведомления не компенсируются
				Data: map[string]interface{}{
					"report_id": reportID,
					"user_id":   userID,
					"type":      "report_ready",
				},
				Status: SagaStepPending,
			},
			{
				ID:         "update-status",
				Name:       "Update Report Status",
				Service:    "report-service",
				Action:     "update_status",
				Compensate: "none", // Статус не компенсируется
				Data: map[string]interface{}{
					"user_id": userID,
					"status":  "completed",
				},
				Status: SagaStepPending,
			},
		},
	}
}

// Execute выполняет идемпотентную Saga
func (s *IdempotentReportCreationSaga) Execute(ctx context.Context, coordinator *IdempotentSagaCoordinator) error {
	log.Printf("Начинаем выполнение идемпотентной Saga создания отчета %s", s.ID)

	// Создаем объект Saga для передачи в coordinator
	saga := &Saga{
		ID:        s.ID,
		Name:      "Idempotent Report Creation Saga",
		Status:    SagaStatusPending,
		Steps:     s.Steps,
		Data:      make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Запускаем Saga через идемпотентный coordinator
	if err := coordinator.StartSaga(ctx, saga); err != nil {
		return fmt.Errorf("ошибка запуска Saga: %w", err)
	}

	// Выполняем шаги последовательно
	for i, step := range s.Steps {
		log.Printf("Выполняем шаг %d: %s", i+1, step.Name)

		// Получаем актуальное состояние саги перед выполнением шага
		saga, err := coordinator.GetSagaState(ctx, s.ID)
		if err != nil {
			log.Printf("Ошибка получения состояния Saga: %v", err)
			return fmt.Errorf("ошибка получения состояния Saga: %w", err)
		}

		// Находим актуальный шаг в состоянии саги
		var actualStep *SagaStep
		for _, s := range saga.Steps {
			if s.ID == step.ID {
				actualStep = s
				break
			}
		}
		if actualStep == nil {
			log.Printf("Шаг %s не найден в состоянии Saga", step.ID)
			return fmt.Errorf("шаг %s не найден в состоянии Saga", step.ID)
		}

		// Выполняем шаг через идемпотентный coordinator
		err = coordinator.ExecuteStep(ctx, s.ID, step.ID)
		if err != nil {
			log.Printf("Ошибка выполнения шага %s: %v", step.Name, err)

			// Обновляем статус Saga на Failed
			if updateErr := coordinator.UpdateSagaStatus(ctx, s.ID, SagaStatusFailed); updateErr != nil {
				log.Printf("Ошибка обновления статуса Saga: %v", updateErr)
			}

			// Компенсируем выполненные шаги
			return s.compensate(ctx, coordinator, i)
		}

		log.Printf("Шаг %s выполнен успешно", step.Name)
	}

	// Обновляем статус Saga на Completed
	if err := coordinator.UpdateSagaStatus(ctx, s.ID, SagaStatusCompleted); err != nil {
		log.Printf("Ошибка обновления статуса Saga на Completed: %v", err)
	}

	log.Printf("Идемпотентная Saga создания отчета %s выполнена успешно", s.ID)
	return nil
}

// compensate компенсирует выполненные шаги
func (s *IdempotentReportCreationSaga) compensate(ctx context.Context, coordinator *IdempotentSagaCoordinator, failedStepIndex int) error {
	log.Printf("Начинаем компенсацию идемпотентной Saga %s с шага %d", s.ID, failedStepIndex)

	// Компенсируем шаги в обратном порядке
	for i := failedStepIndex - 1; i >= 0; i-- {
		step := s.Steps[i]
		if step.Compensate == "none" {
			log.Printf("Шаг %s не требует компенсации", step.Name)
			continue
		}

		log.Printf("Компенсируем шаг: %s", step.Name)

		// Выполняем компенсацию через идемпотентный coordinator
		err := coordinator.CompensateStep(ctx, s.ID, step.ID)
		if err != nil {
			log.Printf("Ошибка компенсации шага %s: %v", step.Name, err)
			// Продолжаем компенсацию других шагов
		}
	}

	// Обновляем статус Saga на Compensated
	if err := coordinator.UpdateSagaStatus(ctx, s.ID, SagaStatusCompensated); err != nil {
		log.Printf("Ошибка обновления статуса Saga на Compensated: %v", err)
	}

	return fmt.Errorf("идемпотентная Saga %s выполнена с ошибками и компенсирована", s.ID)
}

// RetryFailedSaga повторяет выполнение неудачной Saga
func (s *IdempotentReportCreationSaga) RetryFailedSaga(ctx context.Context, coordinator *IdempotentSagaCoordinator) error {
	log.Printf("Повторное выполнение неудачной Saga %s", s.ID)

	// Получаем текущее состояние Saga
	saga, err := coordinator.GetSaga(ctx, s.ID)
	if err != nil {
		return fmt.Errorf("ошибка получения состояния Saga: %w", err)
	}

	// Проверяем, что Saga действительно неудачная
	if saga.Status != SagaStatusFailed {
		return fmt.Errorf("Saga %s не в статусе Failed, текущий статус: %s", s.ID, saga.Status)
	}

	// Сбрасываем статусы шагов для повторного выполнения
	for _, step := range s.Steps {
		if step.Status == SagaStepFailed {
			step.Status = SagaStepPending
			step.Error = ""
			step.ExecutedAt = nil
			step.CompletedAt = nil
		}
	}

	// Повторно выполняем Saga
	return s.Execute(ctx, coordinator)
}

// GetSagaProgress возвращает прогресс выполнения Saga
func (s *IdempotentReportCreationSaga) GetSagaProgress(ctx context.Context, coordinator *IdempotentSagaCoordinator) (*SagaProgress, error) {
	saga, err := coordinator.GetSaga(ctx, s.ID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения состояния Saga: %w", err)
	}

	completedSteps := 0
	failedSteps := 0
	compensatedSteps := 0

	for _, step := range saga.Steps {
		switch step.Status {
		case SagaStepCompleted:
			completedSteps++
		case SagaStepFailed:
			failedSteps++
		case SagaStepCompensated:
			compensatedSteps++
		}
	}

	return &SagaProgress{
		SagaID:           saga.ID,
		Status:           saga.Status,
		TotalSteps:       len(saga.Steps),
		CompletedSteps:   completedSteps,
		FailedSteps:      failedSteps,
		CompensatedSteps: compensatedSteps,
		ProgressPercent:  float64(completedSteps) / float64(len(saga.Steps)) * 100,
		CreatedAt:        saga.CreatedAt,
		UpdatedAt:        saga.UpdatedAt,
		CompletedAt:      saga.CompletedAt,
	}, nil
}

// SagaProgress представляет прогресс выполнения Saga
type SagaProgress struct {
	SagaID           string     `json:"saga_id"`
	Status           SagaStatus `json:"status"`
	TotalSteps       int        `json:"total_steps"`
	CompletedSteps   int        `json:"completed_steps"`
	FailedSteps      int        `json:"failed_steps"`
	CompensatedSteps int        `json:"compensated_steps"`
	ProgressPercent  float64    `json:"progress_percent"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}
