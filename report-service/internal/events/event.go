package events

import (
	"encoding/json"
	"time"
)

type EventType string

const (
	// Report Events
	ReportCreated   EventType = "report.created"
	ReportUpdated   EventType = "report.updated"
	ReportDeleted   EventType = "report.deleted"
	ReportGenerated EventType = "report.generated"
	ReportCompleted EventType = "report.completed"
	ReportFailed    EventType = "report.failed"

	// Saga Events
	SagaStarted     EventType = "saga.started"
	SagaCompleted   EventType = "saga.completed"
	SagaFailed      EventType = "saga.failed"
	SagaCompensated EventType = "saga.compensated"

	// User Events (для Saga)
	UserValidated        EventType = "user.validated"
	UserValidationFailed EventType = "user.validation_failed"

	// Template Events (для Saga)
	TemplateValidated        EventType = "template.validated"
	TemplateValidationFailed EventType = "template.validation_failed"

	// Data Events (для Saga)
	DataCollected        EventType = "data.collected"
	DataCollectionFailed EventType = "data.collection_failed"

	// Storage Events (для Saga)
	FileStored        EventType = "file.stored"
	FileStorageFailed EventType = "file.storage_failed"
)

// Event представляет базовое событие
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// NewEvent создает новое событие
func NewEvent(eventType EventType, source string, data map[string]interface{}) *Event {
	return &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Source:    source,
		Timestamp: time.Now(),
		Data:      data,
		Metadata:  make(map[string]interface{}),
	}
}

// ToJSON конвертирует событие в JSON
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON создает событие из JSON
func FromJSON(data []byte) (*Event, error) {
	var event Event
	err := json.Unmarshal(data, &event)
	return &event, err
}

// generateEventID генерирует уникальный ID события
func generateEventID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString генерирует случайную строку заданной длины
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
