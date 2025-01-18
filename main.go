package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gofrs/uuid"
)

//go:embed templates/* static/*
var content embed.FS

type Todo struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Category  string     `json:"category"`
	Priority  int        `json:"priority"`
	Done      bool       `json:"done"`
	CreatedAt time.Time  `json:"created_at"`
	DoneAt    *time.Time `json:"done_at,omitempty"`
}

type Analytics struct {
	TotalTodos     int            `json:"total_todos"`
	CompletedTodos int            `json:"completed_todos"`
	CompletionRate float64        `json:"completion_rate"`
	AverageTime    float64        `json:"average_time"`
	CategoryCounts map[string]int `json:"category_counts"`
	PriorityCounts map[int]int    `json:"priority_counts"`
}

type AppState struct {
	Todos []Todo
	mu    sync.RWMutex
}

func main() {
	state := &AppState{
		Todos: make([]Todo, 0),
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 静的ファイルの提供
	fileServer := http.FileServer(http.FS(content))
	r.Handle("/static/*", fileServer)

	// HTMLテンプレート
	tmpl := template.Must(template.ParseFS(content, "templates/*.html"))

	// ルート
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		state.mu.RLock()
		data := map[string]interface{}{
			"Todos": state.Todos,
		}
		state.mu.RUnlock()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	// API エンドポイント
	r.Route("/api", func(r chi.Router) {
		r.Get("/todos", func(w http.ResponseWriter, r *http.Request) {
			state.mu.RLock()
			defer state.mu.RUnlock()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(state.Todos)
		})

		r.Post("/todos", func(w http.ResponseWriter, r *http.Request) {
			var todo Todo
			if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			todo.ID = uuid.Must(uuid.NewV4()).String()
			todo.CreatedAt = time.Now()

			state.mu.Lock()
			state.Todos = append(state.Todos, todo)
			state.mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(todo)
		})

		r.Put("/todos/{id}/toggle", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			state.mu.Lock()
			defer state.mu.Unlock()

			for i := range state.Todos {
				if state.Todos[i].ID == id {
					state.Todos[i].Done = !state.Todos[i].Done
					if state.Todos[i].Done {
						now := time.Now()
						state.Todos[i].DoneAt = &now
					} else {
						state.Todos[i].DoneAt = nil
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(state.Todos[i])
					return
				}
			}
			http.Error(w, "Todo not found", http.StatusNotFound)
		})

		r.Get("/analytics", func(w http.ResponseWriter, r *http.Request) {
			state.mu.RLock()
			defer state.mu.RUnlock()

			analytics := Analytics{
				TotalTodos:     len(state.Todos),
				CategoryCounts: make(map[string]int),
				PriorityCounts: make(map[int]int),
			}

			var totalCompletionTime float64
			var completedCount int

			for _, todo := range state.Todos {
				if todo.Done {
					analytics.CompletedTodos++
					if todo.DoneAt != nil {
						completionTime := todo.DoneAt.Sub(todo.CreatedAt).Hours()
						totalCompletionTime += completionTime
						completedCount++
					}
				}
				analytics.CategoryCounts[todo.Category]++
				analytics.PriorityCounts[todo.Priority]++
			}

			if analytics.TotalTodos > 0 {
				analytics.CompletionRate = float64(analytics.CompletedTodos) / float64(analytics.TotalTodos) * 100
			}
			if completedCount > 0 {
				analytics.AverageTime = totalCompletionTime / float64(completedCount)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(analytics)
		})
	})

	fmt.Println("Server starting at :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
