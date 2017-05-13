package models

import "time"

// Location structure
type Location struct {
	ID         int
	User       User
	UUID       string
	Name       string
	IsMappable bool
	Address    string
	Created    time.Time
	Modified   time.Time
}
