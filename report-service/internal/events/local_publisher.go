package events

import (
	"context"
	"log"
)

// LocalEventPublisher реализует EventPublisher для локальной публикации
type LocalEventPublisher struct{}

// NewLocalEventPublisher создает новый локальный publisher
func NewLocalEventPublisher() *LocalEventPublisher {
	return &LocalEventPublisher{}
}

// Publish публикует событие локально (просто логирует)
func (p *LocalEventPublisher) Publish(ctx context.Context, event *Event) error {
	log.Printf("Локальная публикация события: %s (ID: %s)", event.Type, event.ID)
	return nil
}

// PublishAsync публикует событие асинхронно локально
func (p *LocalEventPublisher) PublishAsync(ctx context.Context, event *Event) error {
	go func() {
		log.Printf("Асинхронная локальная публикация события: %s (ID: %s)", event.Type, event.ID)
	}()
	return nil
}
