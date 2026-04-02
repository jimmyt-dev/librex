package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"reliquary/internal/db"
	"reliquary/internal/middleware"
	"reliquary/internal/models"
)

// GetReadingProgress returns the reading progress for a book.
func GetReadingProgress(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	bookID := chi.URLParam(r, "id")

	var p models.ReadingProgress
	err := db.DB.QueryRow(r.Context(),
		`SELECT id, user_id, book_id, status, progress, last_read_at, date_started, date_finished, personal_rating
		FROM reading_progress WHERE book_id = $1 AND user_id = $2`,
		bookID, userID).Scan(&p.ID, &p.UserID, &p.BookID, &p.Status, &p.Progress,
		&p.LastReadAt, &p.DateStarted, &p.DateFinished, &p.PersonalRating)
	if err != nil {
		http.Error(w, "progress not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

type progressUpdate struct {
	Status         *string  `json:"status"`
	Progress       *float64 `json:"progress"`
	PersonalRating *float64 `json:"personalRating"`
	DateStarted    *string  `json:"dateStarted"`
	DateFinished   *string  `json:"dateFinished"`
}

// UpdateReadingProgress creates or updates reading progress for a book.
func UpdateReadingProgress(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	bookID := chi.URLParam(r, "id")

	var body progressUpdate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validation
	if body.Progress != nil && (*body.Progress < 0 || *body.Progress > 100) {
		http.Error(w, "progress must be between 0 and 100", http.StatusBadRequest)
		return
	}
	if body.PersonalRating != nil && (*body.PersonalRating < 0 || *body.PersonalRating > 5) {
		http.Error(w, "rating must be between 0 and 5", http.StatusBadRequest)
		return
	}

	// Verify book belongs to user
	var exists bool
	if err := db.DB.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM books WHERE id = $1 AND user_id = $2)",
		bookID, userID).Scan(&exists); err != nil || !exists {
		http.Error(w, "book not found", http.StatusNotFound)
		return
	}

	now := time.Now()

	// Use manually provided dates if given, otherwise auto-calculate from status.
	var dateStarted, dateFinished *time.Time
	if body.DateStarted != nil {
		if t, err := time.Parse(time.RFC3339, *body.DateStarted); err == nil {
			dateStarted = &t
		} else if t, err := time.Parse("2006-01-02", *body.DateStarted); err == nil {
			dateStarted = &t
		}
	}
	if body.DateFinished != nil {
		if t, err := time.Parse(time.RFC3339, *body.DateFinished); err == nil {
			dateFinished = &t
		} else if t, err := time.Parse("2006-01-02", *body.DateFinished); err == nil {
			dateFinished = &t
		}
	}
	// Auto-set dates from status transitions only when not manually provided.
	if body.Status != nil {
		if dateStarted == nil && (*body.Status == "reading" || *body.Status == "finished") {
			dateStarted = &now
		}
		if dateFinished == nil && *body.Status == "finished" {
			dateFinished = &now
		}
	}

	// When dates are explicitly provided by the user, always use them (even overwriting existing).
	// When auto-calculated, only set if no existing value.
	manualStarted := body.DateStarted != nil
	manualFinished := body.DateFinished != nil

	var p models.ReadingProgress
	err := db.DB.QueryRow(r.Context(),
		`INSERT INTO reading_progress (user_id, book_id, status, progress, last_read_at, date_started, date_finished, personal_rating)
		VALUES ($1, $2, COALESCE($3, 'unread'), COALESCE($4, 0), $5, $6, $7, $8)
		ON CONFLICT (user_id, book_id) DO UPDATE SET
			status          = COALESCE($3, reading_progress.status),
			progress        = COALESCE($4, reading_progress.progress),
			last_read_at    = $5,
			date_started    = CASE
				WHEN $9 THEN $6
				WHEN reading_progress.date_started IS NOT NULL THEN reading_progress.date_started
				ELSE $6
			END,
			date_finished   = CASE
				WHEN $10 THEN $7
				WHEN reading_progress.date_finished IS NOT NULL THEN reading_progress.date_finished
				ELSE $7
			END,
			personal_rating = COALESCE($8, reading_progress.personal_rating)
		RETURNING id, user_id, book_id, status, progress, last_read_at, date_started, date_finished, personal_rating`,
		userID, bookID, body.Status, body.Progress, &now, dateStarted, dateFinished, body.PersonalRating,
		manualStarted, manualFinished).
		Scan(&p.ID, &p.UserID, &p.BookID, &p.Status, &p.Progress,
			&p.LastReadAt, &p.DateStarted, &p.DateFinished, &p.PersonalRating)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// DeleteReadingProgress removes reading progress for a book.
func DeleteReadingProgress(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	bookID := chi.URLParam(r, "id")

	_, err := db.DB.Exec(r.Context(),
		`DELETE FROM reading_progress WHERE book_id = $1 AND user_id = $2`,
		bookID, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListReadingSessions returns all reading sessions for a book.
func ListReadingSessions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	bookID := chi.URLParam(r, "id")

	rows, err := db.DB.Query(r.Context(),
		`SELECT id, user_id, book_id, start_time, end_time, duration_seconds, start_progress, end_progress
		FROM reading_sessions
		WHERE book_id = $1 AND user_id = $2
		ORDER BY start_time DESC`, bookID, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	sessions := []models.ReadingSession{}
	for rows.Next() {
		var s models.ReadingSession
		if err := rows.Scan(&s.ID, &s.UserID, &s.BookID, &s.StartTime, &s.EndTime,
			&s.DurationSeconds, &s.StartProgress, &s.EndProgress); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		sessions = append(sessions, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

type sessionCreate struct {
	StartTime       time.Time  `json:"startTime"`
	EndTime         *time.Time `json:"endTime"`
	DurationSeconds *int       `json:"durationSeconds"`
	StartProgress   *float64   `json:"startProgress"`
	EndProgress     *float64   `json:"endProgress"`
}

// CreateReadingSession logs a new reading session for a book.
func CreateReadingSession(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	bookID := chi.URLParam(r, "id")

	var body sessionCreate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.StartTime.IsZero() {
		http.Error(w, "startTime is required", http.StatusBadRequest)
		return
	}

	if body.EndTime != nil && body.EndTime.Before(body.StartTime) {
		http.Error(w, "endTime cannot be before startTime", http.StatusBadRequest)
		return
	}

	// Auto-calculate duration if end time provided but no duration
	if body.EndTime != nil && body.DurationSeconds == nil {
		dur := int(body.EndTime.Sub(body.StartTime).Seconds())
		body.DurationSeconds = &dur
	}

	var s models.ReadingSession
	err := db.DB.QueryRow(r.Context(),
		`INSERT INTO reading_sessions (user_id, book_id, start_time, end_time, duration_seconds, start_progress, end_progress)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, book_id, start_time, end_time, duration_seconds, start_progress, end_progress`,
		userID, bookID, body.StartTime, body.EndTime, body.DurationSeconds,
		body.StartProgress, body.EndProgress).
		Scan(&s.ID, &s.UserID, &s.BookID, &s.StartTime, &s.EndTime,
			&s.DurationSeconds, &s.StartProgress, &s.EndProgress)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}
