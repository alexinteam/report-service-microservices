package database

import (
	"fmt"
	"log"

	"notification-service/internal/config"
	"notification-service/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func Connect(databaseURL string) (*gorm.DB, error) {
	var err error

	db, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения подключения к БД: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("Подключение к базе данных установлено")
	return db, nil
}

func Migrate() error {
	if db == nil {
		return fmt.Errorf("база данных не подключена")
	}

	// Миграция моделей
	if err := db.AutoMigrate(
		&models.NotificationTemplate{},
		&models.Notification{},
		&models.NotificationChannel{},
	); err != nil {
		return fmt.Errorf("ошибка миграции моделей: %w", err)
	}

	log.Println("Миграции выполнены успешно")
	return nil
}

func SeedData() error {
	if db == nil {
		return fmt.Errorf("база данных не подключена")
	}

	// Создание тестовых шаблонов уведомлений
	templates := []models.NotificationTemplate{
		{
			Name:      "Welcome Email",
			Subject:   "Добро пожаловать!",
			Body:      "Привет, {{name}}! Добро пожаловать в нашу систему.",
			Type:      "email",
			Variables: `["name", "email"]`,
			IsActive:  true,
		},
		{
			Name:      "Password Reset",
			Subject:   "Сброс пароля",
			Body:      "Для сброса пароля перейдите по ссылке: {{reset_link}}",
			Type:      "email",
			Variables: `["reset_link", "email"]`,
			IsActive:  true,
		},
		{
			Name:      "SMS Notification",
			Subject:   "",
			Body:      "Ваш код подтверждения: {{code}}",
			Type:      "sms",
			Variables: `["code", "phone"]`,
			IsActive:  true,
		},
		{
			Name:      "Push Notification",
			Subject:   "Новое уведомление",
			Body:      "У вас новое сообщение: {{message}}",
			Type:      "push",
			Variables: `["message", "user_id"]`,
			IsActive:  true,
		},
	}

	for _, t := range templates {
		if err := db.Create(&t).Error; err != nil {
			log.Printf("Ошибка создания шаблона %s: %v", t.Name, err)
		}
	}

	// Создание тестовых каналов уведомлений
	channels := []models.NotificationChannel{
		{
			Name:     "SMTP Email",
			Type:     "email",
			Config:   `{"host": "smtp.gmail.com", "port": 587, "username": "noreply@example.com"}`,
			IsActive: true,
		},
		{
			Name:     "SMS Gateway",
			Type:     "sms",
			Config:   `{"provider": "twilio", "account_sid": "AC123", "auth_token": "token"}`,
			IsActive: true,
		},
		{
			Name:     "Firebase Push",
			Type:     "push",
			Config:   `{"server_key": "AAAA...", "project_id": "my-project"}`,
			IsActive: true,
		},
		{
			Name:     "Webhook",
			Type:     "webhook",
			Config:   `{"url": "https://webhook.site/123", "method": "POST"}`,
			IsActive: true,
		},
	}

	for _, c := range channels {
		if err := db.Create(&c).Error; err != nil {
			log.Printf("Ошибка создания канала %s: %v", c.Name, err)
		}
	}

	log.Println("Тестовые данные созданы успешно")
	return nil
}

func Cleanup() error {
	if db == nil {
		return fmt.Errorf("база данных не подключена")
	}

	log.Println("Данные очищены успешно")
	return nil
}

func MigrateWithConfig(cfg *config.Config) error {
	_, err := Connect(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	if err := Migrate(); err != nil {
		return fmt.Errorf("ошибка миграции: %w", err)
	}

	if cfg.SeedData {
		if err := SeedData(); err != nil {
			return fmt.Errorf("ошибка заполнения тестовыми данными: %w", err)
		}
	}

	log.Println("Миграции выполнены успешно")
	return nil
}

func CleanupWithConfig(cfg *config.Config) error {
	_, err := Connect(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	if err := Cleanup(); err != nil {
		return fmt.Errorf("ошибка очистки данных: %w", err)
	}

	log.Println("Данные очищены успешно")
	return nil
}
