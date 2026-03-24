package models

type Library struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Icon      *string `json:"icon"`
	Folder    *string `json:"folder"`
	UserID    string  `json:"userId"`
	BookCount int     `json:"bookCount"`
}
