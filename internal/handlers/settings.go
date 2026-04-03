package handlers

import (
	"encoding/json"
	"net/http"

	"librex/internal/db"
	"librex/internal/middleware"
	"librex/internal/models"
)

const defaultFileNamingPattern = "{authors}/{title}{ext}"

// GetSettings returns the user's settings, creating a default row if none exists.
func GetSettings(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var s models.UserSettings
	err := db.DB.QueryRow(r.Context(),
		`SELECT id, user_id, file_naming_pattern, write_metadata_to_file, max_upload_size_mb FROM user_settings WHERE user_id = $1`,
		userID).Scan(&s.ID, &s.UserID, &s.FileNamingPattern, &s.WriteMetadataToFile, &s.MaxUploadSizeMb)
	if err != nil {
		// Auto-create default settings
		err = db.DB.QueryRow(r.Context(),
			`INSERT INTO user_settings (user_id, file_naming_pattern, write_metadata_to_file, max_upload_size_mb)
			VALUES ($1, $2, false, 100)
			RETURNING id, user_id, file_naming_pattern, write_metadata_to_file, max_upload_size_mb`,
			userID, defaultFileNamingPattern).Scan(&s.ID, &s.UserID, &s.FileNamingPattern, &s.WriteMetadataToFile, &s.MaxUploadSizeMb)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

type settingsUpdate struct {
	FileNamingPattern   *string `json:"fileNamingPattern"`
	WriteMetadataToFile *bool   `json:"writeMetadataToFile"`
	MaxUploadSizeMb     *int    `json:"maxUploadSizeMb"`
}

// UpdateSettings updates the user's settings.
func UpdateSettings(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var body settingsUpdate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure settings row exists
	var s models.UserSettings
	err := db.DB.QueryRow(r.Context(),
		`INSERT INTO user_settings (user_id, file_naming_pattern, write_metadata_to_file, max_upload_size_mb)
		VALUES ($1, $2, false, 100)
		ON CONFLICT (user_id) DO UPDATE SET user_id = EXCLUDED.user_id
		RETURNING id, user_id, file_naming_pattern, write_metadata_to_file, max_upload_size_mb`,
		userID, defaultFileNamingPattern).Scan(&s.ID, &s.UserID, &s.FileNamingPattern, &s.WriteMetadataToFile, &s.MaxUploadSizeMb)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	if body.FileNamingPattern != nil {
		s.FileNamingPattern = *body.FileNamingPattern
	}
	if body.WriteMetadataToFile != nil {
		s.WriteMetadataToFile = *body.WriteMetadataToFile
	}
	if body.MaxUploadSizeMb != nil {
		s.MaxUploadSizeMb = *body.MaxUploadSizeMb
	}

	_, err = db.DB.Exec(r.Context(),
		`UPDATE user_settings SET file_naming_pattern = $1, write_metadata_to_file = $2, max_upload_size_mb = $3 WHERE id = $4`,
		s.FileNamingPattern, s.WriteMetadataToFile, s.MaxUploadSizeMb, s.ID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}
