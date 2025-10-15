package events

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ReportCreationSaga представляет Saga для создания отчета
type ReportCreationSaga struct {
	ID    string
	Steps []*SagaStep
}

// NewReportCreationSaga создает новую Saga для создания отчета
func NewReportCreationSaga(userID, templateID string, parameters map[string]interface{}) *ReportCreationSaga {
	return &ReportCreationSaga{
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
					"user_id": userID,
					"type":    "report_ready",
				},
				Status: SagaStepPending,
			},
		},
	}
}

// Execute выполняет Saga
func (s *ReportCreationSaga) Execute(ctx context.Context, coordinator SagaManager) error {
	log.Printf("Начинаем выполнение Saga создания отчета %s", s.ID)

	for i, step := range s.Steps {
		log.Printf("Выполняем шаг %d: %s", i+1, step.Name)

		// Обновляем статус шага на "executing"
		step.Status = SagaStepExecuting
		now := time.Now()
		step.ExecutedAt = &now

		// Выполняем шаг через Saga Manager
		err := coordinator.ExecuteStep(ctx, s.ID, step.ID)
		if err != nil {
			log.Printf("Ошибка выполнения шага %s: %v", step.Name, err)
			step.Error = err.Error()
			step.Status = SagaStepFailed

			// Компенсируем выполненные шаги
			return s.compensate(ctx, coordinator, i)
		}

		// Обновляем статус шага
		step.Status = SagaStepCompleted
		now = time.Now()
		step.CompletedAt = &now
	}

	log.Printf("Saga создания отчета %s выполнена успешно", s.ID)
	return nil
}

// compensate компенсирует выполненные шаги
func (s *ReportCreationSaga) compensate(ctx context.Context, coordinator SagaManager, failedStepIndex int) error {
	log.Printf("Начинаем компенсацию Saga %s с шага %d", s.ID, failedStepIndex)

	// Компенсируем шаги в обратном порядке
	for i := failedStepIndex - 1; i >= 0; i-- {
		step := s.Steps[i]
		if step.Compensate == "none" {
			log.Printf("Шаг %s не требует компенсации", step.Name)
			continue
		}

		log.Printf("Компенсируем шаг: %s", step.Name)

		// Выполняем компенсацию
		err := coordinator.CompensateStep(ctx, s.ID, step.ID)
		if err != nil {
			log.Printf("Ошибка компенсации шага %s: %v", step.Name, err)
			// Продолжаем компенсацию других шагов
		}

		step.Status = SagaStepCompensated
	}

	return fmt.Errorf("Saga %s выполнена с ошибками и компенсирована", s.ID)
}

// generateSagaID генерирует уникальный ID для Saga
func generateSagaID() string {
	return "saga-" + time.Now().Format("20060102150405") + "-" + randomString(6)
}
