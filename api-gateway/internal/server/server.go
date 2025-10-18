package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-gateway/internal/config"
	"api-gateway/internal/handlers"
	"api-gateway/internal/jwt"
	"api-gateway/internal/metrics"
	"api-gateway/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Server struct {
	config  *config.Config
	handler *handlers.GatewayHandler
	router  *gin.Engine
	metrics *metrics.Metrics
}

func NewServer(cfg *config.Config) *Server {
	gatewayHandler := handlers.NewGatewayHandler(cfg)
	jwtManager := jwt.NewManager(cfg.JWTSecret)

	// Инициализация метрик
	serviceMetrics := metrics.NewMetrics("api-gateway")

	router := gin.Default()

	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.Metrics())
	router.Use(middleware.Timeout(30 * time.Second))

	// Настройка метрик
	serviceMetrics.SetupMetricsEndpoint(router, "api-gateway")

	setupRoutes(router, gatewayHandler, jwtManager)

	return &Server{
		config:  cfg,
		handler: gatewayHandler,
		router:  router,
		metrics: serviceMetrics,
	}
}

func (s *Server) Start() error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: s.router,
	}

	go func() {
		logrus.Infof("API Gateway запущен на порту %d", s.config.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Ошибка запуска сервера:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Остановка API Gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatal("Принудительная остановка сервера:", err)
	}

	logrus.Info("API Gateway остановлен")
	return nil
}

func setupRoutes(router *gin.Engine, gatewayHandler *handlers.GatewayHandler, jwtManager *jwt.Manager) {
	router.GET("/health", gatewayHandler.Health)

	// API Gateway маршруты
	api := router.Group("/api/v1")
	{
		public := api.Group("/public")
		{
			public.GET("/health", gatewayHandler.Health)
		}

		protected := api.Group("/")
		protected.Use(middleware.Auth(jwtManager))
		{
			// Проксирование запросов к микросервисам
			protected.Any("/users/*path", gatewayHandler.ProxyToUserService)
			protected.Any("/templates/*path", gatewayHandler.ProxyToTemplateService)
			protected.Any("/reports/*path", gatewayHandler.ProxyToReportService)
			protected.Any("/data-sources/*path", gatewayHandler.ProxyToDataService)
			protected.Any("/data/*path", gatewayHandler.ProxyToDataService)
			protected.Any("/notifications/*path", gatewayHandler.ProxyToNotificationService)
			protected.Any("/storage/*path", gatewayHandler.ProxyToStorageService)
		}
	}
}
