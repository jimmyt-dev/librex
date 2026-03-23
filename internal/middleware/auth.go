package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

type contextKey string

const userIDKey contextKey = "userID"

type sessionResponse struct {
	Session any `json:"session"`
	User    struct {
		ID string `json:"id"`
	} `json:"user"`
}

func GetUserID(r *http.Request) string {
	return r.Context().Value(userIDKey).(string)
}

func Auth(next http.Handler) http.Handler {
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
