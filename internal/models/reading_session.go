package models

import "time"

type ReadingSession struct {
	ID              string     `json:"id"`
	UserID          string     `json:"userId"`
	BookID          string     `json:"bookId"`
	StartTime       time.Time  `json:"startTime"`
	EndTime         *time.Time `json:"endTime"`
	DurationSeconds *int       `json:"durationSeconds"`
	StartProgress   *float64   `json:"startProgress"`
	EndProgress     *float64   `json:"endProgress"`
}
