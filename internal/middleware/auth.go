package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"librex/internal/db"
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
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:5173"
		}

		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, frontendURL+"/api/auth/get-session", nil)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Forward bearer token if present
		if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
			req.Header.Set("Authorization", authHeader)
		}

		// Forward session cookie so Better Auth can validate via cookie too
		if cookie := r.Header.Get("Cookie"); cookie != "" {
			req.Header.Set("Cookie", cookie)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			goto fallback
		}
		defer resp.Body.Close()

		{
			var session sessionResponse
			if err := json.NewDecoder(resp.Body).Decode(&session); err == nil && session.Session != nil && session.User.ID != "" {
				ctx := context.WithValue(r.Context(), userIDKey, session.User.ID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

	fallback:
		// Fallback to Basic Auth (for OPDS clients)
		if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Basic ") {
			payload, err := base64.StdEncoding.DecodeString(authHeader[6:])
			if err == nil {
				parts := strings.SplitN(string(payload), ":", 2)
				if len(parts) == 2 {
					username, password := parts[0], parts[1]

					var userID, passwordHash string
					var enabled bool
					err = db.DB.QueryRow(r.Context(),
						"SELECT user_id, password_hash, enabled FROM opds_credentials WHERE username = $1",
						username).Scan(&userID, &passwordHash, &enabled)

					if err == nil && enabled {
						if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err == nil {
							ctx := context.WithValue(r.Context(), userIDKey, userID)
							next.ServeHTTP(w, r.WithContext(ctx))
							return
						}
					}
				}
			}
		}

		http.Error(w, "invalid token", http.StatusUnauthorized)
	})
}
