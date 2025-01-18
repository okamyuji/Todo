package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gofrs/uuid"
	app "github.com/okamyuji/Todo/internal/app"
)

//go:embed internal/app/templates/* internal/app/static/js/*
var content embed.FS

func main() {
	state := &app.AppState{
		Todos: make([]app.Todo, 0),
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 静的ファイルの提供
	fileServer := http.FileServer(http.FS(content))
	r.Handle("/internal/app/static/*", fileServer)

	// HTMLテンプレート
	tmpl := template.Must(template.ParseFS(content, "internal/app/templates/*.html"))

	// ルート
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		state.Mu.RLock()
		data := map[string]interface{}{
			"Todos": state.Todos,
		}
		state.Mu.RUnlock()

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
			state.Mu.RLock()
			defer state.Mu.RUnlock()
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(state.Todos); err != nil {
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				return
			}
		})

		r.Post("/todos", func(w http.ResponseWriter, r *http.Request) {
			var todo app.Todo
			if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			todo.ID = uuid.Must(uuid.NewV4()).String()
			todo.CreatedAt = time.Now()

			state.Mu.Lock()
			state.Todos = append(state.Todos, todo)
			state.Mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(todo); err != nil {
				http.Error(w, "Failed to encode todo", http.StatusInternalServerError)
				return
			}
		})

		r.Put("/todos/{id}/toggle", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			state.Mu.Lock()
			defer state.Mu.Unlock()

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
					if err := json.NewEncoder(w).Encode(state.Todos[i]); err != nil {
						http.Error(w, "Failed to encode todo", http.StatusInternalServerError)
						return
					}
					return
				}
			}
			http.Error(w, "Todo not found", http.StatusNotFound)
		})

		r.Get("/analytics", func(w http.ResponseWriter, r *http.Request) {
			state.Mu.RLock()
			defer state.Mu.RUnlock()

			analytics := app.Analytics{
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
			if err := json.NewEncoder(w).Encode(analytics); err != nil {
				http.Error(w, "Failed to encode analytics", http.StatusInternalServerError)
				return
			}
		})
	})

	fmt.Println("Server starting at :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
