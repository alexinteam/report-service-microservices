package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"template-service/internal/config"
	"template-service/internal/database"
	"template-service/internal/handlers"
	"template-service/internal/jwt"
	"template-service/internal/metrics"
	"template-service/internal/middleware"
	"template-service/internal/repository"
	"template-service/internal/services"

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
		logrus.Infof("Template Service запущен на порту %d", s.cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Ошибка запуска сервера:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Остановка Template Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("принудительная остановка сервера: %w", err)
	}

	logrus.Info("Template Service остановлен")
	return nil
}

func (s *Server) setupRouter(db *gorm.DB, jwtManager *jwt.Manager) *gin.Engine {
	router := gin.Default()

	// Инициализация метрик
	serviceMetrics := metrics.NewMetrics("template-service")
	serviceMetrics.SetupMetricsEndpoint(router, "template-service")

	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	templateRepo := repository.NewTemplateRepository(db)
	categoryRepo := repository.NewTemplateCategoryRepository(db)
	variableRepo := repository.NewTemplateVariableRepository(db)

	templateService := services.NewTemplateService(templateRepo)
	categoryService := services.NewTemplateCategoryService(categoryRepo)
	variableService := services.NewTemplateVariableService(variableRepo)

	templateHandler := handlers.NewTemplateHandler(templateService)
	categoryHandler := handlers.NewTemplateCategoryHandler(categoryService)
	variableHandler := handlers.NewTemplateVariableHandler(variableService)

	s.setupRoutes(router, templateHandler, categoryHandler, variableHandler, jwtManager)

	return router
}

func (s *Server) setupRoutes(router *gin.Engine, templateHandler *handlers.TemplateHandler, categoryHandler *handlers.TemplateCategoryHandler, variableHandler *handlers.TemplateVariableHandler, jwtManager *jwt.Manager) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "template-service",
			"version": "1.0.0",
		})
	})

	api := router.Group("/api/v1")
	{
		templates := api.Group("/templates")
		templates.Use(middleware.Auth(jwtManager))
		{
			templates.POST("/", templateHandler.CreateTemplate)
			templates.GET("/", templateHandler.GetTemplates)
			templates.GET("/:id", templateHandler.GetTemplate)
			templates.PUT("/:id", templateHandler.UpdateTemplate)
			templates.DELETE("/:id", templateHandler.DeleteTemplate)
			templates.GET("/search", templateHandler.SearchTemplates)
			templates.POST("/render", templateHandler.RenderTemplate)
		}

		categories := api.Group("/categories")
		categories.Use(middleware.Auth(jwtManager))
		{
			categories.POST("/", categoryHandler.CreateCategory)
			categories.GET("/", categoryHandler.GetCategories)
			categories.GET("/:id", categoryHandler.GetCategory)
			categories.PUT("/:id", categoryHandler.UpdateCategory)
			categories.DELETE("/:id", categoryHandler.DeleteCategory)
		}

		variables := api.Group("/variables")
		variables.Use(middleware.Auth(jwtManager))
		{
			variables.POST("/", variableHandler.CreateVariable)
			variables.GET("/", variableHandler.GetVariables)
			variables.GET("/:id", variableHandler.GetVariable)
			variables.PUT("/:id", variableHandler.UpdateVariable)
			variables.DELETE("/:id", variableHandler.DeleteVariable)
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
