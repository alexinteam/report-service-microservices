package database

import (
	"fmt"
	"log"

	"storage-service/internal/config"
	"storage-service/internal/models"

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
		&models.File{},
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

	// Создание тестовых файлов
	files := []models.File{
		{
			Name:        "README.md",
			Path:        "/tmp/reports/README.md",
			Size:        1024,
			MimeType:    "text/markdown",
			Hash:        "d41d8cd98f00b204e9800998ecf8427e",
			Description: "Документация проекта",
			IsPublic:    true,
		},
		{
			Name:        "config.json",
			Path:        "/tmp/reports/config.json",
			Size:        512,
			MimeType:    "application/json",
			Hash:        "e3b0c44298fc1c149afbf4c8996fb924",
			Description: "Конфигурационный файл",
			IsPublic:    false,
		},
		{
			Name:        "logo.png",
			Path:        "/tmp/reports/logo.png",
			Size:        2048,
			MimeType:    "image/png",
			Hash:        "a1b2c3d4e5f6789012345678901234567",
			Description: "Логотип компании",
			IsPublic:    true,
		},
	}

	for _, f := range files {
		if err := db.Create(&f).Error; err != nil {
			log.Printf("Ошибка создания файла %s: %v", f.Name, err)
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
