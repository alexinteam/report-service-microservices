package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"report-service/internal/config"
	"report-service/internal/database"
	"report-service/internal/events"
	"report-service/internal/handlers"
	"report-service/internal/jwt"
	"report-service/internal/metrics"
	"report-service/internal/middleware"
	"report-service/internal/repository"
	"report-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Server представляет HTTP сервер
type Server struct {
	cfg *config.Config
}

// NewServer создает новый экземпляр сервера
func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg: cfg,
	}
}

// Start запускает HTTP сервер
func (s *Server) Start() error {
	// Настройка логгера
	s.cfg.SetupLogger()

	// Автомиграция если включена
	if s.cfg.AutoMigrate {
		if err := s.migrate(); err != nil {
			return fmt.Errorf("ошибка миграции: %w", err)
		}
	}

	// Подключение к базе данных
	db, err := database.Connect(s.cfg)
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	// Инициализация зависимостей
	reportRepo := repository.NewReportRepository(db)
	reportService := services.NewReportService(reportRepo)
	jwtManager := jwt.NewManager(s.cfg.JWTSecret)
	metricsManager := metrics.NewMetrics("report-service")

	// Инициализация Saga компонентов
	sagaStateStore := events.NewSagaStateStore(db)
	outboxManager := events.NewOutboxManager(db)

	// Создание RabbitMQ publisher (если URL указан)
	var eventPublisher events.EventPublisher
	if s.cfg.RabbitMQURL != "" {
		rabbitPublisher, err := events.NewRabbitMQPublisher(s.cfg.RabbitMQURL)
		if err != nil {
			logrus.WithError(err).Warn("Не удалось подключиться к RabbitMQ, используем локальную публикацию")
			eventPublisher = &events.LocalEventPublisher{}
		} else {
			eventPublisher = rabbitPublisher
			defer rabbitPublisher.Close()
		}
	} else {
		eventPublisher = &events.LocalEventPublisher{}
	}

	// Создание идемпотентного Saga Coordinator
	sagaStepHandler := handlers.NewSagaStepHandler(reportService)
	sagaCoordinator := events.NewIdempotentSagaCoordinator(eventPublisher, sagaStateStore, sagaStepHandler, metricsManager)

	// Запуск Outbox Publisher для надежной публикации событий
	if outboxManager != nil {
		outboxPublisher := events.NewOutboxPublisher(outboxManager, eventPublisher)
		go outboxPublisher.StartPublishing(context.Background(), 5*time.Second, 10)
	}

	// Миграция Saga таблиц
	if err := sagaStateStore.MigrateSagaTables(context.Background()); err != nil {
		logrus.WithError(err).Error("Ошибка миграции Saga таблиц")
	}
	if outboxManager != nil {
		if err := outboxManager.MigrateOutboxTable(context.Background()); err != nil {
			logrus.WithError(err).Error("Ошибка миграции Outbox таблицы")
		}
	}

	// Создание роутера
	router := s.setupRouter(reportService, jwtManager, sagaCoordinator, sagaStateStore, metricsManager)

	// Создание HTTP сервера
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.Port),
		Handler: router,
	}

	// Запуск сервера в горутине
	go func() {
		logrus.Infof("Report Service запущен на порту %d", s.cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Ошибка запуска сервера:", err)
		}
	}()

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Остановка Report Service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("принудительная остановка сервера: %w", err)
	}

	logrus.Info("Report Service остановлен")
	return nil
}

// setupRouter настраивает маршруты и middleware
func (s *Server) setupRouter(reportService *services.ReportService, jwtManager *jwt.Manager, sagaCoordinator *events.IdempotentSagaCoordinator, sagaStateStore *events.SagaStateStore, metricsManager *metrics.Metrics) *gin.Engine {
	router := gin.Default()

	// Инициализация метрик
	metricsManager.SetupMetricsEndpoint(router, "report-service")

	// Middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// Инициализация обработчиков
	reportHandler := handlers.NewReportHandler(reportService, sagaCoordinator, metricsManager)
	sagaHandler := handlers.NewSagaHandler(sagaCoordinator, sagaStateStore)

	// Настройка маршрутов
	s.setupRoutes(router, reportHandler, sagaHandler, jwtManager)

	return router
}

// setupRoutes настраивает маршруты API
func (s *Server) setupRoutes(router *gin.Engine, reportHandler *handlers.ReportHandler, sagaHandler *handlers.SagaHandler, jwtManager *jwt.Manager) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "report-service",
			"timestamp": time.Now().Unix(),
		})
	})

	api := router.Group("/api/v1")
	{
		// Защищенные маршруты (требуют аутентификации)
		protected := api.Group("/reports")
		protected.Use(middleware.Auth(jwtManager))
		{
			protected.POST("/", reportHandler.CreateReport)
			protected.GET("/", reportHandler.GetReports)
			protected.GET("/:id", reportHandler.GetReport)
			protected.GET("/:id/status", reportHandler.GetReportStatus)
			protected.PUT("/:id", reportHandler.UpdateReport)
			protected.DELETE("/:id", reportHandler.DeleteReport)
			protected.POST("/generate", reportHandler.GenerateReport)
			protected.GET("/:id/download", reportHandler.DownloadReport)
		}

		// Saga маршруты
		saga := api.Group("/sagas")
		saga.Use(middleware.Auth(jwtManager))
		{
			saga.POST("/reports", sagaHandler.CreateReportSaga)
			saga.GET("/:id", sagaHandler.GetSagaStatus)
			saga.GET("/:id/progress", sagaHandler.GetSagaProgress)
			saga.POST("/:id/retry", sagaHandler.RetrySaga)
			saga.DELETE("/:id", sagaHandler.CancelSaga)
			saga.POST("/:id/force-complete", sagaHandler.ForceCompleteSaga)
			saga.GET("/", sagaHandler.ListSagas)
		}
	}
}

// migrate выполняет миграции базы данных
func (s *Server) migrate() error {
	_, err := database.Connect(s.cfg)
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
