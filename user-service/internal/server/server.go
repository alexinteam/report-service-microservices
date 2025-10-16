package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"user-service/internal/config"
	"user-service/internal/database"
	"user-service/internal/handlers"
	"user-service/internal/jwt"
	"user-service/internal/middleware"
	"user-service/internal/repository"
	"user-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

	db, err := database.Connect(s.cfg)
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	userRepo := repository.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	jwtManager := jwt.NewManager(s.cfg.JWTSecret)

	router := s.setupRouter(userService, jwtManager)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.Port),
		Handler: router,
	}

	go func() {
		logrus.Infof("User Service запущен на порту %d", s.cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Ошибка запуска сервера:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Остановка User Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("принудительная остановка сервера: %w", err)
	}

	logrus.Info("User Service остановлен")
	return nil
}

func (s *Server) setupRouter(userService *services.UserService, jwtManager *jwt.Manager) *gin.Engine {
	router := gin.Default()

	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	userHandler := handlers.NewUserHandler(userService)

	s.setupRoutes(router, userHandler, jwtManager)

	return router
}

func (s *Server) setupRoutes(router *gin.Engine, userHandler *handlers.UserHandler, jwtManager *jwt.Manager) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "user-service",
			"timestamp": time.Now().Unix(),
		})
	})

	api := router.Group("/api/v1")
	{
		public := api.Group("/users")
		{
			public.POST("/register", userHandler.Register)
			public.POST("/login", userHandler.Login)
		}

		protected := api.Group("/users")
		{
			protected.GET("/profile", userHandler.GetProfile)
			protected.PUT("/profile", userHandler.UpdateProfile)
			protected.GET("/", userHandler.GetUsers)
			protected.GET("/:id", userHandler.GetUser)
			protected.PUT("/:id", userHandler.UpdateUser)
			protected.DELETE("/:id", userHandler.DeleteUser)
		}
	}
}

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
