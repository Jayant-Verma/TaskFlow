package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"taskflow-api/internal/database"
	"taskflow-api/internal/handlers"
	"taskflow-api/internal/middleware"

	_ "taskflow-api/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           TaskFlow API
// @version         1.0
// @description     A minimal, robust task management REST API.
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     Type "Bearer " followed by your token.
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	db, err := database.ConnectAndMigrate()
	if err != nil {
		logger.Error("failed to connect and migrate database", "err", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connected and migrated successfully")

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	authH := &handlers.AuthHandler{DB: db, JWTSecret: jwtSecret}
	projH := &handlers.ProjectHandler{DB: db}
	taskH := &handlers.TaskHandler{DB: db}

	mux := http.NewServeMux()

	// Swagger UI Route
	mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)

	// Public Routes
	mux.HandleFunc("POST /auth/register", authH.Register)
	mux.HandleFunc("POST /auth/login", authH.Login)

	// Protected Routes
	protected := http.NewServeMux()
	protected.HandleFunc("GET /projects", projH.List)
	protected.HandleFunc("POST /projects", projH.Create)
	protected.HandleFunc("GET /projects/{id}", projH.Get)
	protected.HandleFunc("PATCH /projects/{id}", projH.Update)
	protected.HandleFunc("DELETE /projects/{id}", projH.Delete)
	protected.HandleFunc("GET /projects/{id}/stats", projH.Stats)

	protected.HandleFunc("GET /projects/{id}/tasks", taskH.List)
	protected.HandleFunc("POST /projects/{id}/tasks", taskH.Create)
	protected.HandleFunc("PATCH /tasks/{id}", taskH.Update)
	protected.HandleFunc("DELETE /tasks/{id}", taskH.Delete)

	mux.Handle("/", middleware.Auth(jwtSecret, protected))

	// Global Middleware Chain: Logger -> CORS -> Mux
	handler := middleware.Logger(logger, middleware.CORS(mux))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		logger.Info("starting server", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "err", err)
	}
}
