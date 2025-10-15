package events

import (
	"context"
	"time"
)

// EventHandler представляет обработчик событий
type EventHandler interface {
	Handle(ctx context.Context, event *Event) error
	EventType() EventType
}

// EventPublisher представляет издатель событий
type EventPublisher interface {
	Publish(ctx context.Context, event *Event) error
	PublishAsync(ctx context.Context, event *Event) error
}

// EventSubscriber представляет подписчик на события
type EventSubscriber interface {
	Subscribe(ctx context.Context, eventType EventType, handler EventHandler) error
	Unsubscribe(ctx context.Context, eventType EventType) error
}

// SagaStep представляет шаг в Saga
type SagaStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Service     string                 `json:"service"`
	Action      string                 `json:"action"`
	Compensate  string                 `json:"compensate"`
	Data        map[string]interface{} `json:"data"`
	Status      SagaStepStatus         `json:"status"`
	Error       string                 `json:"error,omitempty"`
	ExecutedAt  *time.Time             `json:"executed_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// SagaStepStatus представляет статус шага Saga
type SagaStepStatus string

const (
	SagaStepPending     SagaStepStatus = "pending"
	SagaStepExecuting   SagaStepStatus = "executing"
	SagaStepCompleted   SagaStepStatus = "completed"
	SagaStepFailed      SagaStepStatus = "failed"
	SagaStepCompensated SagaStepStatus = "compensated"
)

// Saga представляет Saga транзакцию
type Saga struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Status      SagaStatus             `json:"status"`
	Steps       []*SagaStep            `json:"steps"`
	Data        map[string]interface{} `json:"data"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// SagaStatus представляет статус Saga
type SagaStatus string

const (
	SagaStatusPending     SagaStatus = "pending"
	SagaStatusExecuting   SagaStatus = "executing"
	SagaStatusCompleted   SagaStatus = "completed"
	SagaStatusFailed      SagaStatus = "failed"
	SagaStatusCompensated SagaStatus = "compensated"
)

// SagaManager управляет Saga транзакциями
type SagaManager interface {
	StartSaga(ctx context.Context, saga *Saga) error
	ExecuteStep(ctx context.Context, sagaID string, stepID string) error
	CompensateStep(ctx context.Context, sagaID string, stepID string) error
	GetSaga(ctx context.Context, sagaID string) (*Saga, error)
	UpdateSagaStatus(ctx context.Context, sagaID string, status SagaStatus) error
}
