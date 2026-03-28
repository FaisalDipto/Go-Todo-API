package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"todo-api/internal/models"

	"golang.org/x/crypto/bcrypt"
)

func (h *TodoHandler) Signup(w http.ResponseWriter, r *http.Request){
	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Invalid Input", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password),bcrypt.DefaultCost)

	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	query := "INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id"
	err = h.Pool.QueryRow(context.Background(), query, u.Username, string(hashedPassword)).Scan(&u.ID)
	if err != nil {
		h.Logger.Printf("SIGNUP ERROR: %v", err)
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	u.Password = ""
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}