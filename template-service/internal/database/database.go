package database

import (
	"fmt"
	"log"

	"template-service/internal/config"
	"template-service/internal/models"

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

	if err := db.AutoMigrate(
		&models.Template{},
		&models.TemplateCategory{},
		&models.TemplateVariable{},
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

	categories := []models.TemplateCategory{
		{
			Name:        "Финансовые отчеты",
			Description: "Отчеты по финансовой деятельности",
			IsActive:    true,
		},
		{
			Name:        "Продажи",
			Description: "Отчеты по продажам и маркетингу",
			IsActive:    true,
		},
		{
			Name:        "HR отчеты",
			Description: "Отчеты по персоналу",
			IsActive:    true,
		},
	}

	for _, c := range categories {
		if err := db.Create(&c).Error; err != nil {
			log.Printf("Ошибка создания категории %s: %v", c.Name, err)
		}
	}

	templates := []models.Template{
		{
			Name:        "Отчет по продажам",
			Description: "Ежемесячный отчет по продажам",
			Content:     "<h1>Отчет по продажам</h1><p>Период: {{start_date}} - {{end_date}}</p><p>Общая сумма: {{total_amount}} руб.</p>",
			Type:        "html",
			Category:    "Продажи",
			Variables:   `["start_date", "end_date", "total_amount"]`,
			IsActive:    true,
		},
		{
			Name:        "Финансовый отчет",
			Description: "Квартальный финансовый отчет",
			Content:     "<h1>Финансовый отчет</h1><p>Квартал: {{quarter}}</p><p>Доходы: {{income}} руб.</p><p>Расходы: {{expenses}} руб.</p>",
			Type:        "html",
			Category:    "Финансовые отчеты",
			Variables:   `["quarter", "income", "expenses"]`,
			IsActive:    true,
		},
		{
			Name:        "Отчет по персоналу",
			Description: "Отчет по сотрудникам",
			Content:     "<h1>Отчет по персоналу</h1><p>Количество сотрудников: {{employee_count}}</p><p>Средняя зарплата: {{avg_salary}} руб.</p>",
			Type:        "html",
			Category:    "HR отчеты",
			Variables:   `["employee_count", "avg_salary"]`,
			IsActive:    true,
		},
	}

	for _, t := range templates {
		if err := db.Create(&t).Error; err != nil {
			log.Printf("Ошибка создания шаблона %s: %v", t.Name, err)
		}
	}

	log.Println("Тестовые данные созданы успешно")
	return nil
}

func Cleanup() error {
	if db == nil {
		return fmt.Errorf("база данных не подключена")
	}

	if err := db.Where("1 = 1").Delete(&models.Template{}).Error; err != nil {
		return fmt.Errorf("ошибка очистки данных: %w", err)
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

	// Очистка данных
	if err := Cleanup(); err != nil {
		return fmt.Errorf("ошибка очистки данных: %w", err)
	}

	log.Println("Данные очищены успешно")
	return nil
}
