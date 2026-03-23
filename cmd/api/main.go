package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type sessionResponse struct {
	Session any `json:"session"`
	User    any `json:"user"`
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

		next.ServeHTTP(w, r)
	})
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(authMiddleware)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	http.ListenAndServe(":5321", r)
}
