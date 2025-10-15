package main

import (
	"api-gateway/internal/config"
	"api-gateway/internal/server"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cfg *config.Config
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "api-gateway",
		Short: "API Gateway",
		Long:  "Единая точка входа для всех микросервисов",
	}

	rootCmd.AddCommand(serveCmd)

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

func serve() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		logrus.Fatal("Ошибка загрузки конфигурации:", err)
	}

	cfg.SetupLogger()

	srv := server.NewServer(cfg)
	if err := srv.Start(); err != nil {
		logrus.Fatal("Ошибка запуска сервера:", err)
	}
}
