package models

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

type FilteredUser struct {
	ID    int32  `json:"id"`
	Email string `json:"email"`
	UUID  string `json:"uuid"`
}

func (user User) ToSafeUser() FilteredUser {
	newUser := FilteredUser{
		ID:    user.ID,
		Email: user.Email,
		UUID:  user.UUID,
	}
	return newUser
}
