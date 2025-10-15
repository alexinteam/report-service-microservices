package events

import (
	"context"
	"log"
	"time"
)

// SagaCoordinator управляет Saga транзакциями
type SagaCoordinator struct {
	publisher EventPublisher
}

// NewSagaCoordinator создает новый Saga Coordinator
func NewSagaCoordinator(publisher EventPublisher) *SagaCoordinator {
	return &SagaCoordinator{
		publisher: publisher,
	}
}

func (sc *SagaCoordinator) StartSaga(ctx context.Context, saga *Saga) error {
	log.Printf("Запуск Saga %s: %s", saga.ID, saga.Name)

	event := NewEvent(SagaStarted, "report-service", map[string]interface{}{
		"saga_id":   saga.ID,
		"saga_name": saga.Name,
		"steps":     len(saga.Steps),
	})

	return sc.publisher.Publish(ctx, event)
}

func (sc *SagaCoordinator) ExecuteStep(ctx context.Context, sagaID string, stepID string) error {
	log.Printf("Выполнение шага %s в Saga %s", stepID, sagaID)

	time.Sleep(100 * time.Millisecond)

	event := NewEvent(ReportGenerated, "report-service", map[string]interface{}{
		"saga_id": sagaID,
		"step_id": stepID,
	})

	return sc.publisher.Publish(ctx, event)
}

func (sc *SagaCoordinator) CompensateStep(ctx context.Context, sagaID string, stepID string) error {
	log.Printf("Компенсация шага %s в Saga %s", stepID, sagaID)

	time.Sleep(50 * time.Millisecond)

	event := NewEvent(SagaCompensated, "report-service", map[string]interface{}{
		"saga_id": sagaID,
		"step_id": stepID,
	})

	return sc.publisher.Publish(ctx, event)
}

func (sc *SagaCoordinator) GetSaga(ctx context.Context, sagaID string) (*Saga, error) {
	return &Saga{
		ID:        sagaID,
		Name:      "Report Creation Saga",
		Status:    SagaStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (sc *SagaCoordinator) UpdateSagaStatus(ctx context.Context, sagaID string, status SagaStatus) error {
	log.Printf("Обновление статуса Saga %s на %s", sagaID, status)

	eventType := SagaCompleted
	if status == SagaStatusFailed {
		eventType = SagaFailed
	}

	event := NewEvent(eventType, "report-service", map[string]interface{}{
		"saga_id": sagaID,
		"status":  string(status),
	})

	return sc.publisher.Publish(ctx, event)
}

func (sc *SagaCoordinator) HandleSagaEvent(ctx context.Context, event *Event) error {
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

func (sc *SagaCoordinator) handleSagaStarted(ctx context.Context, event *Event) error {
	log.Printf("Обработка события SagaStarted для Saga %s", event.Data["saga_id"])
	return nil
}

func (sc *SagaCoordinator) handleSagaCompleted(ctx context.Context, event *Event) error {
	log.Printf("Обработка события SagaCompleted для Saga %s", event.Data["saga_id"])
	return nil
}

func (sc *SagaCoordinator) handleSagaFailed(ctx context.Context, event *Event) error {
	log.Printf("Обработка события SagaFailed для Saga %s", event.Data["saga_id"])
	return nil
}

func (sc *SagaCoordinator) handleSagaCompensated(ctx context.Context, event *Event) error {
	log.Printf("Обработка события SagaCompensated для Saga %s", event.Data["saga_id"])
	return nil
}
