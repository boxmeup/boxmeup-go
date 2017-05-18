package users

import "time"

// User is a user entity structure
type User struct {
	ID            int64     `json:"id"`
	Email         string    `json:"email"`
	Password      string    `json:"-"`
	UUID          string    `json:"uuid"`
	IsActive      bool      `json:"is_active"`
	ResetPassword bool      `json:"-"`
	Created       time.Time `json:"created"`
	Modified      time.Time `json:"modified"`
}
