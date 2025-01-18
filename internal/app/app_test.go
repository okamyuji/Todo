package app_test

import (
	"bytes"
	"embed"
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	app "github.com/okamyuji/Todo/internal/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed templates/*.html
var templatesFS embed.FS

func setupTestApp() (*app.AppState, *chi.Mux) {
	state := &app.AppState{
		Todos: make([]app.Todo, 0),
		Mu:    sync.RWMutex{},
	}

	// テンプレートの初期化
	tmpl := template.Must(template.ParseFS(templatesFS, "templates/*.html"))

	r := chi.NewRouter()
	app.RegisterRoutes(r, state, tmpl)
	return state, r
}

func TestGetTodos(t *testing.T) {
	state, r := setupTestApp()

	// Pre-populate Todos
	state.Mu.Lock()
	state.Todos = []app.Todo{
		{ID: "1", Title: "Test Todo", Category: "Test", Priority: 1, Done: false, CreatedAt: time.Now()},
	}
	state.Mu.Unlock()

	req, _ := http.NewRequest(http.MethodGet, "/api/todos", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var todos []app.Todo
	err := json.NewDecoder(resp.Body).Decode(&todos)
	require.NoError(t, err)
	assert.Len(t, todos, 1)
	assert.Equal(t, "Test Todo", todos[0].Title)
}

func TestCreateTodo(t *testing.T) {
	_, r := setupTestApp()

	todo := app.Todo{
		Title:    "New Todo",
		Category: "Test",
		Priority: 2,
	}
	body, _ := json.Marshal(todo)
	req, _ := http.NewRequest(http.MethodPost, "/api/todos", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)

	var createdTodo app.Todo
	err := json.NewDecoder(resp.Body).Decode(&createdTodo)
	require.NoError(t, err)
	assert.Equal(t, "New Todo", createdTodo.Title)
	assert.NotEmpty(t, createdTodo.ID)
}

func TestToggleTodoStatus(t *testing.T) {
	state, r := setupTestApp()

	// Pre-populate Todos
	state.Mu.Lock()
	state.Todos = []app.Todo{
		{ID: "1", Title: "Test Todo", Done: false, CreatedAt: time.Now()},
	}
	state.Mu.Unlock()

	req, _ := http.NewRequest(http.MethodPut, "/api/todos/1/toggle", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var updatedTodo app.Todo
	err := json.NewDecoder(resp.Body).Decode(&updatedTodo)
	require.NoError(t, err)
	assert.True(t, updatedTodo.Done)
}

func TestAnalytics(t *testing.T) {
	state, r := setupTestApp()

	// Pre-populate Todos
	state.Mu.Lock()
	now := time.Now()
	state.Todos = []app.Todo{
		{ID: "1", Title: "Todo 1", Done: true, CreatedAt: now.Add(-2 * time.Hour), DoneAt: &now},
		{ID: "2", Title: "Todo 2", Done: false, CreatedAt: now.Add(-1 * time.Hour)},
	}
	state.Mu.Unlock()

	req, _ := http.NewRequest(http.MethodGet, "/api/analytics", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var analytics app.Analytics
	err := json.NewDecoder(resp.Body).Decode(&analytics)
	require.NoError(t, err)
	assert.Equal(t, 2, analytics.TotalTodos)
	assert.Equal(t, 1, analytics.CompletedTodos)
	assert.InDelta(t, 50.0, analytics.CompletionRate, 0.01)
}
