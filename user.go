package main

import (
	"time"
)

// User is a user entity structure
type User struct {
	ID            int32  `json:"id"`
	Email         string `json:"email"`
	Password      string
	UUID          string `json:"uuid"`
	IsActive      bool   `json:"is_active"`
	ResetPassword bool
	Created       time.Time `json:"created"`
	Modified      time.Time `json:"modified"`
}
