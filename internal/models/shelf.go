package models

type Shelf struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Icon   *string `json:"icon"`
	UserID string  `json:"userId"`
	Books  int     `json:"books"`
}
