package models

import "time"

// Document represents a document in the system
type Document struct {
	ID         int         `json:"id" db:"id"`
	UserID     int         `json:"userId" db:"user_id"`
	Title      string      `json:"title" db:"title"`
	Content    string      `json:"content,omitempty" db:"content"`
	Operations []Operation `json:"operations,omitempty"`
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at" db:"updated_at"`
}

// DocumentInput for creating/updating documents
type DocumentInput struct {
	Title        string       `json:"title"`
	UserID       int          `json:"userId"`
	AllowedUsers []Permission `json:"allowedUsers,omitempty"`
}
