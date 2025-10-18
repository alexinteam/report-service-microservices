package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"storage-service/internal/config"
	"storage-service/internal/database"
	"storage-service/internal/handlers"
	"storage-service/internal/jwt"
	"storage-service/internal/metrics"
	"storage-service/internal/middleware"
	"storage-service/internal/repository"
	"storage-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Server struct {
	cfg *config.Config
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg: cfg,
	}
}

func (s *Server) Start() error {
	s.cfg.SetupLogger()

	if s.cfg.AutoMigrate {
		if err := s.migrate(); err != nil {
			return fmt.Errorf("ошибка миграции: %w", err)
		}
	}

	db, err := database.Connect(s.cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	jwtManager := jwt.NewManager(s.cfg.JWTSecret)

	router := s.setupRouter(db, jwtManager)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.Port),
		Handler: router,
	}

	go func() {
		logrus.Infof("Storage Service запущен на порту %d", s.cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Ошибка запуска сервера:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Остановка Storage Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("принудительная остановка сервера: %w", err)
	}

	logrus.Info("Storage Service остановлен")
	return nil
}

func (s *Server) setupRouter(db *gorm.DB, jwtManager *jwt.Manager) *gin.Engine {
	router := gin.Default()

	// Инициализация метрик
	serviceMetrics := metrics.NewMetrics("storage-service")
	serviceMetrics.SetupMetricsEndpoint(router, "storage-service")

	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	fileRepo := repository.NewFileRepository(db)

	fileService := services.NewFileService(fileRepo, s.cfg.StoragePath)

	fileHandler := handlers.NewFileHandler(fileService)

	s.setupRoutes(router, fileHandler, jwtManager)

	return router
}

func (s *Server) setupRoutes(router *gin.Engine, fileHandler *handlers.FileHandler, jwtManager *jwt.Manager) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "storage-service",
			"version": "1.0.0",
		})
	})

	api := router.Group("/api/v1")
	{
		files := api.Group("/files")
		{
			files.POST("/upload", fileHandler.UploadFile)
			files.GET("/", fileHandler.GetFiles)
			files.GET("/:id", fileHandler.GetFile)
			files.GET("/:id/download", fileHandler.DownloadFile)
			files.GET("/:id/content", fileHandler.GetFileContent)
			files.PUT("/:id", fileHandler.UpdateFile)
			files.DELETE("/:id", fileHandler.DeleteFile)
			files.GET("/hash/:hash", fileHandler.GetFileByHash)
			files.GET("/search", fileHandler.SearchFiles)
		}

		stats := api.Group("/stats")
		{
			stats.GET("/storage", fileHandler.GetStorageStats)
		}
	}
}

func (s *Server) migrate() error {
	_, err := database.Connect(s.cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	if err := database.Migrate(); err != nil {
		return fmt.Errorf("ошибка миграции: %w", err)
	}

	if s.cfg.SeedData {
		if err := database.SeedData(); err != nil {
			return fmt.Errorf("ошибка заполнения тестовыми данными: %w", err)
		}
	}

	logrus.Info("Миграции выполнены успешно")
	return nil
}
