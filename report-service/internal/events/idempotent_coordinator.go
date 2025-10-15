package events

import (
	"context"
	"fmt"
	"log"
	"time"
)

// IdempotentSagaCoordinator управляет Saga с идемпотентностью
type IdempotentSagaCoordinator struct {
	publisher   EventPublisher
	stateStore  *SagaStateStore
	maxRetries  int
	retryDelay  time.Duration
	stepHandler SagaStepHandlerInterface
}

// SagaStepHandlerInterface интерфейс для обработки шагов Saga
type SagaStepHandlerInterface interface {
	ExecuteStep(ctx context.Context, step *SagaStep) error
	CompensateStep(ctx context.Context, step *SagaStep) error
}

// NewIdempotentSagaCoordinator создает новый идемпотентный Saga Coordinator
func NewIdempotentSagaCoordinator(publisher EventPublisher, stateStore *SagaStateStore, stepHandler SagaStepHandlerInterface) *IdempotentSagaCoordinator {
	return &IdempotentSagaCoordinator{
		publisher:   publisher,
		stateStore:  stateStore,
		maxRetries:  3,
		retryDelay:  5 * time.Second,
		stepHandler: stepHandler,
	}
}

// StartSaga запускает новую Saga с проверкой идемпотентности
func (sc *IdempotentSagaCoordinator) StartSaga(ctx context.Context, saga *Saga) error {
	// Проверяем, не существует ли уже Saga с таким ID
	existingSaga, err := sc.stateStore.GetSagaState(ctx, saga.ID)
	if err == nil {
		// Saga уже существует, проверяем статус
		switch existingSaga.Status {
		case SagaStatusCompleted:
			log.Printf("Saga %s уже выполнена успешно", saga.ID)
			return nil
		case SagaStatusFailed:
			log.Printf("Saga %s ранее завершилась с ошибкой, начинаем повторное выполнение", saga.ID)
			// Сбрасываем статус для повторного выполнения
			saga.Status = SagaStatusPending
		case SagaStatusExecuting:
			log.Printf("Saga %s уже выполняется", saga.ID)
			return fmt.Errorf("Saga %s уже выполняется", saga.ID)
		}
	}

	log.Printf("Запуск Saga %s: %s", saga.ID, saga.Name)

	// Сохраняем начальное состояние Saga
	saga.Status = SagaStatusExecuting
	if err := sc.stateStore.SaveSagaState(ctx, saga); err != nil {
		return fmt.Errorf("ошибка сохранения состояния Saga: %w", err)
	}

	// Публикуем событие начала Saga
	event := NewEvent(SagaStarted, "report-service", map[string]interface{}{
		"saga_id":   saga.ID,
		"saga_name": saga.Name,
		"steps":     len(saga.Steps),
	})

	// Логируем событие для идемпотентности
	if err := sc.stateStore.LogEvent(ctx, saga.ID, event.ID, event.Type); err != nil {
		log.Printf("Предупреждение: не удалось залогировать событие %s: %v", event.ID, err)
	}

	return sc.publisher.Publish(ctx, event)
}

// ExecuteStep выполняет шаг Saga с идемпотентностью
func (sc *IdempotentSagaCoordinator) ExecuteStep(ctx context.Context, sagaID string, stepID string) error {
	// Получаем текущее состояние Saga
	saga, err := sc.stateStore.GetSagaState(ctx, sagaID)
	if err != nil {
		return fmt.Errorf("ошибка получения Saga %s: %w", sagaID, err)
	}

	// Находим шаг
	var step *SagaStep
	for _, s := range saga.Steps {
		if s.ID == stepID {
			step = s
			break
		}
	}
	if step == nil {
		return fmt.Errorf("шаг %s не найден в Saga %s", stepID, sagaID)
	}

	// Проверяем идемпотентность шага
	if step.Status == SagaStepCompleted {
		log.Printf("Шаг %s уже выполнен в Saga %s", stepID, sagaID)
		return nil
	}

	if step.Status == SagaStepExecuting {
		log.Printf("Шаг %s уже выполняется в Saga %s", stepID, sagaID)
		return fmt.Errorf("шаг %s уже выполняется", stepID)
	}

	log.Printf("Выполнение шага %s в Saga %s", stepID, sagaID)

	// Обновляем статус шага
	step.Status = SagaStepExecuting
	now := time.Now()
	step.ExecutedAt = &now

	// Сохраняем состояние
	if err := sc.stateStore.SaveSagaState(ctx, saga); err != nil {
		return fmt.Errorf("ошибка сохранения состояния Saga: %w", err)
	}

	// Выполняем шаг с повторными попытками
	for attempt := 0; attempt <= sc.maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Повторная попытка %d для шага %s", attempt, stepID)
			time.Sleep(sc.retryDelay)
		}

		err := sc.executeStepInternal(ctx, sagaID, stepID, step)
		if err == nil {
			// Шаг выполнен успешно
			step.Status = SagaStepCompleted
			now := time.Now()
			step.CompletedAt = &now
			step.Error = ""

			// Сохраняем обновленное состояние
			if err := sc.stateStore.SaveSagaState(ctx, saga); err != nil {
				log.Printf("Ошибка сохранения состояния после выполнения шага: %v", err)
			}

			log.Printf("Шаг %s выполнен успешно в Saga %s", stepID, sagaID)
			return nil
		}

		// Ошибка выполнения
		log.Printf("Ошибка выполнения шага %s (попытка %d): %v", stepID, attempt+1, err)

		if attempt == sc.maxRetries {
			// Исчерпаны все попытки
			step.Status = SagaStepFailed
			step.Error = err.Error()

			// Сохраняем состояние с ошибкой
			if saveErr := sc.stateStore.SaveSagaState(ctx, saga); saveErr != nil {
				log.Printf("Ошибка сохранения состояния с ошибкой: %v", saveErr)
			}

			// Увеличиваем счетчик попыток Saga
			sc.stateStore.IncrementRetryCount(ctx, sagaID)

			return fmt.Errorf("шаг %s не выполнен после %d попыток: %w", stepID, sc.maxRetries+1, err)
		}
	}

	return nil
}

// executeStepInternal выполняет внутреннюю логику шага
func (sc *IdempotentSagaCoordinator) executeStepInternal(ctx context.Context, sagaID, stepID string, step *SagaStep) error {
	log.Printf("Выполняем %s.%s для Saga %s", step.Service, step.Action, sagaID)

	// Используем обработчик шагов, если он доступен
	if sc.stepHandler != nil {
		if err := sc.stepHandler.ExecuteStep(ctx, step); err != nil {
			return fmt.Errorf("ошибка выполнения шага через обработчик: %w", err)
		}
	} else {
		// Fallback к старой логике
		time.Sleep(100 * time.Millisecond)
		if time.Now().UnixNano()%10 == 0 {
			return fmt.Errorf("симулированная ошибка выполнения %s.%s", step.Service, step.Action)
		}
	}

	// Публикуем событие выполнения шага
	event := NewEvent(ReportGenerated, "report-service", map[string]interface{}{
		"saga_id": sagaID,
		"step_id": stepID,
		"service": step.Service,
		"action":  step.Action,
	})

	// Логируем событие для идемпотентности
	if err := sc.stateStore.LogEvent(ctx, sagaID, event.ID, event.Type); err != nil {
		log.Printf("Предупреждение: не удалось залогировать событие %s: %v", event.ID, err)
	}

	return sc.publisher.Publish(ctx, event)
}

// CompensateStep компенсирует шаг Saga с идемпотентностью
func (sc *IdempotentSagaCoordinator) CompensateStep(ctx context.Context, sagaID string, stepID string) error {
	// Получаем текущее состояние Saga
	saga, err := sc.stateStore.GetSagaState(ctx, sagaID)
	if err != nil {
		return fmt.Errorf("ошибка получения Saga %s: %w", sagaID, err)
	}

	// Находим шаг
	var step *SagaStep
	for _, s := range saga.Steps {
		if s.ID == stepID {
			step = s
			break
		}
	}
	if step == nil {
		return fmt.Errorf("шаг %s не найден в Saga %s", stepID, sagaID)
	}

	// Проверяем идемпотентность компенсации
	if step.Status == SagaStepCompensated {
		log.Printf("Шаг %s уже компенсирован в Saga %s", stepID, sagaID)
		return nil
	}

	if step.Compensate == "none" {
		log.Printf("Шаг %s не требует компенсации", stepID)
		step.Status = SagaStepCompensated
		return sc.stateStore.SaveSagaState(ctx, saga)
	}

	log.Printf("Компенсация шага %s в Saga %s", stepID, sagaID)

	// Выполняем компенсацию с повторными попытками
	for attempt := 0; attempt <= sc.maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Повторная попытка компенсации %d для шага %s", attempt, stepID)
			time.Sleep(sc.retryDelay)
		}

		err := sc.compensateStepInternal(ctx, sagaID, stepID, step)
		if err == nil {
			// Компенсация выполнена успешно
			step.Status = SagaStepCompensated

			// Сохраняем обновленное состояние
			if err := sc.stateStore.SaveSagaState(ctx, saga); err != nil {
				log.Printf("Ошибка сохранения состояния после компенсации: %v", err)
			}

			log.Printf("Шаг %s компенсирован успешно в Saga %s", stepID, sagaID)
			return nil
		}

		log.Printf("Ошибка компенсации шага %s (попытка %d): %v", stepID, attempt+1, err)

		if attempt == sc.maxRetries {
			log.Printf("Не удалось компенсировать шаг %s после %d попыток", stepID, sc.maxRetries+1)
			// Продолжаем компенсацию других шагов
			return err
		}
	}

	return nil
}

// compensateStepInternal выполняет внутреннюю логику компенсации
func (sc *IdempotentSagaCoordinator) compensateStepInternal(ctx context.Context, sagaID, stepID string, step *SagaStep) error {
	log.Printf("Компенсируем %s.%s для Saga %s", step.Service, step.Compensate, sagaID)

	// Используем обработчик шагов, если он доступен
	if sc.stepHandler != nil {
		if err := sc.stepHandler.CompensateStep(ctx, step); err != nil {
			return fmt.Errorf("ошибка компенсации шага через обработчик: %w", err)
		}
	} else {
		// Fallback к старой логике
		time.Sleep(50 * time.Millisecond)
		if time.Now().UnixNano()%20 == 0 {
			return fmt.Errorf("симулированная ошибка компенсации %s.%s", step.Service, step.Compensate)
		}
	}

	// Публикуем событие компенсации
	event := NewEvent(SagaCompensated, "report-service", map[string]interface{}{
		"saga_id": sagaID,
		"step_id": stepID,
		"service": step.Service,
		"action":  step.Compensate,
	})

	// Логируем событие для идемпотентности
	if err := sc.stateStore.LogEvent(ctx, sagaID, event.ID, event.Type); err != nil {
		log.Printf("Предупреждение: не удалось залогировать событие %s: %v", event.ID, err)
	}

	return sc.publisher.Publish(ctx, event)
}

// GetSaga получает информацию о Saga
func (sc *IdempotentSagaCoordinator) GetSaga(ctx context.Context, sagaID string) (*Saga, error) {
	return sc.stateStore.GetSagaState(ctx, sagaID)
}

// UpdateSagaStatus обновляет статус Saga
func (sc *IdempotentSagaCoordinator) UpdateSagaStatus(ctx context.Context, sagaID string, status SagaStatus) error {
	log.Printf("Обновление статуса Saga %s на %s", sagaID, status)

	// Обновляем статус в базе данных
	if err := sc.stateStore.UpdateSagaStatus(ctx, sagaID, status); err != nil {
		return fmt.Errorf("ошибка обновления статуса Saga: %w", err)
	}

	// Публикуем событие обновления статуса
	eventType := SagaCompleted
	if status == SagaStatusFailed {
		eventType = SagaFailed
	}

	event := NewEvent(eventType, "report-service", map[string]interface{}{
		"saga_id": sagaID,
		"status":  string(status),
	})

	// Логируем событие для идемпотентности
	if err := sc.stateStore.LogEvent(ctx, sagaID, event.ID, event.Type); err != nil {
		log.Printf("Предупреждение: не удалось залогировать событие %s: %v", event.ID, err)
	}

	return sc.publisher.Publish(ctx, event)
}

// HandleSagaEvent обрабатывает события Saga с проверкой идемпотентности
func (sc *IdempotentSagaCoordinator) HandleSagaEvent(ctx context.Context, event *Event) error {
	// Проверяем, не было ли событие уже обработано
	processed, err := sc.stateStore.IsEventProcessed(ctx, event.ID)
	if err != nil {
		log.Printf("Ошибка проверки идемпотентности события %s: %v", event.ID, err)
		// Продолжаем обработку, так как это не критично
	} else if processed {
		log.Printf("Событие %s уже было обработано, пропускаем", event.ID)
		return nil
	}

	// Логируем событие как обрабатываемое
	if err := sc.stateStore.LogEvent(ctx, event.Data["saga_id"].(string), event.ID, event.Type); err != nil {
		log.Printf("Предупреждение: не удалось залогировать событие %s: %v", event.ID, err)
	}

	// Обрабатываем событие
	switch event.Type {
	case SagaStarted:
		return sc.handleSagaStarted(ctx, event)
	case SagaCompleted:
		return sc.handleSagaCompleted(ctx, event)
	case SagaFailed:
		return sc.handleSagaFailed(ctx, event)
	case SagaCompensated:
		return sc.handleSagaCompensated(ctx, event)
	default:
		log.Printf("Неизвестный тип события Saga: %s", event.Type)
		return nil
	}
}

func (sc *IdempotentSagaCoordinator) handleSagaStarted(ctx context.Context, event *Event) error {
	log.Printf("Обработка события SagaStarted для Saga %s", event.Data["saga_id"])
	return nil
}

func (sc *IdempotentSagaCoordinator) handleSagaCompleted(ctx context.Context, event *Event) error {
	log.Printf("Обработка события SagaCompleted для Saga %s", event.Data["saga_id"])
	return nil
}

func (sc *IdempotentSagaCoordinator) handleSagaFailed(ctx context.Context, event *Event) error {
	log.Printf("Обработка события SagaFailed для Saga %s", event.Data["saga_id"])
	return nil
}

func (sc *IdempotentSagaCoordinator) handleSagaCompensated(ctx context.Context, event *Event) error {
	log.Printf("Обработка события SagaCompensated для Saga %s", event.Data["saga_id"])
	return nil
}
