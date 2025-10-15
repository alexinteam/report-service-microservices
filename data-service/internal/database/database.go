package database

import (
	"fmt"
	"log"

	"data-service/internal/config"
	"data-service/internal/models"

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
		&models.DataSource{},
		&models.DataCollection{},
		&models.DataRecord{},
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

	dataSources := []models.DataSource{
		{
			Name:        "PostgreSQL Database",
			Description: "Основная база данных PostgreSQL",
			Type:        "database",
			Config:      `{"host": "localhost", "port": 5432, "database": "main_db"}`,
			IsActive:    true,
		},
		{
			Name:        "REST API",
			Description: "Внешний REST API",
			Type:        "api",
			Config:      `{"url": "https://api.example.com", "auth": "bearer"}`,
			IsActive:    true,
		},
		{
			Name:        "CSV File",
			Description: "Файл с данными в формате CSV",
			Type:        "file",
			Config:      `{"path": "/data/sample.csv", "format": "csv"}`,
			IsActive:    true,
		},
	}

	for _, ds := range dataSources {
		if err := db.Create(&ds).Error; err != nil {
			log.Printf("Ошибка создания источника данных %s: %v", ds.Name, err)
		}
	}

	dataCollections := []models.DataCollection{
		{
			Name:         "User Data Collection",
			Description:  "Сбор данных о пользователях",
			DataSourceID: 1,
			Query:        "SELECT id, name, email FROM users WHERE created_at > ?",
			Parameters:   `{"date": "2024-01-01"}`,
			IsActive:     true,
		},
		{
			Name:         "Sales Data Collection",
			Description:  "Сбор данных о продажах",
			DataSourceID: 2,
			Query:        "",
			Parameters:   `{"endpoint": "/sales", "method": "GET"}`,
			IsActive:     true,
		},
	}

	for _, dc := range dataCollections {
		if err := db.Create(&dc).Error; err != nil {
			log.Printf("Ошибка создания сбора данных %s: %v", dc.Name, err)
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
