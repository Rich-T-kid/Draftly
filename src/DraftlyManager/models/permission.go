package models

// Permission defines user access to documents
type Permission struct {
	UserID     int    `json:"userId"`
	Permission string `json:"permission"`
}
