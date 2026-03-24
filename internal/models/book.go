package models

type Book struct {
	ID          string  `json:"id"`
	LibraryID   string  `json:"libraryId"`
	UserID      string  `json:"userId"`
	Title       string  `json:"title"`
	Author      *string `json:"author"`
	Subject     *string `json:"subject"`
	Description *string `json:"description"`
	Publisher   *string `json:"publisher"`
	Contributor *string `json:"contributor"`
	Date        *string `json:"date"`
	Type        *string `json:"type"`
	Format      *string `json:"format"`
	Identifier  *string `json:"identifier"`
	Source      *string `json:"source"`
	Language    *string `json:"language"`
	Relation    *string `json:"relation"`
	Coverage    *string `json:"coverage"`
	Cover       *string `json:"cover"` // URL to cover endpoint, nil if no cover
	FilePath    string  `json:"filePath"`
}
