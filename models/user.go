package models

import (
	"database/sql"
	"log"
	"time"
)

// User is a user entity structure
type User struct {
	ID            int       `json:"id"`
	Email         string    `json:"email"`
	Password      string    `json:"-"`
	UUID          string    `json:"uuid"`
	IsActive      bool      `json:"is_active"`
	ResetPassword bool      `json:"-"`
	Created       time.Time `json:"created"`
	Modified      time.Time `json:"modified"`
}

// GetUserByID resolves with a user on the channel.
func GetUserByID(db *sql.DB, ID int) (User, error) {
	user := User{}
	q := `
		select id, email, password, uuid, is_active, reset_password, created, modified
		from users where id = ?
	`
	err := db.QueryRow(q, ID).Scan(&user.ID, &user.Email, &user.Password, &user.UUID, &user.IsActive, &user.ResetPassword, &user.Created, &user.Modified)
	if err != nil {
		if err == sql.ErrNoRows {
			// user not found
			// @todo consider sending a custom error that the route handler can consume
		} else {
			log.Fatal(err)
		}
	}
	return user, err
}
