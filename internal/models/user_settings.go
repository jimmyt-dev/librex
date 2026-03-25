package models

type UserSettings struct {
	ID                  string `json:"id"`
	UserID              string `json:"userId"`
	FileNamingPattern   string `json:"fileNamingPattern"`
	WriteMetadataToFile bool   `json:"writeMetadataToFile"`
}
