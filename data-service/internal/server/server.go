package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"data-service/internal/config"
	"data-service/internal/database"
	"data-service/internal/handlers"
	"data-service/internal/jwt"
	"data-service/internal/middleware"
	"data-service/internal/repository"
	"data-service/internal/services"

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
		logrus.Infof("Data Service запущен на порту %d", s.cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Ошибка запуска сервера:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Остановка Data Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("принудительная остановка сервера: %w", err)
	}

	logrus.Info("Data Service остановлен")
	return nil
}

func (s *Server) setupRouter(db *gorm.DB, jwtManager *jwt.Manager) *gin.Engine {
	router := gin.Default()

	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	dataSourceRepo := repository.NewDataSourceRepository(db)
	dataCollectionRepo := repository.NewDataCollectionRepository(db)
	dataRecordRepo := repository.NewDataRecordRepository(db)

	dataSourceService := services.NewDataSourceService(dataSourceRepo)
	dataCollectionService := services.NewDataCollectionService(dataCollectionRepo)
	collectDataService := services.NewCollectDataService(dataRecordRepo)

	dataSourceHandler := handlers.NewDataSourceHandler(dataSourceService)
	dataCollectionHandler := handlers.NewDataCollectionHandler(dataCollectionService)
	collectDataHandler := handlers.NewCollectDataHandler(collectDataService)

	s.setupRoutes(router, dataSourceHandler, dataCollectionHandler, collectDataHandler, jwtManager)

	return router
}

func (s *Server) setupRoutes(router *gin.Engine, dataSourceHandler *handlers.DataSourceHandler, dataCollectionHandler *handlers.DataCollectionHandler, collectDataHandler *handlers.CollectDataHandler, jwtManager *jwt.Manager) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "data-service",
			"version": "1.0.0",
		})
	})

	api := router.Group("/api/v1")
	{
		dataSources := api.Group("/data-sources")
		{
			dataSources.POST("/", dataSourceHandler.CreateDataSource)
			dataSources.GET("/", dataSourceHandler.GetDataSources)
			dataSources.GET("/:id", dataSourceHandler.GetDataSource)
			dataSources.PUT("/:id", dataSourceHandler.UpdateDataSource)
			dataSources.DELETE("/:id", dataSourceHandler.DeleteDataSource)
		}

		dataCollections := api.Group("/data-collections")
		{
			dataCollections.POST("/", dataCollectionHandler.CreateDataCollection)
			dataCollections.GET("/", dataCollectionHandler.GetDataCollections)
			dataCollections.GET("/:id", dataCollectionHandler.GetDataCollection)
			dataCollections.PUT("/:id", dataCollectionHandler.UpdateDataCollection)
			dataCollections.DELETE("/:id", dataCollectionHandler.DeleteDataCollection)
		}

		collect := api.Group("/collect")
		{
			collect.POST("/", collectDataHandler.CollectData)
			collect.GET("/records", collectDataHandler.GetDataRecords)
			collect.GET("/records/:id", collectDataHandler.GetDataRecord)
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
