package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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
// @host localhost:9090
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

	mux.Handle("/todos", protectedTodos)
	mux.Handle("/todos/", protectedTodoByID)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	mux.HandleFunc("/signup", h.Signup)
	mux.HandleFunc("/login", h.Login)

	handler := middleware.Logging(logger)(mux)
	handler = middleware.Recovery(logger)(handler)

	fmt.Printf("API running on :%v...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}