package models

type StagedBook struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Author       *string `json:"author"`
	FileName     string  `json:"fileName"`
	Ext          string  `json:"ext"`
	OriginalPath string  `json:"originalPath"`
	UserID       string  `json:"userId"`
}
