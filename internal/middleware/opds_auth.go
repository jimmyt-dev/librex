package middleware

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"librex/internal/db"
)

func OPDSAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Basic ") {
			w.Header().Set("WWW-Authenticate", `Basic realm="Librex OPDS"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		payload, err := base64.StdEncoding.DecodeString(authHeader[6:])
		if err != nil {
			http.Error(w, "invalid auth header", http.StatusBadRequest)
			return
		}

		parts := strings.SplitN(string(payload), ":", 2)
		if len(parts) != 2 {
			http.Error(w, "invalid auth header", http.StatusBadRequest)
			return
		}

		username := parts[0]
		password := parts[1]

		var userID string
		var passwordHash string
		var enabled bool
		err = db.DB.QueryRow(r.Context(),
			"SELECT user_id, password_hash, enabled FROM opds_credentials WHERE username = $1",
			username).Scan(&userID, &passwordHash, &enabled)

		if err != nil || !enabled {
			w.Header().Set("WWW-Authenticate", `Basic realm="Librex OPDS"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
			w.Header().Set("WWW-Authenticate", `Basic realm="Librex OPDS"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
