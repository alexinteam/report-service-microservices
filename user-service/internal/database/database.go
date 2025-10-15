package database

import (
	"fmt"
	"log"

	"user-service/internal/config"
	"user-service/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func Connect(cfg *config.Config) (*gorm.DB, error) {
	var err error

	var logLevel logger.LogLevel
	if cfg.IsDevelopment() {
		logLevel = logger.Info
	} else {
		logLevel = logger.Error
	}

	db, err = gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
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

	err := db.AutoMigrate(
		&models.User{},
	)
	if err != nil {
		return fmt.Errorf("ошибка миграции: %w", err)
	}

	log.Println("Миграции выполнены успешно")
	return nil
}

func SeedData() error {
	if db == nil {
		return fmt.Errorf("база данных не подключена")
	}

	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		log.Println("Тестовые данные уже существуют")
		return nil
	}

	users := []models.User{
		{
			Name:     "Администратор",
			Email:    "admin@example.com",
			Password: "admin123",
			Role:     string(models.RoleAdmin),
			IsActive: true,
		},
		{
			Name:     "Пользователь",
			Email:    "user@example.com",
			Password: "user123",
			Role:     string(models.RoleUser),
			IsActive: true,
		},
		{
			Name:     "Менеджер",
			Email:    "manager@example.com",
			Password: "manager123",
			Role:     string(models.RoleManager),
			IsActive: true,
		},
	}

	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("ошибка создания тестового пользователя %s: %w", user.Email, err)
		}
	}

	log.Println("Тестовые данные созданы успешно")
	return nil
}

func Cleanup() error {
	if db == nil {
		return fmt.Errorf("база данных не подключена")
	}

	if err := db.Where("1 = 1").Delete(&models.User{}).Error; err != nil {
		return fmt.Errorf("ошибка очистки данных: %w", err)
	}

	log.Println("Данные очищены успешно")
	return nil
}

func MigrateWithConfig(cfg *config.Config) error {
	_, err := Connect(cfg)
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
	_, err := Connect(cfg)
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	if err := Cleanup(); err != nil {
		return fmt.Errorf("ошибка очистки данных: %w", err)
	}

	log.Println("Данные очищены успешно")
	return nil
}
