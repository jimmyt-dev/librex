package handlers

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"

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

type OPDSSettings struct {
	Username string `json:"username"`
	Enabled  bool   `json:"enabled"`
}

type OPDSUpdate struct {
	Username *string `json:"username"`
	Password *string `json:"password"`
	Enabled  *bool   `json:"enabled"`
}

func GetOPDSSettings(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var s OPDSSettings
	err := db.DB.QueryRow(r.Context(),
		"SELECT username, enabled FROM opds_credentials WHERE user_id = $1",
		userID).Scan(&s.Username, &s.Enabled)
	if err != nil {
		// Not found is fine, return empty
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"username": "", "enabled": false})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func UpdateOPDSSettings(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var body OPDSUpdate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.Username != nil && *body.Username == "" {
		http.Error(w, "username cannot be empty", http.StatusBadRequest)
		return
	}

	// Get current settings
	var currentUsername string
	var currentEnabled bool
	err := db.DB.QueryRow(r.Context(),
		"SELECT username, enabled FROM opds_credentials WHERE user_id = $1",
		userID).Scan(&currentUsername, &currentEnabled)

	hasSettings := err == nil

	username := currentUsername
	if body.Username != nil {
		username = *body.Username
	}

	enabled := currentEnabled
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	if body.Password != nil && *body.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*body.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if hasSettings {
			_, err = db.DB.Exec(r.Context(),
				"UPDATE opds_credentials SET username = $1, password_hash = $2, enabled = $3 WHERE user_id = $4",
				username, string(hash), enabled, userID)
		} else {
			_, err = db.DB.Exec(r.Context(),
				"INSERT INTO opds_credentials (user_id, username, password_hash, enabled) VALUES ($1, $2, $3, $4)",
				userID, username, string(hash), enabled)
		}
	} else if hasSettings {
		_, err = db.DB.Exec(r.Context(),
			"UPDATE opds_credentials SET username = $1, enabled = $2 WHERE user_id = $3",
			username, enabled, userID)
	} else if body.Username != nil {
		// Can't create without password
		http.Error(w, "password is required to enable OPDS", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
