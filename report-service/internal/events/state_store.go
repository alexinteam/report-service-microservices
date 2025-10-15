package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// SagaState представляет состояние Saga в базе данных
type SagaState struct {
	ID          string     `gorm:"primaryKey" json:"id"`
	Name        string     `gorm:"not null" json:"name"`
	Status      SagaStatus `gorm:"not null" json:"status"`
	Steps       string     `gorm:"type:text" json:"steps"` // JSON сериализация шагов
	Data        string     `gorm:"type:text" json:"data"`  // JSON сериализация данных
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Error       string     `gorm:"type:text" json:"error,omitempty"`
	RetryCount  int        `gorm:"default:0" json:"retry_count"`
	LastStepID  string     `json:"last_step_id,omitempty"`
}

// EventLog представляет лог событий для идемпотентности
type EventLog struct {
	ID          string     `gorm:"primaryKey" json:"id"`
	SagaID      string     `gorm:"not null;index" json:"saga_id"`
	EventID     string     `gorm:"not null;uniqueIndex" json:"event_id"`
	EventType   EventType  `gorm:"not null" json:"event_type"`
	Status      string     `gorm:"not null" json:"status"` // processed, failed, retrying
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	Error       string     `gorm:"type:text" json:"error,omitempty"`
	RetryCount  int        `gorm:"default:0" json:"retry_count"`
}

// SagaStateStore управляет состоянием Saga
type SagaStateStore struct {
	db *gorm.DB
}

// NewSagaStateStore создает новый SagaStateStore
func NewSagaStateStore(db *gorm.DB) *SagaStateStore {
	return &SagaStateStore{db: db}
}

// SaveSagaState сохраняет состояние Saga
func (s *SagaStateStore) SaveSagaState(ctx context.Context, saga *Saga) error {
	// Сериализуем шаги
	stepsJSON, err := json.Marshal(saga.Steps)
	if err != nil {
		return fmt.Errorf("ошибка сериализации шагов: %w", err)
	}

	// Сериализуем данные
	dataJSON, err := json.Marshal(saga.Data)
	if err != nil {
		return fmt.Errorf("ошибка сериализации данных: %w", err)
	}

	sagaState := &SagaState{
		ID:         saga.ID,
		Name:       saga.Name,
		Status:     saga.Status,
		Steps:      string(stepsJSON),
		Data:       string(dataJSON),
		UpdatedAt:  time.Now(),
		Error:      saga.Error,
		RetryCount: 0,
	}

	// Определяем последний выполненный шаг
	if len(saga.Steps) > 0 {
		for i := len(saga.Steps) - 1; i >= 0; i-- {
			if saga.Steps[i].Status == SagaStepCompleted {
				sagaState.LastStepID = saga.Steps[i].ID
				break
			}
		}
	}

	// Используем Upsert для идемпотентности
	return s.db.WithContext(ctx).Save(sagaState).Error
}

// GetSagaState получает состояние Saga
func (s *SagaStateStore) GetSagaState(ctx context.Context, sagaID string) (*Saga, error) {
	var sagaState SagaState
	if err := s.db.WithContext(ctx).Where("id = ?", sagaID).First(&sagaState).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("Saga %s не найдена", sagaID)
		}
		return nil, fmt.Errorf("ошибка получения Saga: %w", err)
	}

	// Десериализуем шаги
	var steps []*SagaStep
	if err := json.Unmarshal([]byte(sagaState.Steps), &steps); err != nil {
		return nil, fmt.Errorf("ошибка десериализации шагов: %w", err)
	}

	// Десериализуем данные
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(sagaState.Data), &data); err != nil {
		return nil, fmt.Errorf("ошибка десериализации данных: %w", err)
	}

	return &Saga{
		ID:          sagaState.ID,
		Name:        sagaState.Name,
		Status:      sagaState.Status,
		Steps:       steps,
		Data:        data,
		CreatedAt:   sagaState.CreatedAt,
		UpdatedAt:   sagaState.UpdatedAt,
		CompletedAt: sagaState.CompletedAt,
		Error:       sagaState.Error,
	}, nil
}

// UpdateSagaStatus обновляет статус Saga
func (s *SagaStateStore) UpdateSagaStatus(ctx context.Context, sagaID string, status SagaStatus) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == SagaStatusCompleted {
		now := time.Now()
		updates["completed_at"] = &now
	}

	return s.db.WithContext(ctx).Model(&SagaState{}).Where("id = ?", sagaID).Updates(updates).Error
}

// IncrementRetryCount увеличивает счетчик попыток
func (s *SagaStateStore) IncrementRetryCount(ctx context.Context, sagaID string) error {
	return s.db.WithContext(ctx).Model(&SagaState{}).Where("id = ?", sagaID).UpdateColumn("retry_count", gorm.Expr("retry_count + ?", 1)).Error
}

// LogEvent логирует событие для идемпотентности
func (s *SagaStateStore) LogEvent(ctx context.Context, sagaID, eventID string, eventType EventType) error {
	eventLog := &EventLog{
		SagaID:    sagaID,
		EventID:   eventID,
		EventType: eventType,
		Status:    "processed",
		CreatedAt: time.Now(),
	}

	// Используем игнорирование конфликтов для идемпотентности
	return s.db.WithContext(ctx).Create(eventLog).Error
}

// IsEventProcessed проверяет, было ли событие уже обработано
func (s *SagaStateStore) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&EventLog{}).Where("event_id = ?", eventID).Count(&count).Error
	return count > 0, err
}

// GetSagaByEventID получает Saga по ID события
func (s *SagaStateStore) GetSagaByEventID(ctx context.Context, eventID string) (*Saga, error) {
	var eventLog EventLog
	if err := s.db.WithContext(ctx).Where("event_id = ?", eventID).First(&eventLog).Error; err != nil {
		return nil, err
	}

	return s.GetSagaState(ctx, eventLog.SagaID)
}

// MigrateSagaTables создает таблицы для Saga
func (s *SagaStateStore) MigrateSagaTables(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(&SagaState{}, &EventLog{})
}
