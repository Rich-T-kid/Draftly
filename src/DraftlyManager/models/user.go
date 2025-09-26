package models

import "time"

// User represents a user entity
type User struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserInput for API requests
type UserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
