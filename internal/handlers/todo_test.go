package handlers

import (
	"context"
	"strings"
	// "encoding/json"
	"errors"
	// "fmt"
	// "io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	// "todo-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	// "github.com/pashagolub/pgxmock/v3"
)

// 1. The Mock implementation
type MockDB struct {
	ShouldFail bool
	QueryFunc func(ctx context.Context, sql string, args ...any)(pgx.Rows, error)
	ExecFunc func(ctx context.Context, sql string, args ...any)(pgconn.CommandTag, error)
}

// 2. The Fake Database Result
type MockTodoRows struct {
	pgx.Rows
	count int
}

func (r *MockTodoRows) Next() bool {
	r.count++
	return r.count <= 2
}

func (r *MockTodoRows) Scan(dest ...any) error {
	switch r.count {
	case 1:
		*dest[0].(*int) = 1
		*dest[1].(*string) = "Finish the Matrix"
		*dest[2].(*bool) = false
	case 2:
		*dest[0].(*int) = 2
		*dest[1].(*string) = "Start Stage 16"
		*dest[2].(*bool) = true
	}
	return nil
}

func (r *MockTodoRows) Close(){}
func (r *MockTodoRows) Err() error {	return nil	}
func TestHandleTodos(t *testing.T) {
	tests := []struct {
		name string
		mockBehavior func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
		withAuth bool
		expectedStatus int
		expectedBody string
	}{
		{
			name: "Database Crash Returns 500",
			mockBehavior: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error){
				return nil, errors.New("Simutated database explosion")
			},
			withAuth: true,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Missing Auth Context Returns 500",
			mockBehavior: nil,
			withAuth: false,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Happy Path Returns 200 OK and Data",
			mockBehavior: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
				return &MockTodoRows{}, nil
			},
			withAuth: true,
			expectedStatus: http.StatusOK,
			expectedBody: "Finish the Matrix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockDB{
				QueryFunc: tt.mockBehavior,
			}

			h := &TodoHandler{
				Pool: mockDB,
				Logger: log.New(os.Stdout, "", 0),
			}

			req := httptest.NewRequest(http.MethodGet, "/todos", nil)

			if tt.withAuth {
				ctx := context.WithValue(req.Context(), "user_id", 1)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			h.HandleTodos(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Test '%s' failed: expected status %d, but got %d", tt.name, tt.expectedStatus, rr.Code)
			}

			if tt.expectedBody != "" && !strings.Contains(rr.Body.String(), tt.expectedBody) {
				t.Errorf("Test '%v' failed: expected to contain '%v', got '%s'", tt.name, tt.expectedBody, rr.Body.String())
			}
		})
	}
}

func (m *MockDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, sql, args...)
	}
	return nil, nil
}

// We just need these to satisfy the interface for now
func (m *MockDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row { return nil }
func (m *MockDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

// func TestHandleTodos_DatabaseError(t *testing.T) {
// 	mock := &MockDB{ShouldFail: true}
// 		// Create a real logger that writes to nowhere
// 	testLogger := log.New(io.Discard, "", 0)
	
// 	h := &TodoHandler{
// 		Pool:   mock, // We pass the mock instead of a real pool!
// 		Logger: testLogger,  // We don't need a real logger for this test
// 	}


// 	req := httptest.NewRequest(http.MethodGet, "/todos", nil)

// 	ctx := context.WithValue(req.Context(), "user_id", 1)
// 	req = req.WithContext(ctx)

// 	rr := httptest.NewRecorder()

// 	h.HandleTodos(rr, req)

// 	if rr.Code != http.StatusInternalServerError {
// 		t.Errorf("Expected status 500, but got %d", rr.Code)
// 	}
// }

// func TestHandleTodos_Success(t *testing.T) {
// 	mock, err := pgxmock.NewPool()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer mock.Close()

// 	h := &TodoHandler{
// 		Pool: mock,
// 		Logger: log.New(io.Discard, "", 0),
// 	}

// 	rows := pgxmock.NewRows([]string{"id", "title", "status"}).AddRow(1, "Finish Stage 7", false).AddRow(2, "Start Swagger", true)

// 	mock.ExpectQuery("SELECT id, title, status FROM todos WHERE user_id = \\$1").WithArgs(1).WillReturnRows(rows)

// 	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
// 	ctx := context.WithValue(req.Context(), "user_id", 1)
// 	req = req.WithContext(ctx)
// 	rr := httptest.NewRecorder()

// 	h.HandleTodos(rr, req)

// 	if rr.Code != http.StatusOK {
// 		t.Errorf("Expected 200, got %d", rr.Code)
// 	}

// 	var todos []models.Todo
// 	json.NewDecoder(rr.Body).Decode(&todos)
// 	if len(todos) != 2 {
// 		t.Errorf("Expected 2 todos, got %d", len(todos))
// 	}

// 	if err := mock.ExpectationsWereMet(); err != nil {
// 		t.Errorf("Unfulfilled expectations: %s", err)
// 	}
// }