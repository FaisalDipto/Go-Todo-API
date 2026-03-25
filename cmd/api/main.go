package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"todo-api/internal/database"
	"todo-api/internal/handlers"
	"todo-api/internal/middleware"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system default variables")
	}

	dbUrl := os.Getenv("DB_URL")
	port := os.Getenv("PORT")
	
	dbPool, err := database.Connect(dbUrl)
	if err != nil {
		log.Fatal("Could not connect to database:", err)
	}
	defer dbPool.Close()

	logger := log.New(os.Stdout, "[TODO-API] ", log.LstdFlags)

	h := &handlers.TodoHandler{
		Pool: dbPool,
		Logger: logger,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/todos", h.HandleTodos)
	mux.HandleFunc("/todos/", h.HandleTodosById)

	handler := middleware.Logging(logger)(mux)
	handler = middleware.Recovery(logger)(handler)

	fmt.Printf("API running on :%v...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}