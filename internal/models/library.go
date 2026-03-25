package models

type Library struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Icon              *string `json:"icon"`
	Folder            *string `json:"folder"`
	FileNamingPattern *string `json:"fileNamingPattern"`
	UserID            string  `json:"userId"`
	BookCount         int     `json:"bookCount"`
}
