package database

import (
	"fmt"
	"log"

	"report-service/internal/config"
	"report-service/internal/models"

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
		&models.Report{},
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
	db.Model(&models.Report{}).Count(&count)
	if count > 0 {
		log.Println("Тестовые данные уже существуют")
		return nil
	}

	reports := []models.Report{
		{
			Name:        "Отчет по продажам за январь",
			Description: "Ежемесячный отчет по продажам",
			TemplateID:  1,
			UserID:      1,
			Status:      string(models.StatusCompleted),
			Parameters:  `{"start_date": "2024-01-01", "end_date": "2024-01-31"}`,
			FilePath:    "/reports/sales_january_2024.pdf",
			FileSize:    1024000,
			MD5Hash:     "abc123def456",
		},
		{
			Name:        "Отчет по клиентам",
			Description: "Отчет по клиентской базе",
			TemplateID:  2,
			UserID:      2,
			Status:      string(models.StatusPending),
			Parameters:  `{"date": "2024-01-01"}`,
		},
	}

	for _, report := range reports {
		if err := db.Create(&report).Error; err != nil {
			return fmt.Errorf("ошибка создания тестового отчета %s: %w", report.Name, err)
		}
	}

	log.Println("Тестовые данные созданы успешно")
	return nil
}

func Cleanup() error {
	if db == nil {
		return fmt.Errorf("база данных не подключена")
	}

	if err := db.Where("1 = 1").Delete(&models.Report{}).Error; err != nil {
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
