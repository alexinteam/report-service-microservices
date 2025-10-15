package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Port        int    `envconfig:"PORT" default:"8080"`
	Environment string `envconfig:"ENVIRONMENT" default:"development"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
	JWTSecret   string `envconfig:"JWT_SECRET" default:"your-secret-key-change-in-production"`

	UserServiceURL         string `envconfig:"USER_SERVICE_URL" default:"http://localhost:8081"`
	TemplateServiceURL     string `envconfig:"TEMPLATE_SERVICE_URL" default:"http://localhost:8082"`
	ReportServiceURL       string `envconfig:"REPORT_SERVICE_URL" default:"http://localhost:8083"`
	DataServiceURL         string `envconfig:"DATA_SERVICE_URL" default:"http://localhost:8084"`
	NotificationServiceURL string `envconfig:"NOTIFICATION_SERVICE_URL" default:"http://localhost:8085"`
	StorageServiceURL      string `envconfig:"STORAGE_SERVICE_URL" default:"http://localhost:8087"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("ошибка обработки конфигурации: %w", err)
	}

	return &cfg, nil
}

func (c *Config) SetupLogger() {
	level, err := logrus.ParseLevel(c.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	if c.Environment == "production" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
}
