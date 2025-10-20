package events

import (
	"context"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// RabbitMQPublisher реализует EventPublisher для RabbitMQ
type RabbitMQPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQPublisher(amqpURL string) (*RabbitMQPublisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("ошибка создания канала: %w", err)
	}

	return &RabbitMQPublisher{
		conn:    conn,
		channel: channel,
	}, nil
}

// Publish публикует событие синхронно
func (p *RabbitMQPublisher) Publish(ctx context.Context, event *Event) error {
	return p.publishEvent(event, false)
}

// PublishAsync публикует событие асинхронно
func (p *RabbitMQPublisher) PublishAsync(ctx context.Context, event *Event) error {
	go func() {
		if err := p.publishEvent(event, true); err != nil {
			log.Printf("Ошибка асинхронной публикации события: %v", err)
		}
	}()
	return nil
}

// publishEvent публикует событие
func (p *RabbitMQPublisher) publishEvent(event *Event, async bool) error {
	// Создаем exchange если не существует
	exchangeName := "events"
	err := p.channel.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("ошибка объявления exchange: %w", err)
	}

	// Конвертируем событие в JSON
	body, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("ошибка сериализации события: %w", err)
	}

	// Публикуем сообщение
	routingKey := string(event.Type)
	err = p.channel.Publish(
		exchangeName, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    event.Timestamp,
			MessageId:    event.ID,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		return fmt.Errorf("ошибка публикации сообщения: %w", err)
	}

	if !async {
		log.Printf("Событие %s опубликовано с ID %s", event.Type, event.ID)
	}

	return nil
}

// Close закрывает соединение
func (p *RabbitMQPublisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

// RabbitMQSubscriber реализует EventSubscriber для RabbitMQ
type RabbitMQSubscriber struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewRabbitMQSubscriber создает новый RabbitMQ subscriber
func NewRabbitMQSubscriber(amqpURL string) (*RabbitMQSubscriber, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("ошибка создания канала: %w", err)
	}

	return &RabbitMQSubscriber{
		conn:    conn,
		channel: channel,
	}, nil
}

// Subscribe подписывается на события определенного типа
func (s *RabbitMQSubscriber) Subscribe(ctx context.Context, eventType EventType, handler EventHandler) error {
	// Создаем exchange если не существует
	exchangeName := "events"
	err := s.channel.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("ошибка объявления exchange: %w", err)
	}

	// Создаем очередь
	queueName := fmt.Sprintf("report-service.events.%s", eventType)
	queue, err := s.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("ошибка объявления очереди: %w", err)
	}

	// Привязываем очередь к exchange
	routingKey := string(eventType)
	err = s.channel.QueueBind(
		queue.Name,   // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("ошибка привязки очереди: %w", err)
	}

	// Настраиваем QoS
	err = s.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("ошибка настройки QoS: %w", err)
	}

	// Начинаем потребление сообщений
	msgs, err := s.channel.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return fmt.Errorf("ошибка начала потребления: %w", err)
	}

	// Обрабатываем сообщения в горутине
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgs:
				s.handleMessage(ctx, msg, handler)
			}
		}
	}()

	log.Printf("Report Service подписан на события %s", eventType)
	return nil
}

// handleMessage обрабатывает входящее сообщение
func (s *RabbitMQSubscriber) handleMessage(ctx context.Context, msg amqp.Delivery, handler EventHandler) {
	// Парсим событие
	event, err := FromJSON(msg.Body)
	if err != nil {
		log.Printf("Ошибка парсинга события: %v", err)
		msg.Nack(false, false)
		return
	}

	// Обрабатываем событие
	err = handler.Handle(ctx, event)
	if err != nil {
		log.Printf("Ошибка обработки события %s: %v", event.Type, err)
		msg.Nack(false, true) // Повторяем попытку
		return
	}

	// Подтверждаем обработку
	msg.Ack(false)
	log.Printf("Событие %s обработано Report Service", event.Type)
}

// Unsubscribe отписывается от событий
func (s *RabbitMQSubscriber) Unsubscribe(ctx context.Context, eventType EventType) error {
	log.Printf("Отписка Report Service от событий %s", eventType)
	return nil
}

// Close закрывает соединение
func (s *RabbitMQSubscriber) Close() error {
	if s.channel != nil {
		s.channel.Close()
	}
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
