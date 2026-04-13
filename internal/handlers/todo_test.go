package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"todo-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v3"
)

// 1. The Mock implementation
type MockDB struct {
	ShouldFail bool
}

func (m *MockDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.ShouldFail {
		return nil, fmt.Errorf("database connection lost")
	}
	return nil, nil 
}

// We just need these to satisfy the interface for now
func (m *MockDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row { return nil }
func (m *MockDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func TestHandleTodos_DatabaseError(t *testing.T) {
	mock := &MockDB{ShouldFail: true}
		// Create a real logger that writes to nowhere
	testLogger := log.New(io.Discard, "", 0)
	
	h := &TodoHandler{
		Pool:   mock, // We pass the mock instead of a real pool!
		Logger: testLogger,  // We don't need a real logger for this test
	}


	req := httptest.NewRequest(http.MethodGet, "/todos", nil)

	ctx := context.WithValue(req.Context(), "user_id", 1)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	h.HandleTodos(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, but got %d", rr.Code)
	}
}

func TestHandleTodos_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	h := &TodoHandler{
		Pool: mock,
		Logger: log.New(io.Discard, "", 0),
	}

	rows := pgxmock.NewRows([]string{"id", "title", "status"}).AddRow(1, "Finish Stage 7", false).AddRow(2, "Start Swagger", true)

	mock.ExpectQuery("SELECT id, title, status FROM todos WHERE user_id = \\$1").WithArgs(1).WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	ctx := context.WithValue(req.Context(), "user_id", 1)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.HandleTodos(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rr.Code)
	}

	var todos []models.Todo
	json.NewDecoder(rr.Body).Decode(&todos)
	if len(todos) != 2 {
		t.Errorf("Expected 2 todos, got %d", len(todos))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}