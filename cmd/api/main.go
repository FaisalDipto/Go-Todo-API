package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todo-api/internal/database"
	"todo-api/internal/handlers"
	"todo-api/internal/middleware"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"

	_ "todo-api/docs" // This is vital!

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Todo API
// @version 1.0
// @description This is a professional Todo list server.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error finding .env file, using system default variables")
	}

	dbUrl := os.Getenv("DB_URL")
	port := os.Getenv("PORT")
	jwtSecret := os.Getenv("JWT_SECRET")

	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is missing from .env file")
	}
	
	dbPool, err := database.Connect(dbUrl)
	if err != nil {
		log.Fatal("Error occured while connecting to database:", err)
	}
	defer dbPool.Close()

	logger := log.New(os.Stdout, "[TODO-API] ", log.LstdFlags)

	v := validator.New()

	h := &handlers.TodoHandler{
		Pool: dbPool,
		Logger: logger,
		JWTSecret: jwtSecret,
		Validator: v,
	}

	mux := http.NewServeMux()

	authMiddleware := middleware.Auth(jwtSecret)

	protectedTodos := authMiddleware(http.HandlerFunc(h.HandleTodos))
	protectedTodoByID := authMiddleware(http.HandlerFunc(h.HandleTodosById))
	protectedTrash := authMiddleware(http.HandlerFunc(h.HandleTrash))
	protectedRestore := authMiddleware(http.HandlerFunc(h.RestoreTrash))

	mux.Handle("/todos", protectedTodos)
	mux.Handle("/todos/", protectedTodoByID)
	mux.Handle("/todos/trash", protectedTrash)
	mux.Handle("/todos/restore/", protectedRestore)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	mux.HandleFunc("/signup", h.Signup)
	mux.HandleFunc("/login", h.Login)

	limiter := middleware.NewIPRateLimiter(2, 20)

	handlerWithRateLimit := limiter.RateLimitMiddleware(mux)

	handler := middleware.Logging(logger)(handlerWithRateLimit)
	handler = middleware.Recovery(logger)(handler)

	srv := &http.Server{
		Addr: ":" + port,
		Handler: handler,
	}

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func ()  {
		fmt.Printf("API running on :%v...\n", port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatalf("Server crashed: %v\n", err)
		}
	}()

	<-quit

	logger.Println("Stop signal received. Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Println("Closing databae connection pool...")
	dbPool.Close()

	logger.Println("Server exited safely")
}