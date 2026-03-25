package models

import "time"

type ReadingProgress struct {
	ID             string     `json:"id"`
	UserID         string     `json:"userId"`
	BookID         string     `json:"bookId"`
	Status         string     `json:"status"`
	Progress       float64    `json:"progress"`
	LastReadAt     *time.Time `json:"lastReadAt"`
	DateStarted    *time.Time `json:"dateStarted"`
	DateFinished   *time.Time `json:"dateFinished"`
	PersonalRating *float64   `json:"personalRating"`
}
