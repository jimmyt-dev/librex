package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type contextKey string

const userIDKey contextKey = "userID"

var db *sql.DB

type sessionResponse struct {
	Session any `json:"session"`
	User    struct {
		ID string `json:"id"`
	} `json:"user"`
}

type Library struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	UserID string `json:"userId"`
}

func initDB() error {
	var err error
	db, err = sql.Open("sqlite", "reliquary.db")
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS libraries (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		user_id TEXT NOT NULL
	)`)
	return err
}

func getUserID(r *http.Request) string {
	return r.Context().Value(userIDKey).(string)
}

func getLibraries(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	rows, err := db.QueryContext(r.Context(), "SELECT id, name, user_id FROM libraries WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	libraries := []Library{}
	for rows.Next() {
		var l Library
		if err := rows.Scan(&l.ID, &l.Name, &l.UserID); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		libraries = append(libraries, l)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(libraries)
}

func createLibrary(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	_, err := db.ExecContext(r.Context(), "INSERT INTO libraries (id, name, user_id) VALUES (?, ?, ?)", id, body.Name, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Library{ID: id, Name: body.Name, UserID: userID})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:5173"
		}

		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, frontendURL+"/api/auth/get-session", nil)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Authorization", authHeader)

		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		defer resp.Body.Close()

		var session sessionResponse
		if err := json.NewDecoder(resp.Body).Decode(&session); err != nil || session.Session == nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		if session.User.ID == "" {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, session.User.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	if err := initDB(); err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(authMiddleware)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	r.Get("/api/libraries", getLibraries)
	r.Post("/api/libraries", createLibrary)

	http.ListenAndServe(":5321", r)
}
