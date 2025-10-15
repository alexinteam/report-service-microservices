package main

import (
	"fmt"

	"data-service/internal/config"
	"data-service/internal/database"
	"data-service/internal/server"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cfg *config.Config
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "data-service",
		Short: "Data Collection Service",
		Long:  "Микросервис для сбора данных из внешних источников",
	}

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(cleanupCmd)

	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Запуск сервера",
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Выполнение миграций",
	Run: func(cmd *cobra.Command, args []string) {
		if err := migrate(); err != nil {
			logrus.Fatal("Ошибка миграции:", err)
		}
	},
}

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Очистка данных",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cleanup(); err != nil {
			logrus.Fatal("Ошибка очистки данных:", err)
		}
	},
}

func serve() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		logrus.Fatal("Ошибка загрузки конфигурации:", err)
	}

	srv := server.NewServer(cfg)
	if err := srv.Start(); err != nil {
		logrus.Fatal("Ошибка запуска сервера:", err)
	}
}

func migrate() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("ошибка загрузки конфигурации: %w", err)
	}

	return database.MigrateWithConfig(cfg)
}

func cleanup() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("ошибка загрузки конфигурации: %w", err)
	}

	return database.CleanupWithConfig(cfg)
}
