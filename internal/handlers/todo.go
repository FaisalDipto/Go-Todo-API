package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"todo-api/internal/models"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type TodoHandler struct {
	Pool *pgxpool.Pool
	Logger *log.Logger
	JWTSecret string
}

func (h *TodoHandler) HandleTodosById(w http.ResponseWriter, r *http.Request) {
	idString := r.URL.Path[len("/todos/"):]
	id , err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	switch r.Method{
	case http.MethodGet:
		query := "SELECT id, title, status FROM todos WHERE id = $1"
		var t models.Todo
		err := h.Pool.QueryRow(context.Background(), query, id).Scan(&t.Id, &t.Title, &t.Status)
		if err != nil {
			http.Error(w, "Todo not found", 404)
			return
		}
		json.NewEncoder(w).Encode(t)

	case http.MethodDelete:
		query := "DELETE FROM todos WHERE id = $1"
		res, _ := h.Pool.Exec(context.Background(), query, id)
		if res.RowsAffected() == 0 {
			http.Error(w, "Todo not found.", 404)
			return
		}
		w.WriteHeader(http.StatusNotFound)

	case http.MethodPut:
		var updateData struct{
			Status bool `json:"status"`
		}
		query := "UPDATE todos SET status = $1 WHERE id = $2"
		json.NewDecoder(r.Body).Decode(&updateData)
		h.Pool.Exec(context.Background(), query, updateData.Status, id)
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *TodoHandler) HandleTodos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := h.Pool.Query(context.Background(), "SELECT id, title, status FROM todos")
		if err != nil {
		h.Logger.Printf("DATABASE ERROR: %v", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer rows.Close()

	var todos []models.Todo = []models.Todo{}

	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(&t.Id, &t.Title, &t.Status); err != nil {
			continue
		}
		todos = append(todos, t)
	}

	if err = rows.Err(); err != nil {
			h.Logger.Printf("ROW ITERATION ERROR: %v", err)
			http.Error(w, "Internal Server Error", 500)
			return
		}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)

	case http.MethodPost:
		var newTodo models.Todo

		if err := json.NewDecoder(r.Body).Decode(&newTodo); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		if newTodo.Title == "" {
			http.Error(w, "Title is required", http.StatusBadRequest)
			return
		}

		query := "INSERT INTO todos (title) VALUES ($1) RETURNING id"
		err := h.Pool.QueryRow(context.Background(), query, newTodo.Title).Scan(&newTodo.Id)
		if err != nil {
			h.Logger.Printf("POST ERROR: %v", err)
			http.Error(w, "Internal Server Error", 500)
			return
		}


		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newTodo)
	
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *TodoHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds models.User
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// 1. Find the user in the database by their username
	var storedUser models.User
	query := "SELECT id, password_hash FROM users WHERE username = $1"
	err := h.Pool.QueryRow(context.Background(), query, creds.Username).Scan(&storedUser.ID, &storedUser.PasswordHash)
	if err != nil {
		// If the user doesn't exist, we send a generic "Unauthorized"
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// 2. Compare the typed password with the hashed password
	err = bcrypt.CompareHashAndPassword([]byte(storedUser.PasswordHash), []byte(creds.Password))
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// 3. The password is correct! Let's build the JWT Passport.
	// We store the User's ID and an Expiration Time (e.g., 24 hours from now)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": storedUser.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	// 4. "Sign" the passport using our secret key
	tokenString, err := token.SignedString([]byte(h.JWTSecret))
	if err != nil {
		h.Logger.Printf("JWT ERROR: %v", err)
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	// 5. Give the token to the user!
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}