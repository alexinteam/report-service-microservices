package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OutboxEvent представляет событие в Outbox таблице
type OutboxEvent struct {
	ID          string     `gorm:"primaryKey" json:"id"`
	EventType   EventType  `gorm:"not null" json:"event_type"`
	AggregateID string     `gorm:"not null" json:"aggregate_id"`
	Data        string     `gorm:"type:text" json:"data"`
	Status      string     `gorm:"not null;default:'pending'" json:"status"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	RetryCount  int        `gorm:"default:0" json:"retry_count"`
	Error       string     `gorm:"type:text" json:"error,omitempty"`
}

// OutboxManager управляет событиями в Outbox таблице
type OutboxManager struct {
	db *gorm.DB
}

// NewOutboxManager создает новый OutboxManager
func NewOutboxManager(db *gorm.DB) *OutboxManager {
	return &OutboxManager{db: db}
}

// SaveEvent сохраняет событие в Outbox таблице
func (om *OutboxManager) SaveEvent(ctx context.Context, event *Event) error {
	eventData, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("ошибка сериализации данных события: %w", err)
	}

	outboxEvent := &OutboxEvent{
		ID:          uuid.New().String(),
		EventType:   event.Type,
		AggregateID: event.ID, // Используем ID события как AggregateID
		Data:        string(eventData),
		Status:      "pending",
		CreatedAt:   time.Now(),
	}

	if err := om.db.WithContext(ctx).Create(outboxEvent).Error; err != nil {
		return fmt.Errorf("ошибка сохранения события в Outbox: %w", err)
	}
	return nil
}

// GetPendingEvents получает все ожидающие события
func (om *OutboxManager) GetPendingEvents(ctx context.Context, limit int) ([]*OutboxEvent, error) {
	var events []*OutboxEvent
	if err := om.db.WithContext(ctx).Where("status = ?", "pending").Limit(limit).Find(&events).Error; err != nil {
		return nil, fmt.Errorf("ошибка получения ожидающих событий: %w", err)
	}
	return events, nil
}

// MarkAsProcessing помечает событие как обрабатываемое
func (om *OutboxManager) MarkAsProcessing(ctx context.Context, eventID string) error {
	if err := om.db.WithContext(ctx).Model(&OutboxEvent{}).Where("id = ?", eventID).Update("status", "processing").Error; err != nil {
		return fmt.Errorf("ошибка пометки события как обрабатываемого: %w", err)
	}
	return nil
}

// MarkAsProcessed помечает событие как обработанное
func (om *OutboxManager) MarkAsProcessed(ctx context.Context, eventID string) error {
	now := time.Now()
	if err := om.db.WithContext(ctx).Model(&OutboxEvent{}).Where("id = ?", eventID).Updates(map[string]interface{}{
		"status":       "processed",
		"processed_at": &now,
	}).Error; err != nil {
		return fmt.Errorf("ошибка пометки события как обработанного: %w", err)
	}
	return nil
}

// MarkAsFailed помечает событие как неудачное
func (om *OutboxManager) MarkAsFailed(ctx context.Context, eventID string, errMsg string) error {
	if err := om.db.WithContext(ctx).Model(&OutboxEvent{}).Where("id = ?", eventID).Updates(map[string]interface{}{
		"status": "failed",
		"error":  errMsg,
	}).Error; err != nil {
		return fmt.Errorf("ошибка пометки события как неудачного: %w", err)
	}
	return nil
}

// IncrementRetryCount увеличивает счетчик попыток
func (om *OutboxManager) IncrementRetryCount(ctx context.Context, eventID string) error {
	if err := om.db.WithContext(ctx).Model(&OutboxEvent{}).Where("id = ?", eventID).UpdateColumn("retry_count", gorm.Expr("retry_count + ?", 1)).Error; err != nil {
		return fmt.Errorf("ошибка увеличения счетчика попыток: %w", err)
	}
	return nil
}

// OutboxPublisher публикует события из Outbox
type OutboxPublisher struct {
	outboxManager  *OutboxManager
	eventPublisher EventPublisher
}

// NewOutboxPublisher создает новый OutboxPublisher
func NewOutboxPublisher(om *OutboxManager, ep EventPublisher) *OutboxPublisher {
	return &OutboxPublisher{
		outboxManager:  om,
		eventPublisher: ep,
	}
}

// StartPublishing запускает процесс публикации событий из Outbox
func (op *OutboxPublisher) StartPublishing(ctx context.Context, interval time.Duration, batchSize int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Остановка Outbox Publisher")
			return
		case <-ticker.C:
			op.publishPendingEvents(ctx, batchSize)
		}
	}
}

func (op *OutboxPublisher) publishPendingEvents(ctx context.Context, batchSize int) {
	eventsToPublish, err := op.outboxManager.GetPendingEvents(ctx, batchSize)
	if err != nil {
		log.Printf("Ошибка получения ожидающих событий из Outbox: %v", err)
		return
	}

	if len(eventsToPublish) == 0 {
		return
	}

	log.Printf("Найдено %d ожидающих событий для публикации", len(eventsToPublish))

	for _, event := range eventsToPublish {
		// Помечаем событие как обрабатываемое
		if err := op.outboxManager.MarkAsProcessing(ctx, event.ID); err != nil {
			log.Printf("Ошибка пометки события %s как обрабатываемого: %v", event.ID, err)
			continue
		}

		// Десериализуем данные события
		var eventData map[string]interface{}
		if err := json.Unmarshal([]byte(event.Data), &eventData); err != nil {
			log.Printf("Ошибка десериализации данных события %s: %v", event.ID, err)
			op.outboxManager.MarkAsFailed(ctx, event.ID, fmt.Sprintf("ошибка десериализации: %v", err))
			continue
		}

		// Создаем объект Event для публикации
		eventToPublish := &Event{
			ID:        event.ID,
			Type:      event.EventType,
			Source:    "report-service",
			Timestamp: event.CreatedAt,
			Data:      eventData,
		}

		// Публикуем событие
		if err := op.eventPublisher.Publish(ctx, eventToPublish); err != nil {
			log.Printf("Ошибка публикации события %s: %v", event.ID, err)
			// Увеличиваем счетчик попыток
			op.outboxManager.IncrementRetryCount(ctx, event.ID)
			// Помечаем как неудачное
			op.outboxManager.MarkAsFailed(ctx, event.ID, err.Error())
			continue
		}

		// Помечаем событие как обработанное
		if err := op.outboxManager.MarkAsProcessed(ctx, event.ID); err != nil {
			log.Printf("Ошибка пометки события %s как обработанного: %v", event.ID, err)
		}
	}
}

// MigrateOutboxTable создает таблицу Outbox
func (om *OutboxManager) MigrateOutboxTable(ctx context.Context) error {
	return om.db.WithContext(ctx).AutoMigrate(&OutboxEvent{})
}
