package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification-service/internal/config"
	"notification-service/internal/database"
	"notification-service/internal/handlers"
	"notification-service/internal/jwt"
	"notification-service/internal/metrics"
	"notification-service/internal/middleware"
	"notification-service/internal/models"
	"notification-service/internal/repository"
	"notification-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

// Server представляет HTTP сервер
type Server struct {
	cfg                 *config.Config
	notificationService *services.NotificationService
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
	db, err := database.Connect(s.cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	// Инициализация зависимостей
	jwtManager := jwt.NewManager(s.cfg.JWTSecret)
	metricsManager := metrics.NewMetrics("notification-service")

	// Создание роутера
	router := s.setupRouter(db, jwtManager, metricsManager)

	// Запуск потребителя RabbitMQ (если URL указан)
	if s.cfg.RabbitMQURL != "" {
		go s.startRabbitConsumer()
	}

	// Создание HTTP сервера
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.Port),
		Handler: router,
	}

	// Запуск сервера в горутине
	go func() {
		logrus.Infof("Notification Service запущен на порту %d", s.cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Ошибка запуска сервера:", err)
		}
	}()

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Остановка Notification Service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("принудительная остановка сервера: %w", err)
	}

	logrus.Info("Notification Service остановлен")
	return nil
}

// startRabbitConsumer запускает consumer для событий report.completed
func (s *Server) startRabbitConsumer() {
	amqpURL := s.cfg.RabbitMQURL
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		logrus.WithError(err).Warn("Не удалось подключиться к RabbitMQ")
		return
	}

	ch, err := conn.Channel()
	if err != nil {
		logrus.WithError(err).Warn("Не удалось открыть канал RabbitMQ")
		return
	}

	// Объявляем exchange и очередь
	if err := ch.ExchangeDeclare("events", "topic", true, false, false, false, nil); err != nil {
		logrus.WithError(err).Warn("Не удалось объявить exchange events")
		return
	}

	q, err := ch.QueueDeclare("notification-report-completed", true, false, false, false, nil)
	if err != nil {
		logrus.WithError(err).Warn("Не удалось объявить очередь notification-report-completed")
		return
	}

	if err := ch.QueueBind(q.Name, "report.completed", "events", false, nil); err != nil {
		logrus.WithError(err).Warn("Не удалось привязать очередь к ключу report.completed")
		return
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		logrus.WithError(err).Warn("Не удалось создать consumer RabbitMQ")
		return
	}

	logrus.Info("RabbitMQ consumer notification-service запущен (report.completed)")

	// Простой обработчик: создаем запись уведомления на основе события
	go func() {
		for m := range msgs {
			logrus.WithField("routing_key", m.RoutingKey).Info("Получено событие из RabbitMQ")
			var evt struct {
				Type string                 `json:"type"`
				Data map[string]interface{} `json:"data"`
			}
			if err := json.Unmarshal(m.Body, &evt); err != nil {
				logrus.WithError(err).Warn("Не удалось распарсить событие")
				continue
			}
			if evt.Type != "report.completed" {
				continue
			}
			userID := ""
			if v, ok := evt.Data["user_id"].(string); ok {
				userID = v
			}
			reportID := ""
			if v, ok := evt.Data["report_id"].(string); ok {
				reportID = v
			}
			req := &models.NotificationCreateRequest{
				TemplateID: 1,
				Recipient:  userID,
				Type:       "report_ready",
				Data: map[string]interface{}{
					"report_id": reportID,
				},
			}
			if s.notificationService == nil {
				logrus.Warn("notificationService не инициализирован")
				continue
			}
			if _, err := s.notificationService.SendNotification(req); err != nil {
				logrus.WithError(err).Warn("Не удалось создать уведомление из события")
			} else {
				logrus.Info("Уведомление создано из события report.completed")
			}
		}
	}()
}

// setupRouter настраивает маршруты и middleware
func (s *Server) setupRouter(db *gorm.DB, jwtManager *jwt.Manager, metricsManager *metrics.Metrics) *gin.Engine {
	router := gin.Default()

	// Инициализация метрик
	metricsManager.SetupMetricsEndpoint(router, "notification-service")

	// Middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// Инициализация репозиториев
	templateRepo := repository.NewNotificationTemplateRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	channelRepo := repository.NewNotificationChannelRepository(db)

	// Инициализация сервисов
	templateService := services.NewNotificationTemplateService(templateRepo)
	notificationService := services.NewNotificationService(notificationRepo, templateRepo)
	channelService := services.NewNotificationChannelService(channelRepo)
	s.notificationService = notificationService

	// Инициализация обработчиков
	templateHandler := handlers.NewNotificationTemplateHandler(templateService, metricsManager)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	channelHandler := handlers.NewNotificationChannelHandler(channelService)

	// Настройка маршрутов
	s.setupRoutes(router, templateHandler, notificationHandler, channelHandler, jwtManager)

	return router
}

// setupRoutes настраивает маршруты API
func (s *Server) setupRoutes(router *gin.Engine, templateHandler *handlers.NotificationTemplateHandler, notificationHandler *handlers.NotificationHandler, channelHandler *handlers.NotificationChannelHandler, jwtManager *jwt.Manager) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "notification-service",
			"version": "1.0.0",
		})
	})

	api := router.Group("/api/v1")
	{
		// Шаблоны уведомлений
		templates := api.Group("/templates")
		// templates.Use(middleware.AuthMiddleware(jwtManager))
		{
			templates.POST("/", templateHandler.CreateTemplate)
			templates.GET("/", templateHandler.GetTemplates)
			templates.GET("/:id", templateHandler.GetTemplate)
			templates.PUT("/:id", templateHandler.UpdateTemplate)
			templates.DELETE("/:id", templateHandler.DeleteTemplate)
		}

		// Уведомления
		notifications := api.Group("/notifications")
		// notifications.Use(middleware.AuthMiddleware(jwtManager))
		{
			notifications.POST("/send", notificationHandler.SendNotification)
			notifications.GET("/", notificationHandler.GetNotifications)
			notifications.GET("/:id", notificationHandler.GetNotification)
			notifications.PUT("/:id/status", notificationHandler.UpdateNotificationStatus)
		}

		// Каналы уведомлений
		channels := api.Group("/channels")
		// channels.Use(middleware.AuthMiddleware(jwtManager))
		{
			channels.POST("/", channelHandler.CreateChannel)
			channels.GET("/", channelHandler.GetChannels)
			channels.GET("/:id", channelHandler.GetChannel)
			channels.PUT("/:id", channelHandler.UpdateChannel)
			channels.DELETE("/:id", channelHandler.DeleteChannel)
		}
	}
}

// migrate выполняет миграции базы данных
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
