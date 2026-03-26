package models

import "time"

type Book struct {
	ID        string           `json:"id"`
	LibraryID string           `json:"libraryId"`
	UserID    string           `json:"userId"`
	FilePath  string           `json:"filePath"`
	AddedOn   time.Time        `json:"addedOn"`
	Metadata  BookMetadata     `json:"metadata"`
	Authors   []Author         `json:"authors"`
	Genres    []Genre          `json:"genres"`
	Tags      []Tag            `json:"tags"`
	Progress  *ReadingProgress `json:"progress,omitempty"`
}
