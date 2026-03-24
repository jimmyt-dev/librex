package models

type StagedBook struct {
	ID          string  `json:"id"`
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
	HasCover    bool    `json:"hasCover"`
	FileName    string  `json:"fileName"`
	Ext         string  `json:"ext"`
	OriginalPath string `json:"originalPath"`
	UserID      string  `json:"userId"`
}
