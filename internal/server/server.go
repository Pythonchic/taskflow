package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"taskflow/internal/email"
	"taskflow/internal/handlers"
	"taskflow/internal/middleware"
	"taskflow/internal/paths"
	"taskflow/internal/config"
	"taskflow/internal/repository"
	prettyprint "taskflow/pkg/pretty_print"
	"taskflow/internal/database"
)

type Server struct {
	router       *gin.Engine
	config       *config.ServerConfig
	appConfig    *config.AppConfig
	emailService *email.Service
	testEmail    string
	http         *http.Server
}

func New(cfg *config.AppConfig, emailService *email.Service) *Server {
	// Устанавливаем режим Gin в зависимости от окружения
	if cfg.IsProd() {
		gin.SetMode(gin.ReleaseMode)
		prettyprint.SetDebugMode(false)
	} else {
		gin.SetMode(gin.DebugMode)
		prettyprint.SetDebugMode(true)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	// Добавляем middleware безопасности
	allowedOrigin := getEnv("ALLOWED_ORIGIN", "http://localhost:8080")
	r.Use(middleware.CORSMiddleware(cfg.IsProd(), allowedOrigin))
	r.Use(middleware.SecurityHeaders(cfg.IsProd()))

	// Rate limiting (10 запросов в секунду, burst 20)
	r.Use(middleware.RateLimiter(rate.Limit(10), 20))

	return &Server{
		router:       r,
		config:       &cfg.Server,
		appConfig:    cfg,
		emailService: emailService,
		testEmail:    cfg.Email.TestEmail,
	}
}

// getEnv читает переменную окружения
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Setup статика и маршруты
func (s *Server) Setup() error {
	// Статика
	err := s.setupStaticFiles()
	if err != nil {
		return err
	}

	// Маршруты
	if err := s.setupRoutes(); err != nil {
		return err
	}

	return nil
}

func (s *Server) setupStaticFiles() error {
	webPath, err := paths.GetWebPath()
	if err != nil {
		return err
	}

	// Загружаем шаблоны
	htmlPattern := filepath.Join(webPath, "html/*")
	prettyprint.Info("Loading HTML templates from: %s", htmlPattern)
	s.router.LoadHTMLGlob(htmlPattern)

	// Статические файлы
	s.router.Static("/css", filepath.Join(webPath, "css"))
	s.router.Static("/js", filepath.Join(webPath, "js"))
	s.router.StaticFile("/favicon.ico", filepath.Join(webPath, "favicon.ico"))

	prettyprint.Success("Static files configured successfully")
	return nil
}

func (s *Server) setupRoutes() error {
	userRepo := repository.NewUserRepository()
	taskRepo := repository.NewTaskRepository()
	authHandler := handlers.NewAuthHandler(userRepo, s.emailService, s.emailService.TestEmail)
	taskHandler := handlers.NewTaskHandler(userRepo, taskRepo)

	// Страницы
	s.router.GET("/", handlers.MainPage)
	s.router.GET("/login", handlers.LoginPage)
	s.router.GET("/tasks", middleware.AuthMiddleware(), taskHandler.TasksPage)

	// API группа
	api := s.router.Group("/api/v1")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
		api.GET("/logout", authHandler.Logout)
		api.POST("/verify", authHandler.Verify)
		api.POST("/resend-code", authHandler.ResendCode)

		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/tasks", taskHandler.GetTasks)
			protected.POST("/tasks", taskHandler.CreateTask)
			protected.PATCH("/tasks/:id", taskHandler.UpdateTask)
			protected.PUT("/tasks/:id/toggle", taskHandler.ToggleTask)
			protected.DELETE("/tasks/:id", taskHandler.DeleteTask)
		}
	}

	prettyprint.Success("Routes configured")
	return nil
}

func (s *Server) Run() error {
	if err := database.Init(); err != nil {
		prettyprint.Fatal("Failed to connect to database: %v", err)
	}

	s.http = &http.Server{
		Addr:         s.config.Port,
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	// Канал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Запуск сервера в горутине
	go func() {
		prettyprint.Info("Starting server on http://localhost%s", s.config.Port)
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			prettyprint.Fatal("Failed to start server: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	sig := <-quit
	prettyprint.Warn("Received signal: %v", sig)
	prettyprint.Progress("Shutting down server")

	// Даем время на завершение запросов
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.http.Shutdown(ctx); err != nil {
		return err
	}

	prettyprint.Success("Server exited gracefully")
	return nil
}
