package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"todo-api/internal/models"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	// "github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type DBInterface interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type TodoHandler struct {
	Pool DBInterface
	Logger *log.Logger
	JWTSecret string
	Validator *validator.Validate
}

// HandleTodosById godoc
// @Summary Get, Update or Delete a todo
// @Description Fetch, update status, or remove a specific todo by ID
// @Tags todos
// @Accept  json
// @Produce  json
// @Param id path int true "Todo ID"
// @Param status body object false "Update status (for PUT only) { 'status': true }"
// @Security BearerAuth
// @Success 200 {object} models.Todo
// @Success 204 "No Content"
// @Router /todos/{id} [get]
// @Router /todos/{id} [put]
// @Router /todos/{id} [delete]
func (h *TodoHandler) HandleTodosById(w http.ResponseWriter, r *http.Request){
	idString := r.URL.Path[len("/todos/"):]
	id, err := strconv.Atoi(idString)
	if err != nil{
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		h.Logger.Println("Critical Error: user_id not found in context")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	switch r.Method{
	case http.MethodGet:
		var t models.Todo
		query := "SELECT id, title, status FROM todos WHERE id = $1 AND user_id = $2"
		err := h.Pool.QueryRow(context.Background(),query, id, userID).Scan(&t.Id, &t.Title, &t.Status)
		if err != nil {
			http.Error(w, "Todo not found", 404)
			return
		}
		json.NewEncoder(w).Encode(t)

	case http.MethodDelete:
		query := "DELETE FROM todos WHERE id = $1 AND user_id = $2"
		res, _ := h.Pool.Exec(context.Background(), query, id, userID)

		if res.RowsAffected() == 0{
			http.Error(w, "Todo not found.", 404)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	case http.MethodPut:
		query := "UPDATE todos SET status = $1 WHERE id = $2 AND user_id = $3"
		var updateData struct{
			Status bool		`json:"status"`
		}

		json.NewDecoder(r.Body).Decode(&updateData)
		h.Pool.Exec(context.Background(), query, updateData.Status, id, userID)
		w.WriteHeader(http.StatusNoContent)
	}
}

// HandleTodos godoc
// @Summary List or create todos
// @Description Get all todos for the current user or create a new todo
// @Tags todos
// @Accept  json
// @Produce  json
// @Param todo body models.Todo false "Todo object (for POST only)"
// @Security BearerAuth
// @Success 200 {array} models.Todo
// @Success 201 {object} models.Todo
// @Router /todos [get]
// @Router /todos [post]
func (h *TodoHandler) HandleTodos(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		h.Logger.Println("Critical Error: user_id not found in context")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	switch r.Method {
	case http.MethodGet:
		rows, err := h.Pool.Query(context.Background(), "SELECT id, title, status FROM todos WHERE user_id = $1", userID)
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

		if err := h.Validator.Struct(newTodo); err != nil {
			formattedErrors := formatValidationErrors(err)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "validation failed",
				"details": formattedErrors,
			})
			return
		}

		query := "INSERT INTO todos (title, user_id) VALUES ($1, $2) RETURNING id"
		err := h.Pool.QueryRow(context.Background(), query, newTodo.Title, userID).Scan(&newTodo.Id)
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

// Login godoc
// @Summary Authenticate a user
// @Tags auth
// @Accept  json
// @Produce  json
// @Param credentials body models.User true "Login Credentials"
// @Success 200 {object} map[string]string
// @Router /login [post]
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

func formatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	if vErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range vErrors {
			field := strings.ToLower(e.Field())
			switch e.Tag() {
			case "required":
				errors[field] = "this field is required"
			case "min":
				errors[field]	= "too short"
			case "max":
				errors[field] = "too long"
			case "alphanum":
				errors[field] = "must contain only letters and numbers"
			default:
				errors[field] = "invalid value"
			}
		}
	}
	return errors
}